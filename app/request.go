package main

type RequestLine struct {
	method string;
	path string;
	httpVersion string;
}

type Request struct {
	requestLine RequestLine;
	headers map[string]string;
}
