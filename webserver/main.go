package main

import (
	"context"
	"crypto/tls"
	kubex509 "k8s.io/apiserver/pkg/authentication/request/x509"
	"k8s.io/apiserver/pkg/endpoints/filters"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/klog/v2"
	"net"
	"net/http"
)

const (
	ClientCA = "client-ca.pem"
	CertFile = "server.crt"
	CertKey  = "server.key"
)

func main() {

	listener, _, err := options.CreateListener("tcp", ":9999", net.ListenConfig{})

	if err != nil {
		klog.Fatalf("create listener failed: %v", err)
	}

	cert, _ := dynamiccertificates.NewDynamicServingContentFromFiles("serving-cert", CertFile, CertKey)
	clientCA, _ := dynamiccertificates.NewDynamicCAContentFromFile("client-ca", ClientCA)
	s := server.SecureServingInfo{
		Listener:      listener,
		Cert:          cert,
		ClientCA:      clientCA,
		MinTLSVersion: tls.VersionTLS12,
	}

	verify, _ := clientCA.VerifyOptions()
	auth := kubex509.New(verify, kubex509.CommonNameUserConversion)

	pathRecorderMux := mux.NewPathRecorderMux("webserver")
	pathRecorderMux.Handle(
		"/hello",
		filters.WithAuthentication(handlerHello(), auth, authFailed(), nil),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		stopCh := server.SetupSignalHandler()
		<-stopCh
		cancel()
	}()

	stopC, err := s.Serve(pathRecorderMux, 0, ctx.Done())
	if err != nil {
		klog.Fatalf("start server failed: %v", err)
	}

	<-stopC
}

func handlerHello() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("hello"))

	}
}

func authFailed() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusForbidden)
		writer.Write([]byte("AuthenFailed"))

	}
}
