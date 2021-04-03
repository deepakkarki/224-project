package tritonhttp

import (
	"log"
	"net"
)

/**
	Initialize the tritonhttp server by populating HttpServer structure
**/
func NewHttpdServer(port, docRoot, mimePath string) (*HttpServer, error) {

	// Initialize mimeMap for server to refer
	// ParseMIME()
	mimeMap, _ := ParseMIME(mimePath)

	// Return pointer to HttpServer
	server := &HttpServer{
		ServerPort: port,
		DocRoot: docRoot,
		MIMEPath: mimePath,
		MIMEMap: mimeMap,
	}

	return server, nil
}

/**
	Start the tritonhttp server
**/
func (hs *HttpServer) Start() (err error) {

	log.Println("Server Created")

	// Start listening to the server port
	sock, err := net.Listen("tcp4", ":"+hs.ServerPort)
	if err != nil {
		log.Panicln(err)
	}
	log.Println("Listening to connections on", sock.Addr())
	defer sock.Close()

	for {
		// Accept connection from client
		conn, err := sock.Accept()
		if err != nil {
			log.Panicln(err)
		}

		// Spawn a go routine to handle request
		go hs.handleConnection(conn)
	}
}

