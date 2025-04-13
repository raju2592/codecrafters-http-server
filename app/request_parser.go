package main

import (
	"errors"
	"strings"
)

func parseRequestSegment(cr *ConnectionReader) ([]byte, error) {
	ret := make([]byte, 1024)
	for {
		b, err := cr.getByte()
		if err != nil {
			return []byte{}, err
		}
		if b != byte('\r') {
			ret = append(ret, b)
			continue
		}

	  nb, err := cr.getByte()
		if err != nil {
			return []byte{}, err
		}

		if nb == byte('\n') {
			return ret, nil
		}

		ret = append(ret, b, nb)
	}
}

func parseRequestLine(cr *ConnectionReader) (*RequestLine, error) {
	segment, err := parseRequestSegment(cr)
	if err != nil {
		return nil, err
	}

	reqLine := string(segment)
	fields := strings.Fields(reqLine)

	if len(fields) != 3 {
		return nil, errors.New("Invalid request line")
	}

	req := &RequestLine{
		method: fields[0],
		path: fields[1],
		httpVersion: fields[2],
	}

	return req, nil
}

func parseRequst(cr *ConnectionReader) (*Request, error) {
	reqLine, err := parseRequestLine(cr)
	if err != nil {
		return nil, err
	}

	req := &Request{
		requestLine: *reqLine,
	}

	return req, nil
}
