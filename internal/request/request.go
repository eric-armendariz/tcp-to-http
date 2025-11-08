package request

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

var (
	cmds = map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"DELETE": true,
	}
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	bufReader := bufio.NewReader(reader)
	line, _, err := bufReader.ReadLine()
	if err != nil {
		return nil, err
	}
	reqLine, err := parseRequestLine(string(line))
	if err != nil {
		return nil, err
	}
	req := &Request{
		RequestLine: *reqLine,
	}

	return req, nil
}

func parseRequestLine(line string) (*RequestLine, error) {
	requestLine := strings.Split(string(line), " ")
	if len(requestLine) != 3 {
		return nil, fmt.Errorf("invalid request line: %s", string(line))
	}

	method, reqTarget := requestLine[0], requestLine[1]
	version := strings.TrimPrefix(requestLine[2], "HTTP/")
	if version != "1.1" {
		return nil, fmt.Errorf("unsupported HTTP version: %s", version)
	}
	if !cmds[method] || !isAllUppercase(method) {
		return nil, fmt.Errorf("unsupported command: %s", requestLine[0])
	}
	if !strings.HasPrefix(reqTarget, "/") {
		return nil, fmt.Errorf("request target must start with /: %s", reqTarget)
	}
	reqLine := &RequestLine{
		Method:        requestLine[0],
		RequestTarget: requestLine[1],
		HttpVersion:   strings.TrimPrefix(requestLine[2], "HTTP/"),
	}
	return reqLine, nil
}

func isAllUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
