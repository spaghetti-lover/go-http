package request

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	"github.com/spaghetti-lover/go-http/internal/headers"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
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
			r.state = StateDone
		}

		return n, nil

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
