package response

import (
	"bytes"
	"fmt"
	"http/internal/headers"
	"io"
	"strconv"
)

type StatusCode int
type writerState int

type Writer struct {
	W     io.Writer
	state writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{W: w, state: stateInitial}
}

const (
	StatusCode200 StatusCode = 200
	StatusCode400 StatusCode = 400
	StatusCode500 StatusCode = 500

	stateInitial writerState = iota
	stateStatusWritten
	stateHeadersWritten
	stateBodyWritten
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != stateInitial {
		return fmt.Errorf("writer not in proper state")
	}
	w.state = stateStatusWritten
	switch statusCode {
	case StatusCode200:
		_, err := w.W.Write([]byte("HTTP/1.1 200 OK\r\n"))
		return err
	case StatusCode400:
		_, err := w.W.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		return err
	case StatusCode500:
		_, err := w.W.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		return err
	default:
		msg := fmt.Sprintf("HTTP/1.1 %d\r\n", statusCode)
		_, err := w.W.Write([]byte(msg))
		return err
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.Headers{}
	h["content-length"] = fmt.Sprintf("%d", contentLen)
	h["connection"] = "close"
	h["content-type"] = "text/plain"
	return h
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.state != stateStatusWritten {
		return fmt.Errorf("writer not in proper state")
	}
	for k, v := range h {
		_, err := w.W.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}
	_, err := w.W.Write([]byte("\r\n"))
	w.state = stateHeadersWritten
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != stateHeadersWritten {
		return 0, fmt.Errorf("writer not in proper state")
	}
	n, err := w.W.Write(p)
	w.state = stateBodyWritten
	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != stateHeadersWritten {
		return 0, fmt.Errorf("writer not in proper state")
	}

	endIdx := bytes.Index(p, []byte(headers.SEPARATOR))
	if endIdx == -1 {
		return 0, fmt.Errorf("invalid chunk")
	}
	hexStr := string(p[:endIdx])
	lineSize, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		fmt.Println("Error converting hexadecimal to integer:", err)
		return 0, err
	}

	idx := endIdx + len(headers.SEPARATOR)
	n, err := w.W.Write(p[idx : idx+int(lineSize)])
	if err != nil {
		return 0, err
	}
	if n != int(lineSize) {
		return 0, fmt.Errorf("short write")
	}

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	w.state = stateBodyWritten
	return 0, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	for k, v := range h {
		_, err := w.W.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}
	_, err := w.W.Write([]byte("\r\n"))
	return err
}
