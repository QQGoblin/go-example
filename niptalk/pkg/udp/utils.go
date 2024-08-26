package udp

import (
	"net"
)

func SendTo(dst *net.UDPAddr, b []byte) error {

	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0,
	})
	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = conn.WriteToUDP(b, dst)
	return err
}
