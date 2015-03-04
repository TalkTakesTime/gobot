package gobot

import (
	"golang.org/x/net/websocket"
	"log"
	"time"
)

type Bot struct {
	config Config

	ws *websocket.Conn

	inQueue  chan string
	outQueue chan string
}

// Receives messages from PS and queues them up to be handled. Use as a
// goroutine or otherwise it will loop infinitely.
func (bot *Bot) Receive() {
	var msg string
	for {
		msg = ""
		if err := websocket.Message.Receive(bot.ws, &msg); err != nil {
			log.Fatal(err)
		}
		log.Printf("\nReceived: %s.\n", msg)
		bot.inQueue <- msg
	}
}

// Sends a message and waits 0.5s to allow the PS chat queue to empty
func (bot *Bot) SendMessage(msg string) {
	websocket.Message.Send(bot.ws, msg)
	time.Sleep(500 * time.Millisecond)
}

// Starts and runs the bot indefinitely, connecting to PS! and delegating
// tasks to other functions.
func (bot *Bot) Start() {
	log.Printf("\nConnecting to %s\n\n", bot.config.Url)

	var err error
	bot.ws, err = websocket.Dial(bot.config.Url, "", "https://play.pokemonshowdown.com")
	if err != nil {
		log.Fatal(err)
	}
	defer bot.ws.Close()

	go bot.Receive()

	for {
		select {
		case msg := <-bot.inQueue:
			log.Printf("\nUnqueued: %s.\n", msg)
			// TODO: add message handling
		default:
			// do nothing
		}
	}
}

func CreateBot(conf Config) Bot {
	return Bot{
		config:   conf,
		inQueue:  make(chan string, 10),
		outQueue: make(chan string, 10),
	}
}
