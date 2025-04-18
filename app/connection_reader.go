package main

import (
	"net"
)

type ConnectionReader struct {
	conn net.Conn
	data []byte
	buf []byte
	nextRead int
}

func NewConnectionReader(conn net.Conn) *ConnectionReader {
	return &ConnectionReader{
		conn: conn,
		nextRead: 0,
		buf: make([]byte, 1024),
		data: make([]byte, 0, 2048),
	}
}

func (cr *ConnectionReader) getNext() byte {
	val := cr.data[cr.nextRead]
	cr.nextRead++
	return val	
}

func (cr *ConnectionReader) getByte() (byte, error) {
	if cr.nextRead < len(cr.data) {
		return cr.getNext(), nil
	}

	n, err := cr.conn.Read(cr.buf)
	if err != nil {
		return 0, err
	}
  
	cr.data = append(cr.data, cr.buf[:n]...)
	return cr.getNext(), nil;
}
