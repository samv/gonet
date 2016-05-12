package gonet

/*import (
	"github.com/hsheth2/gonet/ipv4"
	"github.com/hsheth2/gonet/tcp"
	"sync"
)

type fd int

var fdManager = new(struct {
	current fd
	sync.Mutex
})

func nextFd() fd {
	fdManager.Lock()
	defer fdManager.Unlock()
	fdManager.current++
	return fdManager.current
}

type Socket interface {
	Bind(port uint16, ip *ipv4.Address) error
	Listen(backlog int) error
	Accept() (s Socket, ip *ipv4.Address, port uint16, err error)
	Read(length int) ([]byte, error)
	Write([]byte) error
	Close() error
}

var _ Socket = &tcpSocket{}

type tcpSocket struct {
	fd fd

	server *tcp.Server
	client *tcp.TCB

	// server functions
	bind   func(s *tcpSocket, port uint16, ip *ipv4.Address) error
	listen func(s *tcpSocket, backlog int) error
	accept func(s *tcpSocket) (c *tcpSocket, ip *ipv4.Address, port uint16, err error)

	// client functions
	connect func(s *tcpSocket, port uint16, ip *ipv4.Address) error

	// common functions
	read  func(length int) ([]byte, error)
	write func(data []byte) error
	close func() error
}

func (s *tcpSocket) Bind(port uint16, ip *ipv4.Address) error {
	return s.bind(s, port, ip)
}

func (s *tcpSocket) Listen(backlog int) error {
	return s.listen(s, backlog)
}

func (s *tcpSocket) Accept() (c Socket, ip *ipv4.Address, port uint16, err error) {
	return s.accept(s)
}

func (s *tcpSocket) Connect(port uint16, ip *ipv4.Address) error {
	return s.connect(s, port, ip)
}

func (s *tcpSocket) Read(length int) ([]byte, error) {
	return s.read(length)
}

func (s *tcpSocket) Write(d []byte) error {
	return s.write(d)
}
func (s *tcpSocket) Close() error {
	return s.close()
}

func (s *tcpSocket) tcbReady() {
	s.bind = nil
	s.listen = nil
	s.accept = nil
	s.connect = nil
	s.read = s.client.Recv
	s.write = s.client.Send
	s.close = s.client.Close
}

func blankSocketTCP() *tcpSocket {
	return &tcpSocket{
		fd:      nextFd(),
		server:  nil, // TODO we don't need these
		client:  nil,
		bind:    nil,
		listen:  nil,
		accept:  nil,
		connect: nil,
		read:    nil,
		write:   nil,
		close:   nil,
	}
}

func SocketTCP() (Socket, error) {
	sock := blankSocketTCP()

	sock.bind = func(s *tcpSocket, port uint16, ip *ipv4.Address) error {
		server, err := tcp.NewServer()
		if err != nil {
			return err
		}
		s.server = server

		err = s.server.Bind(port, ip)
		if err != nil {
			return err
		}

		s.bind = nil
		s.connect = nil
		s.close = func() error {
			return s.server.Close()
		}

		s.listen = func(s *tcpSocket, backlog int) error {
			err := s.server.Listen(backlog)
			if err != nil {
				return err
			}

			s.listen = nil
			s.accept = func(s *tcpSocket) (*tcpSocket, *ipv4.Address, uint16, error) {
				q, ip, port, err := s.server.Accept()
				if err != nil {
					return nil, nil, 0, err
				}
				c := blankSocketTCP()
				c.client = q
				c.tcbReady()
				return c, ip, port, nil
			}
			return nil
		}
		return nil
	}

	sock.connect = func(s *tcpSocket, port uint16, ip *ipv4.Address) error {
		client, err := tcp.NewClient(port, ip)
		if err != nil {
			return err
		}

		tcb, err := client.Connect()
		if err != nil {
			return err
		}

		s.client = tcb
		s.tcbReady()
		return nil
	}

	return sock, nil
}

func socketUDP() (Socket, error) {
	// TODO finish UDP socket
	return nil, nil
}
*/
