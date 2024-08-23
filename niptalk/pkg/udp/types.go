package udp

import (
	"fmt"
	"net"
)

type BroadcastMessage struct {
	Message []byte
	remote  *net.UDPAddr
}

type ApplyMessage func(message *BroadcastMessage)

func PrintMessage(m *BroadcastMessage) {
	fmt.Printf("%s -> %s\n", m.remote.String(), m.Message)
}
