package main

type TCP_Connection struct {
}

type TCP_Server_Manager struct {
	reader  *IP_Reader
	servers map[uint16](map[string](chan []byte)) // TODO make this global to TCP
}

func New_TCP_Server_Manager() (*TCP_Server_Manager, error) {
	nr, err := NewNetwork_Reader()
	if err != nil {
		return nil, err
	}

	ipr, err := nr.NewIP_Reader("*", TCP_PROTO)
	if err != nil {
		return nil, err
	}

	x := &TCP_Server_Manager{
		reader:  ipr,
		servers: make(map[uint16](map[string](chan []byte))),
	}

	return x, nil
}

// TODO: bind function

func (*TCP_Server_Manager) Listen(port uint16, ip string) error {
	// TODO bind
	return nil
}

func (*TCP_Server_Manager) Accept() (*TCP_Connection, error) {
	// TODO Listen for SYN
	// TODO Send back SYN + ACK
	// TODO Wait for ACK
	return nil, nil
}

func (*TCP_Connection) Recv() error {
	return nil
}

func (*TCP_Connection) Close() error {
	return nil
}
