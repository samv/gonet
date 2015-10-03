package ipv4

// Checksum computes the 16-bit one's complement checksum of the given data. Done
// according to the procedure outlined in RFC 1071 (https://tools.ietf.org/html/rfc1071).
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

// CalcTransportChecksum calculates a checksum of a transport layer protocol (TCP or UDP)
func CalcTransportChecksum(header []byte, srcIP, dstIP *Address, headerLen uint16, proto uint8) uint16 {
	////ch logs.Trace.Println("Transport Checksum")
	ips := append(srcIP.IP, dstIP.IP...)
	return Checksum(append(append(ips, []byte{0, byte(proto), byte(headerLen >> 8), byte(headerLen)}...), header...))
}

// VerifyTransportChecksum verifies a given checksum from a transport layer protocol (TCP or UDP)
func VerifyTransportChecksum(header []byte, srcIP, dstIP *Address, headerLen uint16, proto uint8) bool {
	// TODO: do TCP/UDP checksum verification
	return true
}
