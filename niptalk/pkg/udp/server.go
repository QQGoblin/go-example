package udp

import (
	"context"
	"github.com/pkg/errors"
	"net"
)

type server struct {
	ListenAddr net.UDPAddr
}

func NewSServer(port int) server {

	return server{
		ListenAddr: net.UDPAddr{IP: net.IPv4bcast, Port: port},
	}

}

func (s *server) Server(ctx context.Context, apply ApplyMessage) error {

	listener, err := net.ListenUDP("udp", &s.ListenAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	data := make([]byte, 1024)

	for {
		n, remoteAddr, err := listener.ReadFromUDP(data)
		apply(&BroadcastMessage{
			Message: data[:n],
			remote:  remoteAddr,
		})

		select {
		case <-ctx.Done():
			return nil
		default:
			if err != nil {
				return errors.Wrapf(err, "read from message error")
			}
		}
	}

}
