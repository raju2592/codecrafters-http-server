package main

type RequestLine struct {
	method string;
	target string;
	httpVersion string;
}

type Requst struct {
	requestLine RequestLine
}
