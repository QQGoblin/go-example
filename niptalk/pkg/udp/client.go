package udp

import (
	"context"
	"github.com/pkg/errors"
	"net"
)

type client struct {
	bPort int
	conn  *net.UDPConn
}

func NewClient(bPort int) (*client, error) {

	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0,
	})
	if err != nil {
		return nil, err
	}

	return &client{
		bPort: bPort,
		conn:  conn,
	}, err

}

func (c *client) Close() {

	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *client) Broadcast(message string) error {

	if _, err := c.conn.WriteToUDP([]byte(message), &net.UDPAddr{IP: net.IPv4bcast, Port: c.bPort}); err != nil {
		return err
	}
	return nil
}

func (c *client) WaitBroadcastReply(ctx context.Context, apply ApplyMessage) error {

	data := make([]byte, 1024)

	for {
		n, remoteAddr, err := c.conn.ReadFromUDP(data)

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
