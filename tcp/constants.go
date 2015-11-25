package tcp

// Finite State Machine
type fsmState int

const ( // TODO use iota
	fsmClosed      fsmState = 1
	fsmListen               = 2
	fsmSynSent              = 3
	fsmSynRcvd              = 4
	fsmEstablished          = 5
	fsmFinWait1             = 6
	fsmFinWait2             = 7
	fsmCloseWait            = 8
	fsmClosing              = 9
	fsmLastAck              = 10
	fsmTimeWait             = 11

	fsmNumStates = 11
)

// TCB Types
type tcbParentType int

const (
	serverParent tcbParentType = iota
	clientParent
)

// Buffer sizes
const (
	incomingBufferSize     = 200
	listenQueueSizeDefault = 120
	ackBufferSize          = 100
)

const basicHeaderSize = 20
const retransmissionLimit = 12

// Window Sizing TODO: set these properly based on the standard values
const (
	maxWindowSize = 65000
	minWindowSize = 500
)

// Flag type
type flag uint8

// Flags
const ( // TODO use iota
	flagFin flag = 0x01
	flagSyn      = 0x02
	flagRst      = 0x04
	flagPsh      = 0x08
	flagAck      = 0x10
	flagUrg      = 0x20
	flagEce      = 0x40
	flagCwr      = 0x80
)

const (
	minPort = uint16(32768)
	maxPort = uint16(61000)
)
