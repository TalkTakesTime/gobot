/*
 * The generic functions of a bot can be found in this file. For chat handling,
 * see chathandler.go, and for battle handling, see battlehandler.go. Config
 * information can be found in config.go and the bot can be run from
 * main/main.go
 *
 * Copyright 2015 (c) Ben Frengley (TalkTakesTime)
 */

package gobot

import (
	"github.com/TalkTakesTime/hookserve/hookserve"
	"github.com/gorilla/websocket"
	"github.com/tonnerre/golang-pretty"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	BufferSize = 4096
)

var PingTicker *time.Ticker

type Bot struct {
	// a Config struct representing the settings for the bot to use when it
	// runs. A config can be loaded from file using `gobot.GetConfig()`.
	// see config.go and main/gobot.go for more information
	config Config

	// the websocket connection for the bot to use to communicate with the
	// server. It is created in `Bot.Start()` so there is no need to generate
	// one yourself
	ws *websocket.Conn

	// queues to store messages while they wait to be processed or sent.
	// Created by `CreateBot` with a default capacity of 100, to allow
	// delayed processing and asynchronicity.
	inQueue  chan string
	outQueue chan string

	// a map of commands, mapping the command name to a handler function. If
	// a message is received that starts with the command character
	// immediately followed by a word that matches a command name, it will
	// execute the handler on the given message.
	commands map[string]func(Message)

	// the server that listens for github webooks
	hookServer *hookserve.Server
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
	for {
		msgType, msg, err := bot.ws.ReadMessage()
		checkError(err)

		if msgType != websocket.TextMessage {
			log.Println("unexpected message type:", msg)
			return
		}

		// log.Printf("\nReceived: %s.\n", msg)
		bot.inQueue <- string(msg)
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
	err := bot.ws.WriteMessage(websocket.TextMessage, []byte(msg))
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
		case <-PingTicker.C:
			err := bot.ws.WriteControl(websocket.PingMessage, []byte("ping"),
				time.Now().Add(10*time.Second))
			if err != nil {
				pretty.Log(err)
			}
		}
	}
}

// Begins the main loop of the bot, which keeps it running indefinitely (or
// until it crashes, since I haven't given it any form of crash handling yet.)
func (bot *Bot) MainLoop() {
	for {
		select {
		case rawMsg := <-bot.inQueue:
			messages := bot.ParseRawMessage(rawMsg)
			for _, msg := range messages {
				pretty.Log(msg)
				bot.ParseMessage(msg)
			}
		}
	}
}

// Connects to PS! and begins the bot running.
func (bot *Bot) Start() {
	conn, err := net.Dial("tcp", bot.config.Server+":"+bot.config.Port)
	checkError(err)

	log.Printf("\nConnecting to %s\n\n", bot.config.URL.String())

	var res *http.Response
	bot.ws, res, err = websocket.NewClient(conn, bot.config.URL, http.Header{
		"Origin": []string{"https://play.pokemonshowdown.com"},
	}, BufferSize, BufferSize)
	if err != nil {
		pretty.Logf("%s: %#v\n", err.Error(), res)
		log.Fatal()
	}

	PingTicker = time.NewTicker(time.Minute)
	bot.ws.SetPongHandler(func(s string) error {
		pretty.Log("received pong:", s)
		return nil
	})

	defer res.Body.Close()
	defer bot.ws.Close()
	defer PingTicker.Stop()

	go bot.Receive()
	go bot.Send()
	if bot.config.EnableHooks {
		bot.CreateHook() // creates and starts the server for github webhooks
	}

	bot.MainLoop()
}

// Creates and returns a bot using the given configuration, loading the
// commands in commands.go
func CreateBot(conf Config) Bot {
	bot := Bot{
		config:   conf,
		inQueue:  make(chan string, 100),
		outQueue: make(chan string, 100),
		// this currently isn't actually needed I don't think, so I should
		// probably remove it
		commands: make(map[string]func(Message)),
	}
	bot.LoadCommands()
	return bot
}
