package icmp

// The minimum size of an ICMP header
const HeaderMinSize = 8

const queueSize = 5000

// Type is the ICMP type
type Type uint8

// Define common ICMP Types
const (
	EchoReply   Type = 0
	EchoRequest Type = 8
)
