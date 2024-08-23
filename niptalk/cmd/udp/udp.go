package udp

import (
	"context"
	"github.com/QQGoblin/niptalk/pkg/udp"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

var (
	bPort int
)

func init() {
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
		return server(bPort)
	},
}

func broadcast(bPort int, message string) error {

	udpCli, err := udp.NewClient(bPort)
	if err != nil {
		return err
	}
	defer udpCli.Close()
	return udpCli.Broadcast(message)
}

func server(port int) error {

	srv := udp.NewSServer(port)

	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal)
	errChan := make(chan error)

	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Server(ctx, udp.PrintMessage); err != nil {
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
