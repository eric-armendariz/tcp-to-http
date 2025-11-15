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

	key, val = strings.TrimSpace(key), strings.TrimSpace(val)
	if len(key) <= 0 {
		return 0, false, fmt.Errorf("invalid header: %s", line)
	}
	h[key] = val
	return endIdx + len(SEPARATOR), false, nil
}

func NewHeaders() Headers {
	return make(Headers)
}
