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
	"strings"
)

var (
	LoginUrl = "https://play.pokemonshowdown.com/action.php"
	IdRegex  = regexp.MustCompile("[^a-z0-9]+")
)

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

// Log in to PS! under the given name and password. See PS! documentation
// if you want to understand exactly what is required for login.
// NOTE: LOGGING IN WITHOUT A PASSWORD DOES NOT WORK CURRENTLY
func (bot *Bot) LogIn(challstr Message) {
	var res *http.Response
	var err error

	if bot.config.Pass == "" {
		// TODO: fix this
		userId := toId(bot.config.Nick)
		res, err = http.Get(LoginUrl + "?act=getassertion&userid=" +
			userId + "&challengekeyid=" + challstr.args[0] + "&challenge=" +
			challstr.args[1])
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

	type LoginDetails struct {
		Assertion string
	}
	data := LoginDetails{}
	err = json.Unmarshal(body[1:], &data)
	checkError(err)

	bot.outQueue <- "|/trn " + bot.config.Nick + ",0," + data.Assertion
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
			msg.args = append(msgData[2:3], strings.Join(msgData[3:], "|"))
		case "c:":
			// we discard the timestamp and pretend it's just a |c| message
			msg.msgType = "c"
			msg.args = append(msgData[3:4], strings.Join(msgData[4:], "|"))
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
