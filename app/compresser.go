package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

func compressStream(stream ReadStream) (ReadStream, int, error) {
	data := make([]byte, 0, 1024);
	for {
		select {
		case d := <- stream.DataC():
			if d == nil {
				goto done
			}
			data = append(data, d...)
		case err := <- stream.ErrC():
			return nil, 0, err
		}		
	}

	done: var buf bytes.Buffer
	cw := gzip.NewWriter(&buf)
	_, err := cw.Write(data)
	if err != nil {
		return nil, 0, err
	}

	err = cw.Close()
	if err != nil {
		return nil, 0, err
	}

	compressd := buf.Bytes()
	fmt.Println("data ", compressd, data)

	return NewStaticReadStream(compressd), len(compressd), nil
}
