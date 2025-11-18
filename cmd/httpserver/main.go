package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"http/internal/headers"
	"http/internal/request"
	"http/internal/response"
	"http/internal/server"
)

const port = 42069

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
