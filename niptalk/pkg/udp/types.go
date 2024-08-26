package udp

import (
	"net"
)

type BroadcastMessage struct {
	Message []byte
	Remote  *net.UDPAddr
}

type ApplyMessage func(message *BroadcastMessage)
