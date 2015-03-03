// Pokemon Showdown! bot in Go, written by TalkTakesTime

package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"math/rand"
	"strconv"
	"time"
)

type Bot struct {
	nick   string
	pass   string
	server string
	port   int

	ws *websocket.Conn

	inQueue  chan string
	outQueue chan string
}

// TODO: move these into another file
var (
	nick   = ""
	pass   = ""
	server = "sim.smogon.com"
	port   = 8000
	url    = "ws://" + server + ":" + strconv.Itoa(port) + "/showdown/"
)

// Generates a random connection string for websocket connection to PS!
// This is apparently what PS wants, don't ask me why
func GenerateConStr() string {
	// generate a random number in the range 100 - 999 inclusive
	id := rand.Intn(900) + 100
	chars := "abcdefghijklmnopqrstuvwxyz0123456789_"
	str := ""

	// and a string of 8 random characters from the above list
	for i, l := 0, len(chars); i < 8; i++ {
		str += string(chars[rand.Intn(l)])
	}

	return url + strconv.Itoa(id) + "/" + str + "/websocket"
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
		fmt.Printf("Received: %s.\n", msg)
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
	// simple time-based seed to get new pseudo-random numbers each time
	rand.Seed(time.Now().UnixNano())

	connectionUrl := GenerateConStr()
	fmt.Printf("Connecting to %s\n\n", connectionUrl)
	var err error
	bot.ws, err = websocket.Dial(connectionUrl, "", "https://play.pokemonshowdown.com")
	if err != nil {
		log.Fatal(err)
	}
	defer bot.ws.Close()

	go bot.Receive()

	for {
		select {
		case msg := <-bot.inQueue:
			fmt.Printf("Unqueued: %s.\n", msg)
			// TODO: add message handling
		default:
			// do nothing
		}
	}
}

func main() {
	psBot := Bot{
		nick:     nick,
		pass:     pass,
		server:   server,
		port:     port,
		inQueue:  make(chan string, 10),
		outQueue: make(chan string, 10),
	}

	psBot.Start()
}
