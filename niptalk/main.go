package main

import (
	"github.com/QQGoblin/niptalk/cmd/ipproto"
	udpcmd "github.com/QQGoblin/niptalk/cmd/udp"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	"os"
)

func NewCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "niptalk",
		Short: "this is healp tools for connect on local network",
	}

	cmd.AddCommand(udpcmd.Command)
	cmd.AddCommand(ipproto.Command)
	return cmd

}

func main() {

	cmd := NewCommand()
	if err := cmd.Execute(); err != nil {
		klog.Errorf("daemon exit, %+v", err)
		os.Exit(1)
	}
}
