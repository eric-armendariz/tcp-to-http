package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

const SEPARATOR = "\r\n"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	if bytes.Index(data, []byte(SEPARATOR)) == -1 {
		return 0, false, nil
	}
	endIdx := bytes.Index(data, []byte(SEPARATOR))
	if endIdx == 0 {
		return len(SEPARATOR), true, nil
	}

	line := string(data[:endIdx])
	key, val, found := strings.Cut(line, ":")
	if !found {
		return 0, false, fmt.Errorf("invalid header: %s", line)
	} else if len(key) >= 1 && key[len(key)-1] == ' ' {
		return 0, false, fmt.Errorf("invalid header: %s", line)
	}

	key = strings.ToLower(strings.TrimSpace(key))
	val = strings.TrimSpace(val)
	validChars := makeValidCharTable()
	if len(key) <= 0 {
		return 0, false, fmt.Errorf("invalid header: %s", line)
	}
	for _, c := range key {
		if !validChars[c] {
			return 0, false, fmt.Errorf("invalid header: %s", line)
		}
	}

	if _, ok := h[key]; ok {
		h[key] += ", " + val
	} else {
		h[key] = val
	}
	return endIdx + len(SEPARATOR), false, nil
}

func (h Headers) Get(key string) (string, bool) {
	val, ok := h[key]
	return val, ok
}

func NewHeaders() Headers {
	return make(Headers)
}

func makeValidCharTable() map[rune]bool {
	validChars := make(map[rune]bool)
	for c := 'a'; c <= 'z'; c++ {
		validChars[c] = true
	}
	for c := 'A'; c <= 'Z'; c++ {
		validChars[c] = true
	}
	for c := '0'; c <= '9'; c++ {
		validChars[c] = true
	}
	for _, c := range "!#$%&'*+-.^_`|~" {
		validChars[c] = true
	}
	return validChars
}
