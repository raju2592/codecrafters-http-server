package main

import (
	"errors"
	"fmt"
	"log"
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
		body: NewStaticReadStream([]byte(str)),
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
		body: NewStaticReadStream([]byte(userAgent)), 
	}
}

func HandleFileGet(req *HandlerReqest) *HandlerResponse {
	file, _ := req.pathParams["filename"]
	root := os.Args[2]
	filePath := root + file

	fmt.Println("path ", filePath)
	fs, err := os.Stat(filePath)

	if errors.Is(err, os.ErrNotExist) {
		return &HandlerResponse{
			status: "404 Not Found",
		}	
	}

	if err != nil {
		log.Fatal("Failed to get file info")
	}

	stream, err := NewFileReadStream(filePath)

	if err != nil {
		log.Fatal("Failed to read file")
	}

	return &HandlerResponse{
		status: "200 OK",
		headers: map[string]string {
			"Content-Type": "application/octet-stream",
			"Content-Length": fmt.Sprintf("%d", fs.Size()),
		},
		body: stream,
	}
}

func HandleFilePost(req *HandlerReqest) *HandlerResponse {
	fileName, _ := req.pathParams["filename"]
	root := os.Args[2]
	filePath := root + fileName

	// fmt.Println("path ", filePath)
	// fs, err := os.Stat(filePath)

	// if errors.Is(err, os.ErrNotExist) {
	// 	return &HandlerResponse{
	// 		status: "404 Not Found",
	// 	}	
	// }

	file, err := os.Create(filePath)

	if err != nil {
		log.Fatal("Failed to get file info")
	}

	for {
		select {
		case data := <- req.request.body.DataC():
			if data == nil {
				goto done
			}

			fmt.Println(data);
			
			_, err := file.Write(data)
			if err != nil {
				log.Fatal("Failed to write file")
			}
		case err := <- req.request.body.ErrC():
			log.Fatal("Failed to read request body: %s", err.Error())
		}
	}

	done: err = file.Close()
	if err != nil {
		log.Fatal("Failed to close file")
	}

	return &HandlerResponse{
		status: "201 Created",
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

	fmt.Println(os.Args)

	// handleConnection(conn)
	server := NewServer()
	server.RegisterRoute("GET", "/echo/{str}", HandleEcho)
	server.RegisterRoute("GET", "/user-agent", HandleUserAgent)
	server.RegisterRoute("GET", "/files/{filename}", HandleFileGet)
	server.RegisterRoute("POST", "/files/{filename}", HandleFilePost)
	server.RegisterRoute("GET", "/", HandleHome)
	server.Listen("0.0.0.0:4221")
}
