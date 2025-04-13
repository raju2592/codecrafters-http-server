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

	fmt.Printf("%s %s %s", reqLine.method, reqLine.target, reqLine.httpVersion)

	res := []byte("HTTP/1.1 404 Not Found\r\n\r\n")

	if reqLine.target == "/" {
		res = []byte("HTTP/1.1 200 OK\r\n\r\n")
	}

	conn.Write(res)
	conn.Close()
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	
	conn, err := l.Accept()
	
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	handleConnection(conn)
}
