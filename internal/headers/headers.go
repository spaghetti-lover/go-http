package headers

import (
	"bytes"
	"fmt"
	"strings"
)

// isToken validates that str contains only valid token characters per RFC 9110
// token = 1*tchar
// tchar = "!" / "#" / "$" / "%" / "&" / "'" / "*" / "+" / "-" / "." /
//
//	"^" / "_" / "`" / "|" / "~" / DIGIT / ALPHA
func isToken(str []byte) bool {
	// token must be at least length 1
	if len(str) == 0 {
		return false
	}

	for _, ch := range str {
		// Check ALPHA (A-Z, a-z)
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
			continue
		}

		// Check DIGIT (0-9)
		if ch >= '0' && ch <= '9' {
			continue
		}

		// Check special tchar characters
		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			continue
		default:
			return false
		}
	}

	return true
}

var rn = []byte("\r\n")

func parseHeader(fieldLine []byte) (name, value string, err error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed field line")
	}

	nameBytes := parts[0]
	valueBytes := bytes.TrimSpace(parts[1])

	// Check for trailing space in header name
	if bytes.HasSuffix(nameBytes, []byte(" ")) {
		return "", "", fmt.Errorf("malformed field name")
	}

	// Check for empty header name
	if len(nameBytes) == 0 {
		return "", "", fmt.Errorf("empty field name")
	}

	return string(nameBytes), string(valueBytes), nil
}

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Get(name string) string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)

	if v, ok := h.headers[name]; ok {
		h.headers[name] = fmt.Sprintf("%s, %s", v, value)
	} else {
		h.headers[name] = value
	}
}

func (h *Headers) All() map[string]string {
	return h.headers
}

func (h *Headers) Parse(data []byte) (bytesRead int, done bool, err error) {
	read := 0
	isDone := false

	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break
		}

		// Empty line indicates end of headers
		if idx == 0 {
			isDone = true
			read += len(rn)
			break
		}

		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}

		if !isToken([]byte(name)) {
			return 0, false, fmt.Errorf("malformed header name")
		}

		h.Set(name, value)
		read += idx + len(rn)
	}
	return read, isDone, nil
}
