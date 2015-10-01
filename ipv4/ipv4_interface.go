package ipv4

import (
	"io"
)

func init() {
	initTypes()
	GlobalSource_IP_Table = initSourceIPTable()
	globalARPv4Table = initARPv4Table()
}

type ipReadT interface {
	ReadFrom() (*IP_Read_Header, error)
}

type ipWriteT interface {
	WriteTo(data []byte) (int, error)
}

type ipCloseT interface {
	io.Closer
}

// Reader allows reading from a specific IP protocol and address
type Reader interface {
	ipReadT
	ipCloseT
}

// Writer allows writing to a specific IP protocol and address
type Writer interface {
	ipWriteT
	ipCloseT
}

// ReadWriter allows a bidirectional IP "connection": one that allows both reading and writing
type ReadWriter interface {
	ipReadT
	ipWriteT
	ipCloseT
}

type ipv4ReadWriter struct {
	read  Reader
	write Writer
}

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

func (irw *ipv4ReadWriter) ReadFrom() (*IP_Read_Header, error) {
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
