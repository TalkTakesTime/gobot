/*
 * Specifies the commands that can be used by the bot. The commands are loaded
 * upon bot creation using `CreateBot` and can be called as described in the
 * documentation.
 *
 * Copyright 2015 (c) Ben Frengley (TalkTakesTime)
 */

package gobot

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	GitHubBaseURL = "https://github.com/"
)

var (
	GitSHARegex  = regexp.MustCompile("^[a-f0-9]{7,40}$")
	GitLineRegex = regexp.MustCompile("(#L?)?([0-9]+)")
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

		// gets the link to a git repository matching the criteria given.
		// TODO: if none can be found, attempts to find a close match
		//
		// Syntax: .git (user/repo|alias) (key:value){0,}
		// Valid keys are:
		//  - branch/b: the branch to look in. Defaults to master if excluded
		//  - commit/c: the SHA-1 hash of the commit. Should match the regex
		//      `[a-f0-9]{7,40}`
		//  - file/f: the file to link to. If in a directory other than the
		//      base directory, give the full path
		//  - line/l: the line number, in one of the forms number, #number, or
		//      #Lnumber
		//
		// Examples:
		// > .git server file:data/moves.js line:200
		// < https://github.com/Zarel/Pokemon-Showdown/blob/master/data/moves.js#L200
		// > .git bot
		// < https://github.com/TalkTakesTime/gobot
		// > .git server c:53e9c8b
		// < https://github.com/Zarel/Pokemon-Showdown/tree/53e9c8b
		// > .git client b:client-overhaul
		// < https://github.com/Zarel/Pokemon-Showdown-Client/tree/client-overhaul
		// > .git client b:client-overhaul f:README.md
		// < https://github.com/Zarel/Pokemon-Showdown-Client/blob/client-overhaul/README.md
		//
		// Notes:
		//  - if a key is present more than once the first instance will be
		//      used and all others ignored
		//  - key:value should not have spaces
		//  - most values are case sensitive
		"git": func(msg Message) {
			if toId(msg.args[1]) == "" || toId(msg.args[1]) == "help" {
				bot.QueueMessage(bot.config.CommandChar+
					"git (user/repo|alias) (key:value){0,}. More detailed"+
					" help can be found at http://git.io/hRt9", msg.room)
				return
			}

			details := map[string]string{
				"branch": "",
				"file":   "",
				"commit": "",
				"line":   "",
			}
			// options should be separated by a space
			args := strings.Split(msg.args[1], " ")

			repo, ok := bot.config.GitAliases[args[0]]
			if !ok { // if it's not a known alias take the literal value
				repo = args[0]
			}

			// test if the repository exists
			res, err := http.Get(GitHubBaseURL + repo)
			defer res.Body.Close()
			if err != nil || res.StatusCode != 200 {
				bot.QueueMessage("Unknown repository: "+repo, msg.room)
				return
			}

			if len(args) == 1 {
				// they just want the repo so exit here
				bot.QueueMessage(GitHubBaseURL+repo, msg.room)
				return
			}

			// otherwise, they gave some options
			unknownKey := []string{}
			view := "tree"
			for _, arg := range args[1:] {
				option := strings.Split(arg, ":")
				if len(option) == 1 || option[1] == "" {
					continue
				}
				if len(option) > 2 {
					option = append(option[:1], strings.Join(option[1:], ":"))
				}

				switch strings.ToLower(option[0]) {
				case "branch", "b":
					details["branch"] = option[1]
				case "file", "f":
					view = "blob"
					details["file"] = option[1]
				case "commit", "c":
					valid := GitSHARegex.MatchString(option[1])
					if !valid {
						continue
					}
					details["commit"] = option[1]
				case "line", "l":
					matches := GitLineRegex.FindStringSubmatch(option[1])
					if matches[0] == "" {
						continue
					}
					details["line"] = matches[2]
				default:
					// unknown key
					unknownKey = append(unknownKey, option[0])
				}
			}

			var response string
			if len(unknownKey) > 0 {
				response += "Unknown key"
				if len(unknownKey) > 1 {
					response += "s"
				}
				response += ": " + strings.Join(unknownKey, ", ") + ". "
			}
			response += GitHubBaseURL + repo + "/"
			// if a commit is given the branch is actually ignored
			if details["commit"] != "" {
				response += view + "/" + details["commit"]
				if details["file"] != "" {
					response += "/" + details["file"]
					if details["line"] != "" {
						response += "#L" + details["line"]
					}
				}
			} else {
				if details["branch"] != "" {
					response += view + "/" + details["branch"]
				} else if details["file"] != "" {
					response += view + "/master"
				}
				if details["file"] != "" {
					response += "/" + details["file"]
					if details["line"] != "" {
						response += "#L" + details["line"]
					}
				}
			}

			bot.QueueMessage(response, msg.room)
		},
	}
}
