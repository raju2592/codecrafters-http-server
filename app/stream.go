package main

import (
	"io"
	"os"
)

type ReadStream interface {
	DataC() chan []byte
	ErrC() chan error
	Close()
}

type FileReadStream struct {
	file *os.File
	buf []byte
	dataC chan []byte
	errC chan error
	closeC chan bool
}

func NewFileReadStream(filePath string) (*FileReadStream, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	stream := &FileReadStream {
		file: file,
		buf: make([]byte, 1024),
		dataC: make(chan []byte, 10),
		errC: make(chan error, 1),
		closeC: make(chan bool, 1),
	}

	go stream.start()
	return stream, nil
}

func (s * FileReadStream) start() {
	for {
		n, err := s.file.Read(s.buf)

		dataChan := s.dataC
		var errChan chan error = nil

		var data []byte = nil

		if err == io.EOF {
			close(s.dataC)
			break
		}
		
		if err != nil {
			dataChan = nil
			errChan = s.errC
		} else {
			data = s.buf[:n]
		}

		select {
		case dataChan <- data:
			continue
		case errChan <- err:
			break
		case <- s.closeC:
			break
		}
	}
}

func (s *FileReadStream) DataC() chan []byte {
	return s.dataC;
}

func (s *FileReadStream) ErrC() chan error {
	return s.errC
}

func (s *FileReadStream) Close() {
	s.closeC <- true
}


type StaticReadStream struct {
	dataC chan []byte
}

func NewStaticReadStream(data []byte) *StaticReadStream {
	dataC := make(chan []byte, 1)
	dataC <- data

	return &StaticReadStream{
		dataC: dataC,		
	}
}

func (s *StaticReadStream) DataC() chan []byte {
	return s.dataC;
}

func (s *StaticReadStream) ErrC() chan error {
	return nil
}

func (s *StaticReadStream) Close() {
}
