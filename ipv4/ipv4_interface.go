package ipv4

import (
	"io"
)

func init() {
	initTypes()
	globalRoutingTable = initSourceIPTable()
	globalARPv4Table = initARPv4Table()
}

// Reader allows reading from a specific IP protocol and address
type Reader interface {
	ReadFrom() (*IPReadHeader, error)
	io.Closer
}

// Writer allows writing to a specific IP protocol and address
type Writer interface {
	WriteTo(data []byte) (int, error)
	io.Closer
}

// ReadWriter allows a bidirectional IP "connection": one that allows both reading and writing
type ReadWriter interface {
	ReadFrom() (*IPReadHeader, error)
	WriteTo(data []byte) (int, error)
	io.Closer
}

type ipv4ReadWriter struct {
	read  Reader
	write Writer
}

// NewReadWriter creates a ReadWriter given an IP Address and an IP protocol
func NewReadWriter(ip *Address, protocol uint8) (ReadWriter, error) {
	read, err := NewReader(ip, protocol)
	if err != nil {
		return nil, err
	}

	write, err := NewWriter(ip, protocol)
	if err != nil {
		read.Close()
		return nil, err
	}

	return &ipv4ReadWriter{
		read:  read,
		write: write,
	}, nil
}

func (irw *ipv4ReadWriter) ReadFrom() (*IPReadHeader, error) {
	return irw.ReadFrom()
}

func (irw *ipv4ReadWriter) WriteTo(data []byte) (int, error) {
	return irw.write.WriteTo(data)
}

func (irw *ipv4ReadWriter) Close() error {
	err1 := irw.write.Close()
	err2 := irw.read.Close()

	if err1 != nil {
		return err1
	}

	if err2 != nil {
		return err2
	}

	return nil
}
