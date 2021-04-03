package tritonhttp

import (
	"strings"
)

type HttpServer	struct {
	ServerPort	string
	DocRoot		string
	MIMEPath	string
	MIMEMap		map[string]string
}

type HttpResponseHeader struct {
	Status		string // e.g. "200 OK"
	StatusCode	int    // e.g. 200
	Proto		string // e.g. "HTTP/1.0"

	Headers map[string]string

}

type HttpRequestHeader struct {
	verb	string
	url		string
	headers map[string]string
}

func (rh HttpRequestHeader) String() string{
	var b strings.Builder
	b.WriteString(rh.verb + " " + rh.url + " HTTP/1.1\n")
	for k, v := range rh.headers {
		b.WriteString(k + ": " + v + "\n")
	}
	return b.String()
}

func (rh HttpResponseHeader) String() string{
	var b strings.Builder
	b.WriteString(rh.Proto + " " + rh.Status + "\n")
	for k, v := range rh.Headers {
		b.WriteString(k + ": " + v + "\n")
	}
	return b.String()
}