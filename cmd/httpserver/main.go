package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"http/internal/headers"
	"http/internal/request"
	"http/internal/response"
	"http/internal/server"
)

const (
	port    = 42069
	httpbin = "/httpbin/"
)

const (
	html400 = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

	html500 = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

	html200 = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
)

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, httpbin) {
		handleHTTPBin(w, req)
		return
	}

	h := headers.Headers{
		"content-type": "text/plain",
		"connection":   "close",
	}
	var msg string
	var statusCode response.StatusCode

	if req.RequestLine.RequestTarget == "/yourproblem" {
		statusCode = response.StatusCode400
		msg = html400
	} else if req.RequestLine.RequestTarget == "/myproblem" {
		statusCode = response.StatusCode500
		msg = html500
	} else {
		statusCode = response.StatusCode200
		msg = html200
	}
	w.WriteStatusLine(statusCode)
	h["content-length"] = fmt.Sprintf("%d", len(msg))
	w.WriteHeaders(h)
	w.WriteBody([]byte(msg))
}

func handleHTTPBin(w *response.Writer, req *request.Request) {
	targetSuffix := strings.TrimPrefix(req.RequestLine.RequestTarget, httpbin)
	targetURL := fmt.Sprintf("https://httpbin.org/%s", targetSuffix)
	resp, err := http.Get(targetURL)
	if err != nil {
		fmt.Printf("Error making HTTP request: %v\n", err)
		w.WriteStatusLine(response.StatusCode500)
		w.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}
	defer resp.Body.Close()
	w.WriteStatusLine(response.StatusCode200)
	h := headers.Headers{
		"content-type":      "text/plain",
		"connection":        "close",
		"transfer-encoding": "chunked",
	}
	w.WriteHeaders(h)
	fullResp := make([]byte, 1024*128)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		fmt.Printf("Read %d bytes from response body\n", n)
		if err != nil && err.Error() != "EOF" {
			fmt.Printf("Error reading response body: %v\n", err)
			break
		}
		var b bytes.Buffer
		chunk := string(buf[:n])

		// Write trailers when last line has been read.
		if n == 0 || err != nil && err.Error() == "EOF" {
			fmt.Fprintf(&b, "0\r\n")
			hash := sha256.Sum256(fullResp)
			t := headers.Headers{
				"x-content-sha256": fmt.Sprintf("%x", hash),
				"x-content-length": fmt.Sprintf("%d", len(fullResp)),
			}
			w.WriteTrailers(t)
			break
		}

		copy(fullResp[len(fullResp):], buf[:n])
		fmt.Fprintf(&b, "%x\r\n%s\r\n", n, chunk)
		encodedBytes := b.Bytes()
		w.WriteChunkedBody(encodedBytes)
	}
	w.WriteChunkedBodyDone()
}
