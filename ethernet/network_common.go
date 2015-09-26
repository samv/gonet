package ethernet

// The MACAddress type holds a 48-bit MAC Address
type MACAddress struct {
	Data []byte
}

// Converts a MAC address to an array of 8 bytes
func (m *MACAddress) make() [8]byte {
	// pad data to 8 bytes
	mac := m.Data
	for len(mac) < 8 {
		mac = append(mac, 0)
	}

	// convert
	var data [8]byte
	for i := 0; i < 8; i++ {
		data[i] = mac[i]
	}
	return data
}

// EtherType implements the ethernet protocol's EtherType (https://en.wikipedia.org/wiki/EtherType).
type EtherType uint16

// Major EtherTypes as constants
const (
	EtherTypeIP  EtherType = 0x0800
	EtherTypeARP EtherType = 0x0806
)

const (
	ethMACAddressSize = 6
	ethEtherTypeSize  = 2
	ethHeaderSize     = ethEtherTypeSize + 2*ethMACAddressSize
)

const maxEthernetFrameSize = 1522 // for 1500 MTU + 22 bytes

const ethProtocolBufferSize = 5000
