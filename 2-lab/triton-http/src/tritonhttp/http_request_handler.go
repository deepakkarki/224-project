package tritonhttp

import (
	"io"
	"log"
	"net"
	"time"
	"strings"
	"regexp"
	"errors"
)

/*
For a connection, keep handling requests until
	1. a timeout occurs or
	2. client closes connection or
	3. client sends a bad request (wrong format or size >8KB)
*/
func (hs *HttpServer) handleConnection(conn net.Conn) {
	log.Println("Accepted new connection from: ", conn.RemoteAddr())
	defer conn.Close()
	defer log.Println("Closed connection.\n\n")

	// create an 8 KB buffer
	sb := SimpleBuffer{ buffer: make([]byte, 32*KB), size: 0}

	// keep looping until we -
	//	1. getNextReq() returns err: timeout, disconnected or Req > 8KB
	//	2. A bad req. - mostly due to parsing err below
	//	3. Request header has "Connection : Close"
	for {
		reqData, err := getNextReq(conn, &sb)
		if err != nil {
			if err != io.EOF && !sb.IsEmpty(){ // timeout or req >8KB
				log.Println("Bad request: timeout or req > 8KB")
				hs.handleBadRequest(conn)
			}
			// else client disconnected, don't do anything
			return
		}

		// create the HttpRequestHeader
		reqHeader, err := makeReqHeader(reqData)
		if err != nil {
			log.Println("Bad request.", err)
			hs.handleBadRequest(conn)
			break
		} else {
			log.Println("Request Parsed:\n", reqHeader)
			hs.handleResponse(&reqHeader, conn)
		}

		if reqHeader.headers["Connection"] == "close" {
			break
		}
	}

}


// this waits on and fetches incoming req.
func getNextReq(conn net.Conn, sb *SimpleBuffer) (string, error){
	// local buf to read from socket
	tmpBuf := make([]byte, 32*KB) // to handle large requests

	// keep looping until we -
	// find a valid req, or buffer is full, or error reading conn (timeout)
	for {
		// does request exist in the buffer?
		if i := sb.IndexOf([]byte(EoR)); i >= 0 {
			data := sb.Read(i+4)
			return data[:i], nil
		}

		// The buffer is full & still no valid req => a bad request
		if sb.IsFull() {
			return "", errors.New("Request too long")
		}

		// if not try to read for more
		timeoutDuration := 5 * time.Second
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		numBytes, err := conn.Read(tmpBuf)
		if err != nil {
			log.Println("Error reading conn data:", err)
			return "", err
		}
		sb.Write(tmpBuf[:numBytes])
	}
}

func makeReqHeader(reqData string) (HttpRequestHeader, error) {
	data := strings.Split(reqData, "\r\n")
	reqLine := strings.Fields(data[0])
	if !(len(reqLine) == 3 &&
		 reqLine[0] == "GET" &&
		 strings.HasPrefix(reqLine[1], "/") &&
		 reqLine[2] == "HTTP/1.1" ){
			reqHeader := HttpRequestHeader{}
			return reqHeader, errors.New("Unexpected request line: " + data[0])
	}

	headers := map[string]string{}
	reg, _ := regexp.Compile(`^([\w-]+): (.*)$`)

	for _, header := range data[1:] {
		kvList := reg.FindStringSubmatch(header)
		if len(kvList) != 3 {
			reqHeader := HttpRequestHeader{}
			return reqHeader, errors.New("Unexpected header format: " + header)
		}
		headers[kvList[1]] = kvList[2]
	}

	// check for Host header
	if _, ok := headers["Host"]; !ok {
		reqHeader := HttpRequestHeader{}
		return reqHeader, errors.New("Missing header 'Host'")
	}

	reqHeader := HttpRequestHeader{
		verb: reqLine[0],
		url: strings.Split(reqLine[1], "?")[0],
		headers: headers,
	}

	return reqHeader, nil
}
