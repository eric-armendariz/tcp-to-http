package request

import (
	"bytes"
	"fmt"
	"http/internal/headers"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       parserState
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

type parserState int

const (
	SEPARATOR = "\r\n"

	stateInitialized    = 0
	stateParsingHeaders = 1
	stateParsingBody    = 2
	stateDone           = 3
)

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
		state:   0,
		Headers: headers.NewHeaders(),
	}

	buf := make([]byte, 8)
	readIdx := 0
	for req.state != stateDone {
		consumed, err := req.parse(buf[:readIdx])
		if err != nil {
			return nil, err
		}

		if consumed > 0 {
			copy(buf, buf[consumed:])
			readIdx -= consumed
			continue
		}

		if readIdx >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf[:readIdx])
			buf = newBuf
		}

		n, err := reader.Read(buf[readIdx:])
		if err == io.EOF {
			if req.state == stateParsingBody {
				return nil, fmt.Errorf("unexpected EOF while reading body")
			}
			req.state = stateDone
			break
		}
		if err != nil {
			return nil, err
		}
		readIdx += n
	}

	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	endIdx := bytes.Index(data, []byte(SEPARATOR))
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
	return reqLine, endIdx + len(SEPARATOR), nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case stateDone:
		return 0, fmt.Errorf("request already parsed")
	case stateInitialized:
		reqLine, bytesRead, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if bytesRead == 0 {
			return 0, nil
		}
		r.RequestLine = *reqLine
		r.state = stateParsingHeaders
		return bytesRead, nil
	case stateParsingHeaders:
		bytesRead, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if !done {
			return bytesRead, nil
		}
		r.state = stateParsingBody
		return bytesRead, nil
	case stateParsingBody:
		sizeVal, ok := r.Headers.Get("content-length")
		if !ok {
			r.state = stateDone
			return 0, nil
		}
		contentSize, err := strconv.Atoi(sizeVal)
		if err != nil {
			return 0, fmt.Errorf("invalid content-length header: %s", sizeVal)
		}
		// Need more data to parse
		if len(data) < contentSize {
			return 0, nil
		}
		// Data size does not match content-length
		if len(data) > contentSize {
			return 0, fmt.Errorf("invalid content-length header: %d", contentSize)
		}
		r.Body = make([]byte, contentSize)
		copy(r.Body, data)
		r.state = stateDone
		return contentSize, nil
	default:
		return 0, fmt.Errorf("invalid state: %d", r.state)
	}
}

func isAllUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
