package tritonhttp


import (
	"os"
	"log"
	"bufio"
	"bytes"
	"strings"
)

func ParseMIME(MIMEPath string) (map[string]string, error) {
	file, err := os.Open(MIMEPath)
	if err != nil {
		log.Panicln(err)
	}

	defer func() {
		if err = file.Close(); err != nil {
			log.Panicln(err)
		}
	}()

	MIMEMap := map[string]string{}

	s := bufio.NewScanner(file)
	for s.Scan() {
		line := strings.Fields(s.Text()) // [".pdf", "application/pdf"]
		MIMEMap[line[0]] = line[1]
	}
	err = s.Err()
	if err != nil {
		log.Panicln(err)
	}

	return MIMEMap, nil
}

// maps the status code to it's description
var StatusDesc = map[int]string{
	200: "OK",
	400: "Bad Request",
	404: "Not Found",
}

// number of bytes in one kilobyte
const KB = 1024

const CRLF = "\r\n"
const EoR = CRLF+CRLF

type SimpleBuffer struct {
	buffer	[]byte	// the data
	size	int		// no. of valid elements in buffer
}

func (sb *SimpleBuffer) Write(data []byte) int {
	space := len(sb.buffer) - sb.size
	toWrite := len(data)
	if space < toWrite{
		toWrite = space
	}
	written := copy(sb.buffer[sb.size:], data[:toWrite])
	sb.size += written
	return written
}

// reads the buffer until index and returns a string
// shift-left  all the elements in the array and adjust size
func (sb *SimpleBuffer) Read(toRead int) string {
	if toRead > sb.size{
		toRead = sb.size
	}
	data := string(sb.buffer[:toRead])

	copy(sb.buffer, sb.buffer[toRead:sb.size])
	sb.size -= len(data)
	return data
}

func (sb *SimpleBuffer) IndexOf(data []byte) int {
	return bytes.Index(sb.buffer[0:sb.size], data)
}

func (sb *SimpleBuffer) IsFull() bool {
	return sb.size == len(sb.buffer)
}

func (sb *SimpleBuffer) IsEmpty() bool {
	return sb.size == 0
}