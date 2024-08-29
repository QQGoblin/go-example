package main

import (
	"fmt"
	"github.com/QQGoblin/niptalk/cmd/ipproto"
	"github.com/QQGoblin/niptalk/cmd/udp"
	"github.com/spf13/cobra"
	"os"
)

func NewCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "niptalk",
		Short: "this is healp tools for connect on local network",
	}

	cmd.AddCommand(udp.Command)
	cmd.AddCommand(ipproto.Command)
	return cmd

}

func main() {

	cmd := NewCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Printf("exit with error, %+v\n", err)
		os.Exit(1)
	}
}
