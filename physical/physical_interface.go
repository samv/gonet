package physical

import (
	"io"
)

func init() {
	globalLoopbackIO = loInit()
	globalTapIO = tapInit()
}

func getInterface(ifindex InternalIndex) physicalInterface {
	if ifindex == LoopbackInternalIndex {
		return globalLoopbackIO
	}
	// TODO support an arbitrary number of InternalIndexes
	return globalTapIO
}

// Write allows blocking writes to any given interface
func Write(ifindex InternalIndex, data []byte) (n int, err error) {
	face := getInterface(ifindex)

	return face.Write(data)
}

// Read allows blocking reads from any given interface
func Read() ([]byte, InternalIndex, error) {
	//logs.Trace.Println("Reading from tap or lo")
	select { // TODO support more interfaces
	case d := <-getInterface(LoopbackInternalIndex).getInput():
		return d, LoopbackInternalIndex, nil
	case d := <-getInterface(ExternalInternalIndex).getInput():
		return d, ExternalInternalIndex, nil
	}
}

type physicalInterface interface {
	Write([]byte) (int, error)
	Read() ([]byte, error)
	io.Closer
	getInput() chan []byte
}
