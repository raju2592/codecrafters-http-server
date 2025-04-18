package main

import (
	"fmt"
	"net"
	"os"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit


func handleConnection(conn net.Conn) {
	cr := NewConnectionReader(conn)
	reqLine, err := parseRequestLine(cr)
	if err != nil {
		fmt.Println("error parsing request line", err)
		return
	}

	fmt.Printf("%s %s %s", reqLine.method, reqLine.path, reqLine.httpVersion)

	res := []byte("HTTP/1.1 404 Not Found\r\n\r\n")

	if reqLine.path == "/" {
		res = []byte("HTTP/1.1 200 OK\r\n\r\n")
	}

	conn.Write(res)
	conn.Close()
}

func HandleEcho(req *HandlerReqest) *HandlerResponse {
	str, _ := req.pathParams["str"]
	return &HandlerResponse{
		status: "200 OK",
		headers: map[string]string {
			"Content-Type": "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(str)),
		},
		body: []byte(str),
	}
}

func HandleHome(req *HandlerReqest) *HandlerResponse {
	return &HandlerResponse{
		status: "200 OK",
	}
}

func HandleUserAgent(req *HandlerReqest) *HandlerResponse {
	userAgent, _ := req.request.headers["User-Agent"]
	return &HandlerResponse{
		status: "200 OK",
		headers: map[string]string {
			"Content-Type": "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(userAgent)),
		},
		body: []byte(userAgent), 
	}
}

func main() {
	// Uncomment this block to pass the first stage
	//
	// l, err := net.Listen("tcp", "0.0.0.0:4221")
	// if err != nil {
	// 	fmt.Println("Failed to bind to port 4221")
	// 	os.Exit(1)
	// }
	
	// conn, err := l.Accept()
	
	// if err != nil {
	// 	fmt.Println("Error accepting connection: ", err.Error())
	// 	os.Exit(1)
	// }

	// handleConnection(conn)
	server := NewServer()
	server.RegisterRoute("GET", "/echo/{str}", HandleEcho)
	server.RegisterRoute("GET", "/user-agent", HandleUserAgent)
	server.RegisterRoute("GET", "/", HandleHome)
	server.Listen("0.0.0.0:4221")
}
