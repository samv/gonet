package arp

import (
	"time"
)

const (
	operationRequest = 1
	operationReply   = 2
)

type hType uint16
type len uint8

const (
	ethernetHType hType = 1
	ethernetHLen  len   = 6
)

const requestTimeout = 500 * time.Millisecond
