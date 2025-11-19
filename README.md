# HTTP Server and Parser
This repo features a Go implementation of an HTTP Parser as well as a server which handles requests.

### Running the HTTP Server

`go run cmd/httpserver/main.go`

Requests can be sent to the specified port using commands like:

`curl -v -X POST http://localhost:42069/hi`

The endpoint `/myproblem` will return an internal server error. The endpoint `/yourproblem` will return a bad request error.

The server also implements chunked encoding which you can test at the `/httpbin` endpoint. An example command to see the raw chunked response:

`echo -e "GET /httpbin/stream/100 HTTP/1.1\r\nHost: localhost:42069\r\nConnection: close\r\n\r\n" | nc localhost 42069`

### HTTP Parser

You can also see the parsed output of the HTTP request sent to the server by running the following:

`go run cmd/tcplistener/main.go`

Then send HTTP requests to the specified port.
