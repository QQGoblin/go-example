package main

import (
	"github.com/QQGoblin/cadvisor/cadvisor"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	server, err := cadvisor.New("/metrics", 9035, []string{"/k8s.io"})
	if err != nil {
		klog.Fatal(err)
	}
	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	server.Run()
	<-exit
	server.Shutdown()
}
