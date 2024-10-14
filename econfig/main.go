package main

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

const (
	DefaultServerCertCommonName = "goblin"
	DefaultServerCertFile       = "/etc/rcos_global/tls_key/config_center/server/server.pem"
	DefaultServerCertKeyFile    = "/etc/rcos_global/tls_key/config_center/server/server-key.pem"
)

var (
	rootCmd = &cobra.Command{
		Use:   "econfig",
		Short: "econfig",
	}
	tlsCmd = &cobra.Command{
		Use:   "tls",
		Short: "generate tls files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateTLS(dnsList, ipList, caCertFile, caKeyFile, certFile, keyFile)
		},
	}
	subnetCmd = &cobra.Command{
		Use:   "subnet",
		Short: "generate a virtual subnet",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateSubnet(underlay, virtualIP, neighIPs, routeIPs)
		},
	}
)

func init() {

	tlsCmd.PersistentFlags().StringVar(&caCertFile, "ca-cert", "", "ca cert file path")
	tlsCmd.PersistentFlags().StringVar(&caKeyFile, "ca-key", "", "ca key file path")
	tlsCmd.PersistentFlags().StringVar(&certFile, "cert", DefaultServerCertFile, "cert file path")
	tlsCmd.PersistentFlags().StringVar(&keyFile, "key", DefaultServerCertKeyFile, "key file path")
	tlsCmd.PersistentFlags().StringSliceVar(&dnsList, "dns-list", []string{}, "dns domain written to tls certificate")
	tlsCmd.PersistentFlags().StringSliceVar(&ipList, "ip-list", []string{}, "ip address written to tls certificate")
	subnetCmd.PersistentFlags().StringVar(&underlay, "iface", "", "network interface for the vxlan device")
	subnetCmd.PersistentFlags().StringVar(&virtualIP, "ip", "", "vxlan device ip address")
	subnetCmd.PersistentFlags().StringSliceVar(&neighIPs, "vtep-ips", []string{}, "ip address of other vtep devices")
	subnetCmd.PersistentFlags().StringSliceVar(&routeIPs, "vxlan-ips", []string{}, "ip address of other vxlan devices")
	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(tlsCmd)
	rootCmd.AddCommand(subnetCmd)

}

func main() {

	if err := rootCmd.Execute(); err != nil {
		klog.Fatal("exit, %+v", err)
	}
}
