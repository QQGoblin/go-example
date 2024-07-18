package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

var (
	certFile string
	keyFile  string
	addr     string
	tls      bool
)

func init() {

	flag.StringVar(&certFile, "cert", "tls.crt", "")
	flag.StringVar(&keyFile, "key", "tls.key", "")
	flag.StringVar(&addr, "addr", ":8181", "")
	flag.BoolVar(&tls, "tls", false, "")

}

func handler(w http.ResponseWriter, req *http.Request) {

	hostname, _ := os.Hostname()
	fmt.Fprintf(w, "Local Host %v\n", hostname)
	fmt.Fprintf(w, "Request Host %v\n", req.Host)
	fmt.Fprintf(w, "Request URL %v\n", req.URL.String())
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "Request Header %v : %v\n", name, h)
		}
	}
}

func main() {

	flag.Parse()

	http.HandleFunc("/", handler)
	if tls {
		http.ListenAndServeTLS(addr, certFile, keyFile, nil)
	} else {
		http.ListenAndServe(addr, nil)
	}

}
