package ipv4

import (
	"io"
	"network/ipv4/ipv4tps"
)

type ip_read_t interface {
	ReadFrom() (*IP_Read_Header, error)
}

type ip_write_t interface {
	WriteTo(data []byte) (int, error)
}

type ip_close_t interface {
	io.Closer
}

type IPv4_Reader interface {
	ip_read_t
	ip_close_t
}

type IPv4_Writer interface {
	ip_write_t
	ip_close_t
}

type IPv4_RW interface {
	ip_read_t
	ip_write_t
	ip_close_t
}

type ipv4_read_writer struct {
	read  *ipv4_reader
	write *ipv4_writer
}

func NewIPv4_RW(ip *ipv4tps.IPaddress, protocol uint8) (IPv4_RW, error) {
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
