package udp

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type server struct {
	ListenAddr net.UDPAddr
}

func NewBDServer(port int) server {

	return server{
		ListenAddr: net.UDPAddr{IP: net.IPv4bcast, Port: port},
	}

}

func (s *server) Server(ctx context.Context) error {

	listener, err := net.ListenUDP("udp", &s.ListenAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	data := make([]byte, 1024)

	for {
		n, remoteAddr, err := listener.ReadFromUDP(data)

		fmt.Printf("Receive UDP Broadcast %s -> %s: %s\n", remoteAddr.String(), s.ListenAddr.String(), string(data[:n]))
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

func Broadcast(port int, message string) error {

	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0,
	})
	if err != nil {
		return errors.Wrapf(err, "create udp connect failed")
	}
	defer conn.Close()

	if _, err = conn.WriteToUDP([]byte(message), &net.UDPAddr{IP: net.IPv4bcast, Port: port}); err != nil {
		return errors.Wrapf(err, "broadcast message failed")
	}
	return nil
}

func Receive(port int) error {

	srv := NewBDServer(port)

	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal)
	errChan := make(chan error)

	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Server(ctx); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-exit:
		cancel()
	case srvErr := <-errChan:
		return srvErr
	}

	return nil
}
