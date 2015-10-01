package ipv4

import (
	"time"
	//"github.com/hsheth2/logs"
)

// IPv4 Protocols
const (
	IPProtoICMP = 1
	IPProtoUDP  = 17
	IPProtoTCP  = 6
)

const (
	ipHeaderLength    = 20
	maxIPPacketLength = 65535
)

const ipReadBufferSize = 5000

const defaultTimeToLive = 64

// The MTU, or maximum transmission unit
const (
	//IPMTU = 1500
	IPMTU = 65535 // TODO change back MTU for non-localhost stuff
)

const (
	fragmentationTimeout        = time.Second * 5
	fragmentAssemblerBufferSize = 35
)
