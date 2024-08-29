package ipproto

import (
	"github.com/QQGoblin/niptalk/pkg/ipproto"
	"github.com/spf13/cobra"
)

var (
	sip   string
	dip   string
	proto int
)

func init() {

	SendCMD.PersistentFlags().StringVar(&sip, "src", "", "send IP Packet with source ip")
	SendCMD.PersistentFlags().StringVar(&dip, "dst", "", "send IP Packet with destination ip")
	Command.PersistentFlags().IntVar(&proto, "proto", 233, "custom protocol num from 101 to 254")

	Command.AddCommand(SendCMD)
	Command.AddCommand(ReceiveCMD)
}

var Command = &cobra.Command{
	Use:   "ipproto",
	Short: "send message by raw socket",
}

var SendCMD = &cobra.Command{
	Use:   "send",
	Short: "send message by custom ipproto",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ipproto.SendIPPkt(sip, dip, proto, []byte(args[0]))
	},
}

var ReceiveCMD = &cobra.Command{
	Use:   "receive",
	Short: "receive message for custom ipproto",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ipproto.ReceiveIPPkt(proto)
	},
}
