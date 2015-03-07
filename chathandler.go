/*
 * Handles incoming messages from PS!.
 *
 * See https://github.com/Zarel/Pokemon-Showdown/blob/master/protocol-doc.md
 * for more information about the messages PS! sends.
 *
 * Copyright 2015 (c) Ben Frengley (TalkTakesTime)
 */

package gobot

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	LoginUrl = "https://play.pokemonshowdown.com/action.php"
	IdRegex  = regexp.MustCompile("[^a-z0-9]+")
)

// A Message struct is a simpler way of dealing with the message data from
// the messages PS! sends, as it contains all the information that is needed
// in a much easier to access form than a raw message from PS!. They can be
// created by passing a raw message into `NewMessage`, along with the room the
// message was received in.
type Message struct {
	room    string
	raw     string
	msgType string
	args    []string
}

// Returns the id of the given string -- that is, the string translated
// to lower case, with all non-alphanumeric characters removed.
func toId(str string) string {
	return IdRegex.ReplaceAllString(strings.ToLower(str), "")
}

// Takes a single raw message from PS! and breaks it up into individual
// messages to respond to, returning a slice of Messages to be dealt with
// by the main parser.
func (bot *Bot) ParseRawMessage(rawMsg string) []Message {
	messages := make([]Message, 0)
	msgList := strings.Split(rawMsg, "\n")

	var room string
	for _, msg := range msgList {
		if strings.HasPrefix(msg, ">") {
			// a message of the form ">ROOMID"
			room = strings.TrimPrefix(msg, ">")
			continue
		}

		messages = append(messages, NewMessage(room, msg))
	}

	return messages
}

// Parses a non-raw message and determines what action to take in reponse.
// Currently most messages are ignored.
func (bot *Bot) ParseMessage(msg Message) {
	switch msg.msgType {
	case "challstr":
		bot.LogIn(msg)
	case "c", "c:", "chat", "pm":
		if strings.HasPrefix(msg.args[1], bot.config.CommandChar) {
			if msg.msgType == "c:" {
				// if it's a |c:| it comes with a timestamp, so we can
				// ignore it if the timestamp was before the bot joined
				// the room it was in
				joinTime, _ := strconv.ParseInt(msg.args[2], 10, 64)
				if joinTime < bot.config.Rooms[msg.room] {
					break
				}
			}
			bot.RunCommand(msg)
		}
	case "updateuser":
		if msg.args[1] == "1" { // the bot is logged in
			for room := range bot.config.Rooms {
				bot.JoinRoom(room)
			}
		}
	}
}

// Checks if the given command exists and executes the function it refers
// to if it does. Otherwise ignores the command.
func (bot *Bot) RunCommand(msg Message) {
	cmd := bot.GetCommand(msg.args[1])
	if cmd != "" && msg.args[0] != bot.config.Nick {
		if _, ok := bot.commands[cmd]; ok {
			msg.args[1] = strings.TrimSpace(strings.TrimPrefix(msg.args[1],
				bot.config.CommandChar+cmd))
			bot.commands[cmd](msg)
		}
	}
}

// Gets the command name from a message, if there is one. If not,
// returns the empty string.
func (bot *Bot) GetCommand(msg string) string {
	if !strings.HasPrefix(msg, bot.config.CommandChar) {
		return ""
	}

	firstSpace := strings.Index(msg, " ")
	if firstSpace != -1 {
		return msg[1:firstSpace]
	} else {
		return msg[1:]
	}
}

// Log in to PS! under the given name and password. See PS! documentation
// if you want to understand exactly what is required for login.
func (bot *Bot) LogIn(challstr Message) {
	var res *http.Response
	var err error

	// NOTE: This part does not match the PS! documentation.
	//
	// PS! documentation does not adequately describe the necessary process
	// for logging in without a password; instead of using a POST request,
	// a GET request should be used instead with the following fields:
	//   - act: getassertion
	//   - userid: the id version of the nick in config (use toId to get it)
	//   - challengekeyid: the first part of |challstr| (a single digit)
	//   - challenge: the second part of |challstr| (a string of characters)
	// See the below documentation for the difference in the HTTP response.
	if bot.config.Pass == "" {
		userId := toId(bot.config.Nick)
		res, err = http.Get(LoginUrl +
			"?act=getassertion&userid=" + userId +
			"&challengekeyid=" + challstr.args[0] +
			"&challenge=" + challstr.args[1])
	} else {
		res, err = http.PostForm(LoginUrl, url.Values{
			"act":            {"login"},
			"name":           {bot.config.Nick},
			"pass":           {bot.config.Pass},
			"challengekeyid": {challstr.args[0]},
			"challenge":      {challstr.args[1]},
		})
	}
	checkError(err)
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	checkError(err)

	// NOTE: This part does not match the PS! documentation.
	//
	// According to PS! documentation, the HTTP response should be a string
	// beginning with "]", followed by a JSON object, which contains a field
	// called `assertion`, which contains the message needed to complete login
	// using /trn. However, when logging in without a password, the HTTP
	// response is actually only the data that would normally be contained
	// in the `assertion` field. Because of this, the body of the HTTP response
	// can be used directly in /trn without needing to parse it as JSON.
	if bot.config.Pass == "" {
		bot.outQueue <- "|/trn " + bot.config.Nick + ",0," + string(body)
	} else {
		type LoginDetails struct {
			Assertion string
		}
		data := LoginDetails{}
		err = json.Unmarshal(body[1:], &data)
		checkError(err)

		bot.outQueue <- "|/trn " + bot.config.Nick + ",0," + data.Assertion
	}
}

// Determines the type of the message and its contents from its raw data.
// Certain messages get parts of them discarded as we're not interested
// in those parts (e.g., the timestamp in a |c:| message).
func (msg *Message) GetArgs() {
	if !strings.HasPrefix(msg.raw, "|") || strings.HasPrefix(msg.raw, "||") {
		// generic typeless message
		msg.msgType = ""
		msg.args = []string{msg.raw}
	} else {
		msgData := strings.Split(msg.raw, "|")
		msg.msgType = msgData[1] // first should be ""

		switch msg.msgType {
		case "c", "chat": // certain messages can contain | in them
			msg.args = append(msgData[2:3],
				strings.TrimSpace(strings.Join(msgData[3:], "|")))
		case "c:":
			// move the timestamp to the end
			msg.args = append(msgData[3:4],
				strings.TrimSpace(strings.Join(msgData[4:], "|")),
				msgData[2])
		case "pm":
			// PMs get treated differently so that they can be run through
			// commands the same way a normal chat message can be
			msg.room = "user:" + msgData[2]
			// discard the bot's name
			msg.args = append(msgData[2:3], msgData[4:]...)
		default:
			msg.args = msgData[2:]
		}
	}
}

// Creates a Message object for the given message with all required information
// for parsing it and responding later.
func NewMessage(room, raw string) Message {
	msg := Message{
		room: room,
		raw:  raw,
	}
	msg.GetArgs()

	return msg
}
