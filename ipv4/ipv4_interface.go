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
