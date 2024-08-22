package main

import (
	"flag"
	"github.com/QQGoblin/ebpf-dump/dump"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var (
	mapName string
	list    bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&mapName, "name", "n", "cilium_lb4_backends_v2", "指定 ebpf 表名称")
	rootCmd.PersistentFlags().BoolVarP(&list, "list", "l", false, "打印所有 ebpf 表名称")
	klog.InitFlags(nil)
	if err := flag.Set("logtostderr", "true"); err != nil {
		klog.Exitf("failed to set logtostderr flag: %v", err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "dpcilium",
	Short: "dump cilium ebpf map",
	RunE: func(cmd *cobra.Command, args []string) error {
		if list {
			return dump.ListMap()
		}
		return dump.DumpMap(mapName)
	},
}

func main() {

	if err := rootCmd.Execute(); err != nil {
		klog.Fatalf(err.Error())
	}

	klog.Flush()

}
