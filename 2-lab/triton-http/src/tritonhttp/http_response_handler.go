package tritonhttp

import (
	"os"
	"io"
	"log"
	"net"
	"time"
	"strconv"
	"strings"
	"path/filepath"
)

func fileExists(path string) bool {
    fileInfo, err := os.Stat(path)
    return !os.IsNotExist(err) && !fileInfo.IsDir()
}

func isDir(path string) bool {
    fileInfo, err := os.Stat(path)
    return !os.IsNotExist(err) && fileInfo.IsDir()
}

func fileLastModified(path string) string {
	fileInfo, err := os.Stat(path)
	if err != nil { // should result in server error
		log.Println(err)
		return "" // ok?
	}
	return fileInfo.ModTime().Format(time.RFC1123)
}

func fileSize(path string) int64 {
	fileInfo, err := os.Stat(path)
	if err != nil { // should be server error
		log.Println(err)
		return 0 // ok?
	}
	return fileInfo.Size()
}

func headerToString(resHeader HttpResponseHeader) string{
	var b strings.Builder

	// HTTP/1.1 200 OK\r\n
	b.WriteString(resHeader.Proto + " " + resHeader.Status + CRLF)

	for key, value := range resHeader.Headers {
		b.WriteString(key + ": " + value + CRLF)
	}
	b.WriteString(CRLF)
	return b.String()
}

func (hs *HttpServer) handleBadRequest(conn net.Conn) {
	body := "Bad Request"

	headers := map[string]string{ "Server": "TritonHTTP", "Connection": "close",
				"Content-Type": "text/plain", "Content-Length": strconv.Itoa(len(body))}
	resHeader := HttpResponseHeader{Proto: "HTTP/1.1", Headers: headers,
										Status: "400 Bad Request", StatusCode:400}


	headerString := headerToString(resHeader)
	log.Println("Sending response:\n", headerString)
	conn.Write([]byte(headerString+body))
}

func (hs *HttpServer) handleResponse(requestHeader *HttpRequestHeader, conn net.Conn) {

	// check if file exists under the server-root dir
	file, _ := filepath.Abs(hs.DocRoot + requestHeader.url)
	if isDir(file){
		file += "/index.html"
	}

	headers := map[string]string{ "Server" : "TritonHTTP"}
	if requestHeader.headers["Connection"] == "close"{
		headers["Connection"] = "close"
	}
	resHeader := HttpResponseHeader{Proto: "HTTP/1.1", Headers: headers}

	if (strings.HasPrefix(file, hs.DocRoot) && fileExists(file)) {

		// ext = filepath.Ext(file)
		ext := filepath.Ext(file)
		contentType, exists := hs.MIMEMap[ext]
		if !exists {
			contentType = "application/octet-stream"
		}

		resHeader.Status = "200 OK"
		resHeader.StatusCode = 200
		headers["Content-Type"] = contentType
		headers["Content-Length"] = strconv.FormatInt(fileSize(file), 10)
		headers["Last-Modified"] = fileLastModified(file)
		hs.sendResponse(resHeader, conn, file)
	} else{
		resHeader.Status = "404 Not Found"
		resHeader.StatusCode = 404
		headers["Content-Type"] = "text/plain"
		hs.handleFileNotFoundRequest(resHeader, conn)
	}
}

func (hs *HttpServer) handleFileNotFoundRequest(responseHeader HttpResponseHeader, conn net.Conn) {
	body := "404"
	responseHeader.Headers["Content-Length"] = strconv.Itoa(len(body))
	headerString := headerToString(responseHeader)
	log.Println("Sending response:\n", headerString)
	conn.Write([]byte(headerString))
	conn.Write([]byte(body))
}

func (hs *HttpServer) sendResponse(responseHeader HttpResponseHeader, conn net.Conn, path string) {
	// Send headers
	headerString := headerToString(responseHeader)
	log.Println("Sending response:\n", headerString)
	conn.Write([]byte(headerString))

	//Create 8KB buffer (or should I decide dynamically based on file size)
	buffer := make([]byte, 8*KB)

	f, err := os.Open(path)
	if err!= nil {
		return // what would one really do here?
	}
	defer f.Close()

	// can use bufio WriteTo instead
	for {
		num, err := f.Read(buffer)
		if err == io.EOF {
			break
		}
		conn.Write(buffer[:num])
	}
}
