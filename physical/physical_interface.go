package physical

import "io"

type Physical_IO_T struct {}
var Physical_IO *Physical_IO_T = &Physical_IO_T{} // TODO fix this

func (pio *Physical_IO_T) getInterface(ifindex Internal_Index) (Physical_Interface_IO) {
	if ifindex == Loopback_Internal_Index {
		return GlobalLoopbackIO
	} else { // TODO check properly
		return GlobalTapIO
	}
}

// blocking write to any interface
func (pio *Physical_IO_T) Write(ifindex Internal_Index, data []byte) (n int, err error) {
	face := pio.getInterface(ifindex)

	return face.Write(data)
}

// blocking read from any interface
func (pio *Physical_IO_T) Read() (Internal_Index, []byte, error) {
	select { // TODO support more interfaces
	case d := <-pio.getInterface(Loopback_Internal_Index).getInput():
		return Loopback_Internal_Index, d, nil
	case d := <-pio.getInterface(External_Internal_Index).getInput():
		return External_Internal_Index, d, nil
	}
}

type Physical_Interface_IO interface {
	io.Writer
	Read() ([]byte, error)
	io.Closer
	getInput() chan []byte
}