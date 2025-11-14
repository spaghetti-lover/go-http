package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
	// Check if this is a proxy request to httpbin
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		handleProxy(w, req)
		return
	}

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

func handleProxy(w *response.Writer, req *request.Request) {
	// Remove /httpbin prefix and build httpbin.org URL
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	url := "https://httpbin.org" + path

	log.Printf("Proxying request to: %s", url)

	// Make request to httpbin.org
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error making requét to httpbin.org: %v", err)
		writeError(w, response.InternalServerError, "Failed to proxy request")
		return
	}
	defer resp.Body.Close()

	// Write status line
	statusCode := response.StatusCode(fmt.Sprintf("%d", resp.StatusCode))
	err = w.WriteStatusLine(statusCode)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	// Create headers - remove Content-Length and add Transfer-Encoding
	h := headers.NewHeaders()

	// Copy headers from httpbin response (except Content-Length)
	for key, values := range resp.Header {
		if strings.ToLower(key) != "content-length" {
			for _, value := range values {
				h.Set(key, value)
			}
		}
	}

	// Add Transfer-Encoding: chunked
	h.Override("Transfer-Encoding", "chunked")
	h.Set("Connection", "close")

	// Write headers
	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	// Stream response body in chunks
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)

		if n > 0 {
			log.Printf("Read %d bytes from httpbin.org", n)

			// Write chunk
			_, writeErr := w.WriteChunkedBody(buf[:n])
			if writeErr != nil {
				log.Printf("Error writing chunk: %v", writeErr)
				return
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Error reading from httpbin.org: %v", err)
			break
		}
	}

	// Write final chunk
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		log.Printf("Erorr writing final chunk: %v", err)
		return
	}

	log.Println("Proxy request completed successfully")
}

func writeError(w *response.Writer, statusCode response.StatusCode, message string) {
	err := w.WriteStatusLine(statusCode)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	h := headers.NewHeaders()
	h.Set("Content-Length", strconv.Itoa(len(message)))
	h.Set("Connection", "close")
	h.Override("Content-Type", "text/plain")

	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	_, err = w.WriteBody([]byte(message))
	if err != nil {
		log.Printf("Error writing body: %v", err)
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
