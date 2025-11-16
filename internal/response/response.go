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

type writerState string

const (
	stateInit     writerState = "init"
	stateStatus   writerState = "status"
	stateHeaders  writerState = "headers"
	stateBody     writerState = "body"
	stateTrailers writerState = "trailers"
)

type Writer struct {
	writer io.Writer
	state  writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state:  stateInit,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != stateInit {
		return fmt.Errorf("WriteStatusLine must be called first")
	}

	var statusLine string
	switch statusCode {
	case OK:
		statusLine = "HTTP/1.1 200 OK"
	case BadRequest:
		statusLine = "HTTP/1.1 400 Bad Request"
	case InternalServerError:
		statusLine = "HTTP/1.1 500 internal Server Error"
	default:
		statusLine = "HTTP/1.1 " + string(statusCode)
	}

	_, err := w.writer.Write([]byte(statusLine + "\r\n"))
	if err != nil {
		return fmt.Errorf("error writing status line: %w", err)
	}

	w.state = stateStatus
	return nil
}

func (w *Writer) WriteHeaders(h *headers.Headers) error {
	if w.state != stateStatus {
		return fmt.Errorf("WriteHeaders must be called after WriteStatusLine")
	}

	allHeaders := h.All()

	for key, value := range allHeaders {
		headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.writer.Write([]byte(headerLine))
		if err != nil {
			return fmt.Errorf("error writing header: %w", err)
		}
	}

	// Write empty line to separate headers from body
	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("error writing header separator: %w", err)
	}

	w.state = stateHeaders
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != stateHeaders {
		return 0, fmt.Errorf("WriteBody must be called after WriteHeaders")
	}

	n, err := w.writer.Write(p)
	if err != nil {
		return n, fmt.Errorf("error writing body: %w", err)
	}

	w.state = stateBody
	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != stateHeaders && w.state != stateBody {
		return 0, fmt.Errorf("WriteChunkedBody must be called after WriteHeaders")
	}

	if len(p) == 0 {
		return 0, nil
	}

	// Write chunk size in hexadecimal
	chunkSize := fmt.Sprintf("%X\r\n", len(p))
	_, err := w.writer.Write([]byte(chunkSize))
	if err != nil {
		return 0, fmt.Errorf("error writing chunk size: %w", err)
	}

	// Write chunk data
	n, err := w.writer.Write(p)
	if err != nil {
		return n, fmt.Errorf("error writing chunk data: %w", err)
	}

	// Write trailing CRLF
	_, err = w.writer.Write([]byte("\r\n"))
	if err != nil {
		return n, fmt.Errorf("error writing chunk trailing CRLF: %w", err)
	}

	w.state = stateBody
	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != stateHeaders && w.state != stateBody {
		return 0, fmt.Errorf("WriteChunkedBodyDone muse be called after WriteHeaders or WriteChunkedBody")
	}

	// Write final chunk: "0\r\n\r\n"
	n, err := w.writer.Write([]byte("0\r\n\r\n"))
	if err != nil {
		return n, fmt.Errorf("error writing final chunk: %w", err)
	}

	w.state = stateBody
	return n, nil
}

func (w *Writer) WriteTrailers(h *headers.Headers) error {
	if w.state != stateBody {
		return fmt.Errorf("WriteTrailers must be called after WriteChunkedBodyDone")
	}

	allHeaders := h.All()

	for key, value := range allHeaders {
		trailerLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.writer.Write([]byte(trailerLine))
		if err != nil {
			return fmt.Errorf("error writing trailer: %w", err)
		}
	}

	// Write final CRLF to end trailers
	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("error writing trailer separator: %w", err)
	}

	w.state = stateTrailers
	return nil
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var statusLine string
	switch statusCode {
	case OK:
		statusLine = "HTTP/1.1 200 OK"
	case BadRequest:
		statusLine = "HTTP/1.1 400 Bad Request"
	case InternalServerError:
		statusLine = "HTTP/1.1 500 internal Server Error"
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
