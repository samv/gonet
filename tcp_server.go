package main

type TCP_Server struct {
}

type TCP_Server_Manager struct {
	reader  *IP_Reader
	servers map[uint16](map[string](chan []byte))
}

func New_TCP_Server_Manager() (*TCP_Server_Manager, error) {
	nr, err := NewNetwork_Reader()
	if err != nil {
		return nil, err
	}

	ipr, err := nr.NewIP_Reader("*", 6)
	if err != nil {
		return nil, err
	}

	x := &TCP_Server_Manager{
		reader:  ipr,
		servers: make(map[uint16](map[string](chan []byte))),
	}

	return x, nil
}

func (*TCP_Server_Manager) Listen(port uint16, ip string) (*TCP_Server, error) {
	// TODO Listen for SYN
	// TODO Send back SYN + ACK
	// TODO Wait for ACK
}

func (*TCP_Server) Close() {}
