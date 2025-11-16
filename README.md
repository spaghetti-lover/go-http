# Go-http

Lighweight library for HTTP/1.1 inspired by Primeagen

## Usage as a library

### 1. Installation

```bash
go get github.com/spaghetti-lover/go-http
```

### 2. Quick Start

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/spaghetti-lover/go-http/pkg/server"
    "github.com/spaghetti-lover/go-http/pkg/response"
    "github.com/spaghetti-lover/go-http/pkg/headers"
)

func main() {
    // Create handler
    handler := func(w *response.Writer, req *request.Request) {
        // Write status
        w.WriteStatusLine(response.OK)

        // Write headers
        h := headers.NewHeaders()
        h.Set("Content-Length", "13")
        h.Override("Content-Type", "text/plain")
        w.WriteHeaders(h)

        // Write body
        w.WriteBody([]byte("Hello, world!"))
    }

    // Start server
    srv, err := server.Serve(8080, handler)
    if err != nil {
        log.Fatal(err)
    }
    defer srv.Close()

    // Wait for interrupt
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
    <-sig
}
```

### 3. Public API

#### Server

```go
// Start server on port with handler
server.Serve(port int, handler Handler) (*Server, error)

// Handler signature
type Handler func(w *response.Writer, req *request.Request)
```

#### Response Writer

```go
// Status codes
response.OK                  // 200
response.BadRequest          // 400
response.InternalServerError // 500

// Write methods
w.WriteStatusLine(statusCode StatusCode) error
w.WriteHeaders(h *headers.Headers) error
w.WriteBody(p []byte) (int, error)

// Chunked encoding
w.WriteChunkedBody(p []byte) (int, error)
w.WriteChunkedBodyDone() (int, error)
w.WriteTrailers(h *headers.Headers) error
```

#### Headers

```go
h := headers.NewHeaders()
h.Set("Header-Name", "value")      // Add or append
h.Override("Header-Name", "value") // Replace
value := h.Get("Header-Name")      // Get value
```

#### Request

```go
// Request fields
req.RequestLine.Method        // GET, POST, etc.
req.RequestLine.RequestTarget // /path?query
req.RequestLine.HttpVersion   // HTTP/1.1
req.Headers                   // *headers.Headers
req.Body                      // []byte
```

### 4. Advanced Examples

#### Chunked Response with Trailers

```go
handler := func(w *response.Writer, req *request.Request) {
    w.WriteStatusLine(response.OK)

    h := headers.NewHeaders()
    h.Override("Transfer-Encoding", "chunked")
    h.Set("Trailer", "X-Content-Length")
    w.WriteHeaders(h)

    // Send chunks
    data1 := []byte("chunk1")
    w.WriteChunkedBody(data1)

    data2 := []byte("chunk2")
    w.WriteChunkedBody(data2)

    // End chunks
    w.WriteChunkedBodyDone()

    // Send trailers
    trailers := headers.NewHeaders()
    trailers.Set("X-Content-Length", "12")
    w.WriteTrailers(trailers)
}
```

#### Serve Binary File

```go
handler := func(w *response.Writer, req *request.Request) {
    if req.RequestLine.RequestTarget == "/video" {
        data, _ := os.ReadFile("video.mp4")

        w.WriteStatusLine(response.OK)

        h := headers.NewHeaders()
        h.Set("Content-Length", strconv.Itoa(len(data)))
        h.Override("Content-Type", "video/mp4")
        w.WriteHeaders(h)

        w.WriteBody(data)
    }
}
```

### 5. Notes

⚠️ This is an educational project - **not production-ready**

## Usage as a demo

1. Clone repo
2. Start the Server

```bash
# Using go run
go run  ./cmd/httpserver

# Using make
make run
```

3. Test endpoints

```bash
curl http://localhost:42069/

# 400 Bad Request
curl http://localhost:42069/yourproblem

# 500 Internal Server Error
curl http://localhost:42069/myproblem

# View in browser (need to add a video name "naruto.mp4" in folder asset)
open http://localhost:42069/video

# Download with curl (need to add a video name "naruto.mp4" in folder asset)
curl http://localhost:42069/video --output video.mp4

# Check headers (need to add a video name "naruto.mp4" in folder asset)
curl -I http://localhost:42069/video

# Stream data in chunks from httpbin.org
curl -v http://localhost:42069/httpbin/get

# See raw chunked response
echo -e "GET /httpbin/stream/3 HTTP/1.1\r\nHost: localhost:42069\r\nConnection: close\r\n\r\n" | nc localhost 42069

# View with curl --raw to see trailers
curl --raw http://localhost:42069/httpbin/get
```
