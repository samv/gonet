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

// Other Consts
const TCP_INCOMING_BUFF_SZ = 200
const TCP_BASIC_HEADER_SZ = 20
const TCP_LISTEN_DEFAULT_QUEUE_SZ = 120
const TCP_RESEND_LIMIT = 12
const ACK_BUF_SZ = 100

// Window Sizing
const MAX_WINDOW_SZ = 65000
const MIN_WINDOW_SZ = 500

// TODO: set these properly based on the standard values

// Flags
const ( // TODO use iota
	TCP_FIN = 0x01
	TCP_SYN = 0x02
	TCP_RST = 0x04
	TCP_PSH = 0x08
	TCP_ACK = 0x10
	TCP_URG = 0x20
	TCP_ECE = 0x40
	TCP_CWR = 0x80
)
