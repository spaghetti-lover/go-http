package utils

import (
	"errors"
	"strings"
	"testing"
)

// Mock ReadCloser for testing
type mockReadCloser struct {
	*strings.Reader
	closed bool
	err    error
}

func (m *mockReadCloser) Close() error {
	m.closed = true
	return m.err
}

func newMockReadCloser(content string) *mockReadCloser {
	return &mockReadCloser{
		Reader: strings.NewReader(content),
		closed: false,
	}
}

func newMockReadCloserWithError(content string, closeErr error) *mockReadCloser {
	return &mockReadCloser{
		Reader: strings.NewReader(content),
		closed: false,
		err:    closeErr,
	}
}

func TestGetLinesChannel_Normal(t *testing.T) {
	content := "line1\nline2\nline3"
	reader := newMockReadCloser(content)

	lineChannel, errorChannel := GetLinesChannel(reader)

	expectedLines := []string{"line1", "line2", "line3"}
	var receivedLines []string

	for line := range lineChannel {
		receivedLines = append(receivedLines, line)
	}

	// Check for errors
	select {
	case err := <-errorChannel:
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	default:
		// No error, which is expected
	}

	if len(receivedLines) != len(expectedLines) {
		t.Errorf("Expected %d lines, got %d", len(expectedLines), len(receivedLines))
	}

	for i, expected := range expectedLines {
		if i >= len(receivedLines) || receivedLines[i] != expected {
			t.Errorf("Expected line %d to be '%s', got '%s'", i, expected, receivedLines[i])
		}
	}

	if !reader.closed {
		t.Error("Expected reader to be closed")
	}
}

func TestGetLinesChannel_EmptyFile(t *testing.T) {
	reader := newMockReadCloser("")

	lineChannel, errorChannel := GetLinesChannel(reader)

	var receivedLines []string
	for line := range lineChannel {
		receivedLines = append(receivedLines, line)
	}

	// Check for errors
	select {
	case err := <-errorChannel:
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	default:
		// No error, which is expected
	}

	if len(receivedLines) != 0 {
		t.Errorf("Expected 0 lines, got %d", len(receivedLines))
	}

	if !reader.closed {
		t.Error("Expected reader to be closed")
	}
}

func TestGetLinesChannel_SingleLine(t *testing.T) {
	reader := newMockReadCloser("single line")

	lineChannel, errorChannel := GetLinesChannel(reader)

	var receivedLines []string
	for line := range lineChannel {
		receivedLines = append(receivedLines, line)
	}

	// Check for errors
	select {
	case err := <-errorChannel:
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	default:
		// No error, which is expected
	}

	if len(receivedLines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(receivedLines))
	}

	if receivedLines[0] != "single line" {
		t.Errorf("Expected 'single line', got '%s'", receivedLines[0])
	}

	if !reader.closed {
		t.Error("Expected reader to be closed")
	}
}

func TestGetLinesChannel_CloseError(t *testing.T) {
	closeError := errors.New("close error")
	reader := newMockReadCloserWithError("line1\nline2", closeError)

	lineChannel, errorChannel := GetLinesChannel(reader)

	var receivedLines []string
	for line := range lineChannel {
		receivedLines = append(receivedLines, line)
	}

	// Check for errors
	select {
	case err := <-errorChannel:
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	default:
		// No error, which is expected for scanner errors
	}

	expectedLines := []string{"line1", "line2"}
	if len(receivedLines) != len(expectedLines) {
		t.Errorf("Expected %d lines, got %d", len(expectedLines), len(receivedLines))
	}

	if !reader.closed {
		t.Error("Expected reader to be closed")
	}
}

func TestGetLinesChannel_ChannelsClosed(t *testing.T) {
	reader := newMockReadCloser("test line")

	lineChannel, errorChannel := GetLinesChannel(reader)

	// Consume all lines
	for range lineChannel {
	}

	// Check that channels are closed
	select {
	case _, ok := <-lineChannel:
		if ok {
			t.Error("Expected line channel to be closed")
		}
	default:
		t.Error("Line channel should be closed and readable")
	}

	select {
	case _, ok := <-errorChannel:
		if ok {
			t.Error("Expected error channel to be closed")
		}
	default:
		t.Error("Error channel should be closed and readable")
	}
}
