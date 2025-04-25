package main

import (
	"errors"
	"fmt"
	"strconv"
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

func completeRequest(cr *ConnectionReader, req *Request, contentLength int) (*Request, error) {
	stream := NewManualReadStream()

	rem := contentLength

	go func() {
		fmt.Print("read loop ")
		for {
			bufSz := 1024;
			if rem > 0 {
				bufSz = min(bufSz, rem)
			}
			
			readBuf := make([]byte, 0, bufSz)

			for i := 0; i < bufSz; i++ {
				b, err := cr.getByte()

				if err != nil {
					fmt.Printf("Got an error")
					stream.SendError(err)
					req.end <- false
					return
				}
		
				readBuf = append(readBuf, b)
			}


			select {
			case stream.dataC <- readBuf:
				rem -= bufSz

				if (rem == 0) {
					close(stream.dataC)
					req.end <- true
					return
				}
	
			case <- stream.closeC:
				req.end <- false
				return
			}
		}
	}()

	req.body = stream
	return req, nil
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

	contentLength, _ := strconv.Atoi(headers["Content-Length"])

	req := &Request{
		requestLine: *reqLine,
		headers: headers,
		end: make(chan bool, 1),
	}

	if contentLength == 0 {
		req.end <- true
		return req, nil;
	}

	fmt.Println("Will parse body")
	return completeRequest(cr, req, contentLength);
}
