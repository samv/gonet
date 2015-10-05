package udp

import (
	"io"
	"network/ipv4"
)

type Reader interface {
	Read(len int) ([]byte, error)
	io.Closer
}

type Writer interface {
	Write(data []byte) (int, error)
	io.Closer
}

type ReadWriter interface {
	Read(len int) ([]byte, error)
	Write(data []byte) (int, error)
	io.Closer
}

type readWriter struct {
	Reader
	Writer
}

func NewReadWriter(local, remote Port, ip *ipv4.Address) (ReadWriter, error) {
	r, err := NewReader(local, ip)
	if err != nil {
		return nil, err
	}

	w, err := NewWriter(local, remote, ip)
	if err != nil {
		return nil, err
	}

	return &readWriter{r, w}, nil
}

func (rw *readWriter) Close() error {
	err1 := rw.Reader.Close()
	err2 := rw.Writer.Close()

	if err1 != nil {
		return err1
	}

	if err2 != nil {
		return err2
	}

	return nil
}
