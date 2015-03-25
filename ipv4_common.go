package main

import (
//    "fmt"
)

const (
// 768 = htons(ETH_P_ALL) = htons(3)
// see http://ideone.com/2eunQu

// 17 = AF_PACKET
// see http://ideone.com/TGYlGc
    MAX_IP_PACKET_LEN = 65535
    SOCK_DGRAM      = 2
    SOCK_RAW        = 3
    AF_PACKET       = 17
    HTONS_ETH_P_ALL = 768
    HTONS_ETH_P_IP  = 8
    ETH_ALEN        = 6
)

func checksum(head []byte) uint16 {
	totalSum := uint64(0)
	for ind, elem := range head {
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
		totalSum = uint64(totalSum&0xffff) + prefix
	}
	//fmt.Println("Checksum after carry: ", totalSum)

	carried := uint16(totalSum)

	flip := ^carried
	//fmt.Println("Checksum: ", flip)

	return flip
}

func calculateChecksum(header []byte) uint16 {
    header[10] = 0
    header[11] = 0
    return checksum(header)
}
func verifyChecksum(header []byte) uint16 {
    return checksum(header)
}