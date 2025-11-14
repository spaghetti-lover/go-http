package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/spaghetti-lover/go-http/internal/headers"
)

type StatusCode string

const (
	OK                  StatusCode = "200"
	BadRequest          StatusCode = "400"
	InternalServerError StatusCode = "500"
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var statusLine string
	switch statusCode {
	case OK:
		statusLine = "HTTP/1.1 200 OK"
	case BadRequest:
		statusLine = "HTTP/1.1 400 Bad Request"
	case InternalServerError:
		statusLine = "HTTP/1.1 500 Internal Server Error"
	default:
		statusLine = "HTTP/1.1 " + string(statusCode)
	}

	_, err := w.Write([]byte(statusLine + "\r\n"))
	if err != nil {
		return fmt.Errorf("error writing status line: %w", err)
	}

	return nil
}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, h *headers.Headers) error {
	allHeaders := h.All()

	for key, value := range allHeaders {
		headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.Write([]byte(headerLine))
		if err != nil {
			return fmt.Errorf("error writing header: %w", err)
		}
	}

	// Write empty line to seperate headers from body
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("error writing header separator: %w", err)
	}

	return nil
}
