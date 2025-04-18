package main

import (
	"errors"
	"strings"
)

func parseRequestSegment(cr *ConnectionReader) ([]byte, error) {
	ret := make([]byte, 0, 1024)
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

func consumeByte(cr *ConnectionReader) error {
	_, err := cr.getByte()
	return err
}

func readUntil(cr *ConnectionReader, b byte) ([]byte, error) {
	ret := make([]byte, 0, 10)
	
	for {
		nb, err := cr.getByte()
		if err != nil {
			return []byte{}, err
		}
	
		if nb == b {
			return ret, nil
		}

		ret = append(ret, nb)
	}
}

func parseHeaders(cr *ConnectionReader) (map[string]string, error) {
	headers := make(map[string]string)

	for {
		b, err := cr.getByte()
		if err != nil {
			return nil, err
		}

		if b == byte('\r') {
			err := consumeByte(cr)
			if err != nil {
				return nil, err
			}

			return headers, nil
		}

		restName, err := readUntil(cr, byte(':'))
		if err != nil {
			return nil, err
		}

		err = consumeByte(cr)
		if err != nil {
			return nil, err
		}
		
		name := string(append([]byte{b}, restName...))

		value, err := readUntil(cr, byte('\r'))
		if err != nil {
			return nil, err
		}

		err = consumeByte(cr)
		if err != nil {
			return nil, err
		}

		headers[name] = string(value)
	}
}

func parseRequst(cr *ConnectionReader) (*Request, error) {
	reqLine, err := parseRequestLine(cr)
	if err != nil {
		return nil, err
	}

	headers, err := parseHeaders(cr)
	if err != nil {
		return nil, err
	}

	req := &Request{
		requestLine: *reqLine,
		headers: headers,
	}

	return req, nil
}
