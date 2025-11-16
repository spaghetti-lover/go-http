# Go HTTP Server

A custom HTTP/1.1 server implementation built from scratch in Go, demonstrating low-level HTTP protocol handling without using the standard `net/http` package. And yep, this README is AI generated because this project is for learning purpose ^^. U can check my docs why i code like this or just wath Primeagen.

## 🎯 Overview

This project implements a fully functional HTTP/1.1 server that handles:

- Request parsing (request line, headers, body)
- Response writing with proper status codes and headers
- Chunked transfer encoding
- HTTP trailers
- Binary data (video streaming)
- Proxy requests to external APIs

## 🚀 Features

### Core HTTP Components

- **Request Parser**: Parses HTTP/1.1 requests including:

  - Request line (method, target, version)
  - Headers (with case-insensitive handling and RFC 9110 compliance)
  - Request body with Content-Length support

- **Response Writer**: State-machine based response writer that ensures proper HTTP response structure:

  - Status line
  - Headers
  - Body (text or binary)
  - Chunked encoding support
  - Trailers support

- **Header Management**:
  - RFC-compliant header parsing and validation
  - Case-insensitive header operations
  - Multiple value support for duplicate headers

### Advanced Features

- **Chunked Transfer Encoding**: Stream large responses in chunks
- **HTTP Trailers**: Send metadata after response body (hash, content length)
- **Binary Data Support**: Serve video files and other binary content
- **Proxy Handler**: Forward requests to external APIs (httpbin.org)
- **Graceful Shutdown**: Proper signal handling and server cleanup

## 📁 Project Structure

```
.
├── cmd/
│   ├── httpserver/      # Main HTTP server application
│   └── tcplistener/     # Basic TCP listener for debugging
├── internal/
│   ├── headers/         # HTTP header parsing and management
│   ├── request/         # HTTP request parsing
│   ├── response/        # HTTP response writing
│   ├── server/          # Server implementation
│   └── utils/           # Utility functions
├── docs/                # Documentation
├── assets/              # Static assets (ignored in git)
└── README.md
```

## 🛠️ Installation

### Prerequisites

- Go 1.24.1 or higher
- Make (optional)

### Setup

1. Clone the repository:

```bash
git clone https://github.com/spaghetti-lover/go-http.git
cd go-http
```

2. Download dependencies:

```bash
go mod download
```

3. (Optional) Download sample video for binary data testing:

```bash
mkdir assets
curl -o assets/vim.mp4 https://storage.googleapis.com/qvault-webapp-dynamic-assets/lesson_videos/vim-vs-neovim-prime.mp4
```

## 🎮 Usage

### Start the Server

```bash
# Using go run
go run ./cmd/httpserver

# Using make
make run
```

The server will start on port `42069`.

### Test Endpoints

#### Basic HTML Responses

```bash
# Success response
curl http://localhost:42069/

# 400 Bad Request
curl http://localhost:42069/yourproblem

# 500 internal Server Error
curl http://localhost:42069/myproblem
```

#### Video Streaming (Binary Data)

```bash
# View in browser
open http://localhost:42069/video

# Download with curl
curl http://localhost:42069/video --output video.mp4

# Check headers
curl -I http://localhost:42069/video
```

#### Proxy with Chunked Encoding

```bash
# Stream data in chunks from httpbin.org
curl -v http://localhost:42069/httpbin/get

# See raw chunked response
echo -e "GET /httpbin/stream/3 HTTP/1.1\r\nHost: localhost:42069\r\nConnection: close\r\n\r\n" | nc localhost 42069

# View with curl --raw to see trailers
curl --raw http://localhost:42069/httpbin/get
```

The proxy endpoint will:

- Forward requests to `https://httpbin.org`
- Stream response using chunked transfer encoding
- Add trailers with content SHA256 hash and length

## 🧪 Testing

Run all tests:

```bash
# Run tests
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Using make
make test
```

## 📚 Implementation Details

### Request Parsing

The request parser uses a state machine approach with these states:

- `StateInit`: Parse request line
- `StateHeaders`: Parse headers
- `StateBody`: Read body based on Content-Length
- `StateDone`: Parsing complete

### Response Writing

The response writer enforces proper HTTP response order:

1. `WriteStatusLine()` - must be called first
2. `WriteHeaders()` - must be called after status line
3. `WriteBody()` or `WriteChunkedBody()` - write response data
4. `WriteTrailers()` - optional, for chunked responses

### Chunked Encoding Format

```
<chunk-size-hex>\r\n
<chunk-data>\r\n
...
0\r\n
<trailers>\r\n
\r\n
```

## 🔧 Configuration

Server configuration is hardcoded in `main.go`:

- Port: `42069`
- Buffer size: `1024 bytes`
- Chunk size for proxy: `1024 bytes`

## 🚦 Signal Handling

The server handles graceful shutdown on:

- `SIGINT` (Ctrl+C)
- `SIGTERM`

## 📖 HTTP Protocol Notes

### HTTP/1.1 vs HTTP/2 vs HTTP/3

**HTTP/1.1** (this implementation):

- Text-based protocol
- One request per connection (or pipelining)
- Header compression: None
- Transport: TCP

**HTTP/2**:

- Binary protocol
- Multiplexing (multiple requests on single connection)
- Header compression (HPACK)
- Server push support
- Transport: TCP

**HTTP/3**:

- Built on QUIC (UDP-based)
- Mandates encryption
- Faster connection establishment
- Better performance on lossy networks

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📝 License

This project is open source and available for educational purposes.

## 🙏 Acknowledgments

- Built as a learning project to understand HTTP protocol internals
- Inspired by low-level network programming concepts
- RFC 9110 (HTTP Semantics) for protocol specifications

## 📧 Contact

For questions or feedback, please open an issue on GitHub.

---

**Note**: This is an educational project and should not be used in production environments. For production use cases, please use the standard Go `net/http` package or other battle-tested HTTP servers.
