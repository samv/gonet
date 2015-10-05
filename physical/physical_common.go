package physical

// An InternalIndex allows simpler referencing to physical and virtual network interfaces.
// They are similar to Linux's ifindex
type InternalIndex int

// InternalIndexes are hardcoded
const (
	LoopbackInternalIndex InternalIndex = 1
	ExternalInternalIndex InternalIndex = 2
)

const rxQueueSize = 5000
const maxFrameSize = 1526
