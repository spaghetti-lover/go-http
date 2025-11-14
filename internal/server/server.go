package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/spaghetti-lover/go-http/internal/request"
	"github.com/spaghetti-lover/go-http/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func (he *HandlerError) Write(w io.Writer) error {
	// Write status line
	err := response.WriteStatusLine(w, he.StatusCode)
	if err != nil {
		return fmt.Errorf("error writing status line: %w", err)
	}

	// Get headers with content length
	headers := response.GetDefaultHeaders(len(he.Message))

	// Write headers
	err = response.WriteHeaders(w, headers)
	if err != nil {
		return fmt.Errorf("error writing headers: %w", err)
	}

	// Write error message body
	_, err = w.Write([]byte(he.Message))
	if err != nil {
		return fmt.Errorf("error writing body: %w", err)
	}

	return nil
}

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
			// Ignote errors after servers is closed
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

	// Create a new empty bytes.Buffer for the handler to write to
	buf := &bytes.Buffer{}

	// Call the handler function
	handleErr := s.handler(buf, req)

	// If the handler errors, write the error to the connection
	if handleErr != nil {
		err = handleErr.Write(conn)
		if err != nil {
			log.Printf("Error writing handler error: %v", err)
		}
		return
	}

	// If the handler succeeds:
	// Create new default response headers
	headers := response.GetDefaultHeaders(buf.Len())

	// Write status line
	err = response.WriteStatusLine(conn, response.OK)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	// Write the headers
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	// Write the response body from the handler's buffer
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		log.Printf("Error writing response body: %v", err)
		return
	}
}
