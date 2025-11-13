package request

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/spaghetti-lover/go-http/internal/headers"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

type Line struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Line) ValidHTTP() bool {
	return r.HttpVersion == "1.1"
}

type Request struct {
	RequestLine Line
	Headers     *headers.Headers
	Body        []byte
	state       parserState
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

var ErrMalformedRequestLine = fmt.Errorf("malformed start line")
var ErrUnsupportedHTTPVersion = fmt.Errorf("unsupported http version")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")
var ErrBodyTooLarge = fmt.Errorf("body exceeds content-length")
var SEPARATOR = []byte("\r\n")

func parseRequestLine(b []byte) (*Line, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrMalformedRequestLine
	}

	rl := &Line{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case StateError:
		return 0, ErrorRequestInErrorState
	case StateInit:
		rl, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if n == 0 {
			return 0, nil
		}

		r.RequestLine = *rl
		r.state = StateHeaders
		return n, nil

	case StateHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if done {
			r.state = StateBody
		}

		return n, nil

	case StateBody:
		contentLengthStr := r.Headers.Get("Content-Length")
		if contentLengthStr == "" {
			r.state = StateDone
			return 0, nil
		}

		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("invalid content-length: %w", err)
		}

		// Append all available data to body
		r.Body = append(r.Body, data...)

		// Check if body exceeds content length
		if len(r.Body) > contentLength {
			return 0, ErrBodyTooLarge
		}

		// Check if receiving all the body data
		if len(r.Body) == contentLength {
			r.state = StateDone
		}

		// Report that consuming all the data
		return len(data), nil

	case StateDone:
		return 0, nil
	}

	return 0, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.state != StateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}

		if n == 0 {
			break
		}

		totalBytesParsed += n
	}

	return totalBytesParsed, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func (r *Request) error() bool {
	return r.state == StateError
}

func (r *Request) String() string {
	var buf bytes.Buffer

	buf.WriteString("Request line:\n")
	buf.WriteString(fmt.Sprintf("- Method: %s\n", r.RequestLine.Method))
	buf.WriteString(fmt.Sprintf("- Target: %s\n", r.RequestLine.RequestTarget))
	buf.WriteString(fmt.Sprintf("- Version: %s\n", r.RequestLine.HttpVersion))

	buf.WriteString("Headers:\n")

	// Get all headers and sort keys for consistent output
	allHeaders := r.Headers.All()
	keys := make([]string, 0, len(allHeaders))
	for k := range allHeaders {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		buf.WriteString(fmt.Sprintf("- %s: %s\n", k, allHeaders[k]))
	}

	buf.WriteString("Body:\n")
	buf.WriteString(string(r.Body))
	buf.WriteString("\n")

	return buf.String()
}

func FromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	// NOTE: buffer could get overrun... a header/body that exceed 1k byte would do that
	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n

		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	if request.error() {
		return nil, fmt.Errorf("request parsing failed")
	}

	return request, nil
}
