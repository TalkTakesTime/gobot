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
	// delayed processing and asynchronicity
	inQueue  chan string
	outQueue chan string
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

func (bot *Bot) SendMessage(msg string) {
	log.Printf("\nSent message: %s\n", msg)
	err := websocket.Message.Send(bot.ws, msg)
	checkError(err)
}

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

// Starts and runs the bot indefinitely, connecting to PS! and delegating
// tasks to other functions.
func (bot *Bot) Start() {
	log.Printf("\nConnecting to %s\n\n", bot.config.Url)

	var err error
	bot.ws, err = websocket.Dial(bot.config.Url, "",
		"https://play.pokemonshowdown.com")
	checkError(err)
	defer bot.ws.Close()

	go bot.Receive()
	go bot.Send()

	for {
		select {
		case rawMsg := <-bot.inQueue:
			// TODO: add proper message handling
			messages := bot.ParseRawMessage(rawMsg)
			for _, msg := range messages {
				pretty.Log(msg)
				if msg.msgType == "challstr" {
					bot.LogIn(msg)
				}
			}
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
