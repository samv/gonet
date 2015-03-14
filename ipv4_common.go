package main

import (
//    "fmt"
)

func calcChecksum(head []byte, excludeChecksum bool) uint16 {
    totalSum := uint64(0)
    for ind, elem := range head {
        if (ind == 10 || ind == 11) && excludeChecksum { // Ignore the checksum in some situations
            continue
        }

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