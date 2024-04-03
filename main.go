package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/Richtermnd/anecdotebot/bot"
)

func main() {
	log.Println("start")
	bot.InitBot()
	go bot.Listen()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	log.Println("stop")
	bot.SaveSessions()
}
