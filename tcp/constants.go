package tcp

// Finite State Machine
const ( // TODO use iota
	CLOSED      = 1
	LISTEN      = 2
	SYN_SENT    = 3
	SYN_RCVD    = 4
	ESTABLISHED = 5
	FIN_WAIT_1  = 6
	FIN_WAIT_2  = 7
	CLOSE_WAIT  = 8
	CLOSING     = 9
	LAST_ACK    = 10
	TIME_WAIT   = 11

	FSM_NUM_STATES = 11
)

// TCB Types
const (
	TCP_SERVER = iota
	TCP_CLIENT
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
