/*
 * Specifies the commands that can be used by the bot. The commands are loaded
 * upon bot creation using `CreateBot` and can be called as described in the
 * documentation.
 *
 * Copyright 2015 (c) Ben Frengley (TalkTakesTime)
 */

package gobot

// Loads the commands that are specified within the function. A command can
// then be called using `bot.commands["name"](msg)`.
func (bot *Bot) LoadCommands() {
	bot.commands["test"] = func(msg Message) {
		bot.QueueMessage("response", msg.room)
	}
}
