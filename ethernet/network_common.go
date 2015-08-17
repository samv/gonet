package ethernet

type internal_index int

type MAC_Address struct {
	Data []byte
}

func (m *MAC_Address) Make() [8]byte {
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

func extract_dst(ethpacket []byte) *MAC_Address {
	return &MAC_Address{Data: ethpacket[:ETH_MAC_ADDR_SZ]}
}

func extract_src(ethpacket []byte) *MAC_Address {
	return &MAC_Address{Data: ethpacket[ETH_MAC_ADDR_SZ:2*ETH_MAC_ADDR_SZ]}
}

//type Ethernet_Addr struct {
//	MAC    *MAC_Address
//	intind InternalIndex
//}

const (
	// 768 = htons(ETH_P_ALL) = htons(3)
	// see http://ideone.com/2eunQu

	// 17 = AF_PACKET
	// see http://ideone.com/TGYlGc

	SOCK_DGRAM      = 2
	SOCK_RAW        = 3
	AF_PACKET       = 17
	HTONS_ETH_P_ALL = 768
	ETH_ALEN        = 6
)

type EtherType uint16

const (
	ETHERTYPE_IP = 0x0800
	ETHERTYPE_ARP = 0x0806
)

const (
	ETH_MAC_ADDR_SZ       = 6
	ETH_HEADER_SZ         = 14
	MAX_ETHERNET_FRAME_SZ = 1522 // for 1500 MTU + 22 bytes
	ETH_PROTOCOL_BUF_SZ   = 5000
)
