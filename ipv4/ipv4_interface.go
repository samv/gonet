package ipv4

import (
	"io"
	"network/ipv4/ipv4tps"
)

type ipReadT interface {
	ReadFrom() (*IP_Read_Header, error)
}

type ipWriteT interface {
	WriteTo(data []byte) (int, error)
}

type ipCloseT interface {
	io.Closer
}

type Reader interface {
	ipReadT
	ipCloseT
}

type Writer interface {
	ipWriteT
	ipCloseT
}

type ReadWriter interface {
	ipReadT
	ipWriteT
	ipCloseT
}

type ipv4_read_writer struct {
	read  *ipReader
	write *ipv4_writer
}

func NewIPv4_RW(ip *ipv4tps.IPAddress, protocol uint8) (ReadWriter, error) {
	read, err := NewIP_Reader(ip, protocol)
	if err != nil {
		return nil, err
	}

	write, err := NewIP_Writer(ip, protocol)
	if err != nil {
		read.Close()
		return nil, err
	}

	return &ipv4_read_writer{
		read:  read,
		write: write,
	}, nil
}

func (irw *ipv4_read_writer) ReadFrom() (*IP_Read_Header, error) {
	return irw.ReadFrom()
}

func (irw *ipv4_read_writer) WriteTo(data []byte) (int, error) {
	return irw.write.WriteTo(data)
}

func (irw *ipv4_read_writer) Close() error {
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
