/*
 * Specifies the commands that can be used by the bot. The commands are loaded
 * upon bot creation using `CreateBot` and can be called as described in the
 * documentation.
 *
 * Copyright 2015 (c) Ben Frengley (TalkTakesTime)
 */

package gobot

import (
	"strconv"
	"time"
)

// Loads the commands that are specified within the function. A command can
// then be called using `bot.commands["name"](msg)`.
//
// Messages will have arguments of the following form:
//    {user, args, command name[, timestamp]}
func (bot *Bot) LoadCommands() {
	bot.commands = map[string]func(Message){
		// say the current time for the server the bot is hosted on,
		// as well as the Unix timestamp (seconds since 00:00:00 1 Jan
		// 1970)
		"now": func(msg Message) {
			now := time.Now()
			bot.QueueMessage(now.String()+" -- "+
				strconv.FormatInt(now.Unix(), 10), msg.room)
		},

		// say "response" in the current room
		"test": func(msg Message) {
			bot.QueueMessage("response", msg.room)
		},

		// say the time the current room was joined as a Unix timestamp
		"getjoin": func(msg Message) {
			joinTime := bot.config.Rooms[toId(msg.args[1])]
			bot.QueueMessage(strconv.FormatInt(joinTime, 10), msg.room)
		},

		// say the value returned by applying `toId` to the arguments
		// following the command name
		"toid": func(msg Message) {
			bot.QueueMessage(toId(msg.args[1]), msg.room)
		},
	}

}
