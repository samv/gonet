package ipv4

import (
	"time"
	//"github.com/hsheth2/logs"
	"network/ipv4/ipv4tps"
)

const (
	IPProtoICMP = 1
	IPProtoUDP = 17
	IPProtoTCP = 6
)

const (
	ipHeaderLength = 20
	maxIPPacketLength = 65535
)

const ipReadBufferSize = 5000

const defaultTimeToLive = 64

const (
	//IPMTU = 1500
	IPMTU = 65535 // TODO change back MTU for non-localhost stuff
)

const (
	fragmentationTimeout = time.Second * 5
	fragmentAssemblerBufferSize = 35
)

func Checksum(data []byte) uint16 {
	////ch logs.Trace.Println(data)
	totalSum := uint64(0)
	for ind, elem := range data {
		if ind%2 == 0 {
			totalSum += (uint64(elem) << 8)
		} else {
			totalSum += uint64(elem)
		}
	}
	//fmt.Println("Checksum total: ", totalSum)

	for prefix := (totalSum >> 16); prefix != 0; prefix = (totalSum >> 16) {
		//        fmt.Println(prefix)
		//        fmt.Println(totalSum)
		//        fmt.Println(totalSum & 0xffff)
		totalSum = totalSum&0xffff + prefix
	}
	//fmt.Println("Checksum after carry: ", totalSum)

	carried := uint16(totalSum)

	flip := ^carried
	//fmt.Println("Checksum: ", flip)

	return flip
}

func calculateIPChecksum(header []byte) uint16 {
	header[10] = 0
	header[11] = 0
	////ch logs.Trace.Println("Compute IP Checksum")
	return Checksum(header)
}
func verifyIPChecksum(header []byte) bool {
	////ch logs.Trace.Println("Verify Checksum")
	return Checksum(header) == 0
}

func CalcTransportChecksum(header []byte, srcIP, dstIP *ipv4tps.IPAddress, headerLen uint16, proto uint8) uint16 {
	////ch logs.Trace.Println("Transport Checksum")
	ips := append(srcIP.IP, dstIP.IP...)
	return Checksum(append(append(ips, []byte{0, byte(proto), byte(headerLen >> 8), byte(headerLen)}...), header...))
}

func VerifyTransportChecksum(header []byte, srcIP, dstIP *ipv4tps.IPAddress, headerLen uint16, proto uint8) bool {
	// TODO: do TCP/UDP checksum verification
	return true
}
