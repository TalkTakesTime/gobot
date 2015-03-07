/*
 * The generic functions of a bot can be found in this file. For chat handling,
 * see chathandler.go, and for battle handling, see battlehandler.go. Config
 * information can be found in config.go and the bot can be run from
 * main/main.go
 *
 * Copyright 2013 (c) Ben Frengley (TalkTakesTime)
 */

package gobot

import (
	"github.com/tonnerre/golang-pretty"
	"golang.org/x/net/websocket"
	"log"
	"strings"
	"time"
)

type Bot struct {
	// a Config struct representing the settings for the bot to use when it
	// runs. A config can be loaded from file using `gobot.GetConfig()`.
	// see config.go and main/main.go for more information
	config Config

	// the websocket connection for the bot to use to communicate with the
	// server. It is created in `Bot.Start()` so there is no need to generate
	// one yourself
	ws *websocket.Conn

	// queues to store messages while they wait to be processed or sent.
	// Created by `CreateBot` with a default capacity of 10, to allow
	// delayed processing and asynchronicity.
	inQueue  chan string
	outQueue chan string

	// a map of commands, mapping the command name to a handler function. If
	// a message is received that starts with the command character
	// immediately followed by a word that matches a command name, it will
	// execute the handler on the given message.
	commands map[string]func(Message)
}

// Simple error handler. Will probably improve it at some point.
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Receives messages from PS and queues them up to be handled. Use as a
// goroutine or otherwise it will loop infinitely.
func (bot *Bot) Receive() {
	var msg string
	for {
		msg = ""
		err := websocket.Message.Receive(bot.ws, &msg)
		checkError(err)

		log.Printf("\nReceived: %s.\n", msg)
		bot.inQueue <- msg
	}
}

// Causes the bot to join the given room and record when it joined in
// bot.config.Rooms
func (bot *Bot) JoinRoom(room string) {
	// track when the bot joined the room
	bot.config.Rooms[room] = time.Now().Unix()
	bot.QueueMessage("/join "+room, "")
}

// Adds a message for the given room to the outgoing queue. If the message
// is a PM, the room should be of the form "user:name", and the message
// will automatically get sent as a PM, so there is no need to add "/pm user, "
// to the front.
func (bot *Bot) QueueMessage(text, room string) {
	var msgData string
	if strings.HasPrefix(room, "user:") {
		msgData = "|/pm " + room[strings.Index(room, "user:")+5:] +
			"," + text
	} else {
		msgData = room + "|" + text
	}

	bot.outQueue <- msgData
}

// Sends a queued message through the websocket connection
func (bot *Bot) SendMessage(msg string) {
	log.Printf("\nSent message: %s\n", msg)
	err := websocket.Message.Send(bot.ws, msg)
	checkError(err)
}

// Reads messages from the out queue and sends them to PS, one each 0.5s or so
// to avoid the chat queue at the PS end filling up and blocking more messages
func (bot *Bot) Send() {
	for {
		select {
		case msg := <-bot.outQueue:
			bot.SendMessage(msg)
			time.Sleep(500 * time.Millisecond)
		default:
			// do nothing
			// is this necessary?
			// allows non-blocking reading from channels so I'll keep it
		}
	}
}

// Begins the main loop of the bot, which keeps it running indefinitely (or
// until it crashes, since I haven't given it any form of crash handling yet.)
func (bot *Bot) MainLoop() {
	for {
		select {
		case rawMsg := <-bot.inQueue:
			// TODO: add proper message handling
			messages := bot.ParseRawMessage(rawMsg)
			for _, msg := range messages {
				pretty.Log(msg)
				bot.ParseMessage(msg)
			}
		default:
			// do nothing
		}
	}
}

// Connects to PS! and begins the bot running.
func (bot *Bot) Start() {
	log.Printf("\nConnecting to %s\n\n", bot.config.Url)

	var err error
	bot.ws, err = websocket.Dial(bot.config.Url, "",
		"https://play.pokemonshowdown.com")
	checkError(err)
	defer bot.ws.Close()

	go bot.Receive()
	go bot.Send()

	bot.MainLoop()
}

// Creates and returns a bot using the given configuration, loading the
// commands in commands.go
func CreateBot(conf Config) Bot {
	bot := Bot{
		config:   conf,
		inQueue:  make(chan string, 100),
		outQueue: make(chan string, 100),
		commands: make(map[string]func(Message)),
	}
	bot.LoadCommands()
	return bot
}
