package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type HandlerResponse struct {
	status int
	headers map[string]string
	body []byte
}

type HandlerReqest struct {
	request *Request
	pathParams map[string]string
}

type Handler func(*HandlerReqest) *HandlerResponse

type Server struct {
	l net.Listener
	handlers map[string]map[string]Handler
}

func NewServer() *Server {
	return &Server{
		handlers: make(map[string]map[string]Handler),
	}
}

func (s *Server) Listen(address string) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	s.l = l;
	go s.accept()
	return nil
}

func (s *Server) accept() {
	conn, err := s.l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	cr := NewConnectionReader(conn)
	req, err := parseRequst(cr)
	fmt.Printf("Error parsing request: %s", err.Error())

	s.handleRequest(req)

}

type PathMatch struct {
	pathParams map[string]string
	handler Handler
}

func (s *Server) handleRequest(req *Request) *HandlerResponse {
	pathMatch := s.route(req)
	if pathMatch == nil {
		return &HandlerResponse{
			status: 404,
		}
	}

	handler := pathMatch.handler
	pathParams := pathMatch.pathParams

	return handler(&HandlerReqest{
		request: req,
		pathParams: pathParams,
	})
}

func (s *Server) route(req * Request) *PathMatch {
	method := req.requestLine.method
	requestPath := req.requestLine.path

	requestHandlers, ok := s.handlers[method]
	if !ok {
		return nil
	}

	for route, handler := range requestHandlers {
		pathMatch := matchPath(requestPath, route)
		if pathMatch != nil {
			return &PathMatch{
				pathParams: pathMatch,
				handler: handler,
			}
		}
	}

	return nil
}

func matchPath(requestPath string, routePath string) map[string]string {
	requestSegments := strings.Split(requestPath, "/")
	routeSegments := strings.Split(routePath, "/")

	if len(requestSegments) != len(routeSegments) {
		return nil
	}

	params := make(map[string]string)

	for i := 1; i < len(requestSegments); i++ {
		requestSeg := requestSegments[i]
		routeSeg := routeSegments[i]

		if isParam(routeSeg) {
			params[getParamName(routeSeg)] = requestSeg
		} else if (requestSeg != routeSeg) {
			return nil
		}
	}

	return params
}

func isParam(seg string) bool {
	return len(seg) > 2 && seg[0] == '{' && seg[len(seg) - 1] == '}'
}

func getParamName(seg string) string {
	return seg[1: len(seg) - 1]
}

func (s *Server) RegisterRoute(method string, path string, handler Handler) {
	methodHandlers, ok := s.handlers[method]
	if !ok {
		methodHandlers = make(map[string]Handler)
		s.handlers[method] = methodHandlers
	}

	methodHandlers[path] = handler
}
