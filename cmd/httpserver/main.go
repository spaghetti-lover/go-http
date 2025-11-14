package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/spaghetti-lover/go-http/internal/headers"
	"github.com/spaghetti-lover/go-http/internal/request"
	"github.com/spaghetti-lover/go-http/internal/response"
	"github.com/spaghetti-lover/go-http/internal/server"
)

const (
	html400 = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`
	html500 = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`
	html200 = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`
)

func handleRequest(w *response.Writer, req *request.Request) {
	var statusCode response.StatusCode
	var body string

	// Determine response based on request target
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		statusCode = response.BadRequest
		body = html400
	case "/myproblem":
		statusCode = response.InternalServerError
		body = html500
	default:
		statusCode = response.OK
		body = html200
	}

	// Write status line
	err := w.WriteStatusLine(statusCode)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	// Create headers with HTML content type
	h := headers.NewHeaders()
	h.Set("Content-Length", strconv.Itoa(len(body)))
	h.Set("Connection", "close")
	h.Override("Content-Type", "text/html")

	// Write headers
	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	// Write body
	_, err = w.WriteBody([]byte(body))
	if err != nil {
		log.Printf("Error writing body: %v", err)
		return
	}
}

func main() {
	const port = 42069
	srv, err := server.Serve(port, handleRequest)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
