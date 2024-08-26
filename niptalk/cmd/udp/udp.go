package udp

import (
	"context"
	"github.com/QQGoblin/niptalk/pkg/udp"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	bPort            int
	waitReply        bool
	replyTimeout     time.Duration
	replyByBroadcast bool
)

func init() {
	broadcastCMD.PersistentFlags().BoolVarP(&waitReply, "wait", "w", false, "wait server reply")
	broadcastCMD.PersistentFlags().DurationVar(&replyTimeout, "reply-timeout", 20*time.Second, "wait server reply timeout")
	serverCMD.PersistentFlags().BoolVar(&replyByBroadcast, "reply-broadcast", false, "reply client by broadcast")
	Command.PersistentFlags().IntVarP(&bPort, "broadcast-port", "p", 18789, "udp listen port")
	Command.AddCommand(broadcastCMD)
	Command.AddCommand(serverCMD)
}

var Command = &cobra.Command{
	Use:   "udp",
	Short: "send udp broadcast or run a simple udp server",
}

var broadcastCMD = &cobra.Command{
	Use:     "broadcast",
	Aliases: []string{"bd"},
	Short:   "send udp broadcast",
	RunE: func(cmd *cobra.Command, args []string) error {
		return broadcast(bPort, args[0])
	},
}

var serverCMD = &cobra.Command{
	Use:   "server",
	Short: "run a simple udp server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return server(bPort, replyByBroadcast)
	},
}

func broadcast(bPort int, message string) error {

	udpCli, err := udp.NewClient(bPort)
	if err != nil {
		return err
	}
	defer udpCli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), replyTimeout)
	exit := make(chan os.Signal)
	errChan := make(chan error)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	if waitReply {
		go func() {
			if replyErr := udpCli.WaitBroadcastReply(ctx, func(m *udp.BroadcastMessage) {
				if m.Remote == nil || m.Remote.String() == udpCli.SourceAddress() {
					return
				}
				klog.V(0).Infof("-> %s : %s", m.Remote.String(), m.Message)
			}); replyErr != nil {
				errChan <- replyErr
			}
		}()
	}

	if err = udpCli.Broadcast(message); err != nil {
		return err
	}

	if waitReply {
		select {
		case <-exit:
			cancel()
		case e := <-errChan:
			return e
		}
	}
	return nil
}

func server(port int, replyByBroadcast bool) error {

	srv := udp.NewSServer(port)

	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal)
	errChan := make(chan error)

	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	hostname, _ := os.Hostname()

	go func() {
		if err := srv.Server(ctx, func(m *udp.BroadcastMessage) {
			if m.Remote == nil {
				return
			}

			replyDst := m.Remote
			if replyByBroadcast {
				klog.V(0).Infof("-> broadcast : %s", m.Message)
				replyDst = &net.UDPAddr{
					IP:   net.IPv4bcast,
					Port: m.Remote.Port,
				}
			} else {
				klog.V(0).Infof("-> %s : %s", m.Remote.String(), m.Message)
			}

			if err := udp.SendTo(replyDst, []byte(hostname)); err != nil {
				klog.Warning("reply %s failed: %v", m.Remote, err)
			} else {
				klog.V(0).Infof("<- %s : reply my hostname", m.Remote.String())
			}
		}); err != nil {
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
