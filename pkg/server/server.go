package server

import (
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/spaghetti-lover/go-http/pkg/request"
	"github.com/spaghetti-lover/go-http/pkg/response"
)

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Printf("Error listening to port: %v", err)
		return nil, err
	}

	log.Println("Server listening on port", port)

	server := &Server{
		listener: listener,
		handler:  handler,
	}

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// Ignore errors after server is closed
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	// Parse the request from the connection
	req, err := request.FromReader(conn)
	if err != nil {
		log.Printf("Error reading from %s: %v", conn.RemoteAddr(), err)
		return
	}

	// Create a response writer
	writer := response.NewWriter(conn)

	// Call the handler function
	s.handler(writer, req)
}
