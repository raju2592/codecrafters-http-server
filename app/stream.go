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

type ManualReadStream struct {
	dataC chan []byte
	errC chan error
	closeC chan bool
}

func NewManualReadStream() *ManualReadStream {
	return &ManualReadStream{
		dataC: make(chan []byte, 10),
		errC: make(chan error, 1),
		closeC: make(chan bool, 1),
	}
}

func (ms * ManualReadStream) DataC() chan []byte {
	return ms.dataC;
}

func (ms * ManualReadStream) ErrC() chan error {
	return ms.errC;
}

func (ms *ManualReadStream) Close() {
	ms.closeC <- true
}

func (ms *ManualReadStream) SendData(data []byte) {
	go func () {
		if (data == nil) {
			close(ms.dataC)
		} else {
			ms.dataC <- data
		}
	}()
}

func (ms *ManualReadStream) SendError(err error) {
	go func () {
		ms.ErrC() <- err
	}()
}

type ReaderReadStream struct {
	// file *os.File
	rem int
	reader io.Reader
	dataC chan []byte
	errC chan error
	closeC chan bool
}

type FileReadStream struct {
	*ReaderReadStream
	file *os.File
}

func NewFileReadStream(filePath string) (*FileReadStream, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	readerStream := &ReaderReadStream {
		reader: file,
		dataC: make(chan []byte, 10),
		errC: make(chan error, 1),
		closeC: make(chan bool, 1),
	}

	fileStream := &FileReadStream{
		ReaderReadStream: readerStream,
		file: file,
	}

	go fileStream.start()
	return fileStream, nil
}

func (fs *FileReadStream) Close() {
	fs.ReaderReadStream.Close()
	fs.file.Close()
}

func NewReaderReadStream(reader io.Reader, n int) (*ReaderReadStream, error) {
	stream := &ReaderReadStream {
		rem: n,
		reader: reader,
		dataC: make(chan []byte, 10),
		errC: make(chan error, 1),
		closeC: make(chan bool, 1),
	}

	go stream.start()
	return stream, nil
}

func (s * ReaderReadStream) start() {
	for {
		bufSz := 1024;
		if s.rem > 0 {
			bufSz = min(bufSz, s.rem)
		}

		readBuf := make([]byte, bufSz)

		n, err := s.reader.Read(readBuf)

		dataChan := s.dataC
		var errChan chan error = nil

		var data []byte = nil

		if err == io.EOF {
			close(s.dataC)
			return
		}
		
		if err != nil {
			dataChan = nil
			errChan = s.errC
		} else {
			data = readBuf[:n]
		}

		select {
		case dataChan <- data:
			if s.rem > 0 {
				s.rem -= n
			}

			if s.rem == 0 {
				close(s.dataC)
				return
			}
		case errChan <- err:
			return
		case <- s.closeC:
			return
		}
	}
}

func (s *ReaderReadStream) DataC() chan []byte {
	return s.dataC;
}

func (s *ReaderReadStream) ErrC() chan error {
	return s.errC
}

func (s *ReaderReadStream) Close() {
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
