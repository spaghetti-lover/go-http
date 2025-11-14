package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spaghetti-lover/go-http/internal/request"
	"github.com/spaghetti-lover/go-http/internal/response"
	"github.com/spaghetti-lover/go-http/internal/server"
)

func handleRequest(w io.Writer, req *request.Request) *server.HandlerError {
	// Check the request target (path)
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: response.BadRequest,
			Message:    "Your problem is not my problem\n",
		}
	case "/myproblem":
		return &server.HandlerError{
			StatusCode: response.InternalServerError,
			Message:    "Woopsie, my bad\n",
		}
	default:
		// Write the success message to the response body
		_, err := w.Write([]byte("All good, frfr\n"))
		if err != nil {
			return &server.HandlerError{
				StatusCode: response.InternalServerError,
				Message:    "Error writing response\n",
			}
		}
		return nil
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
