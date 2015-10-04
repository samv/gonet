package icmp

const HeaderMinSize = 8

const queueSize = 5000

type Type uint8

// Define common ICMP Types
const (
	EchoReply   Type = 0
	EchoRequest Type = 8
)
