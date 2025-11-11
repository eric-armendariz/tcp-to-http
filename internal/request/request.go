package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte
	state       parserState
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

type parserState int

var (
	cmds = map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"DELETE": true,
	}
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{
		state: 0,
	}

	buf := make([]byte, 8)
	readIdx := 0
	for req.state != 1 {
		if readIdx >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf[:readIdx])
			buf = newBuf
		}

		n, err := reader.Read(buf[readIdx:])
		if err == io.EOF {
			req.state = 1
			break
		}
		if err != nil {
			return nil, err
		}
		readIdx += n

		consumed, err := req.parse(buf[:readIdx])
		if err != nil {
			return nil, err
		}

		if consumed > 0 {
			copy(buf, buf[consumed:])
			readIdx -= consumed
		}
	}

	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	endIdx := -1
	for i := 0; i < len(data)-1; i++ {
		if data[i] == '\r' && data[i+1] == '\n' {
			endIdx = i
			break
		}
	}
	if endIdx == -1 {
		return nil, 0, nil
	}
	line := string(data[:endIdx])

	requestLine := strings.Split(string(line), " ")
	if len(requestLine) != 3 {
		return nil, 0, fmt.Errorf("invalid request line: %s", string(line))
	}

	method, reqTarget := requestLine[0], requestLine[1]
	version := strings.TrimPrefix(requestLine[2], "HTTP/")
	if version != "1.1" {
		return nil, 0, fmt.Errorf("unsupported HTTP version: %s", version)
	}
	if !cmds[method] || !isAllUppercase(method) {
		return nil, 0, fmt.Errorf("unsupported command: %s", requestLine[0])
	}
	if !strings.HasPrefix(reqTarget, "/") {
		return nil, 0, fmt.Errorf("request target must start with /: %s", reqTarget)
	}
	reqLine := &RequestLine{
		Method:        requestLine[0],
		RequestTarget: requestLine[1],
		HttpVersion:   strings.TrimPrefix(requestLine[2], "HTTP/"),
	}
	return reqLine, endIdx + 2, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == 1 {
		return 0, fmt.Errorf("request already parsed")
	}
	reqLine, bytesRead, err := parseRequestLine(data)
	if err != nil {
		return 0, err
	}
	if bytesRead == 0 {
		return 0, nil
	}
	r.RequestLine = *reqLine
	r.state = 1
	return bytesRead, nil
}

func isAllUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
