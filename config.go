/*
 * Configuration management for the bot.
 *
 * The default config can be found in main/config-example.yaml and should
 * be copied into main/config.yaml, then edited to meet requirements.
 *
 * Copyright 2015 (c) Ben Frengley (TalkTakesTime)
 */

package gobot

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/url"
)

type Config struct {
	/**** General config ****/
	// The nickname to use on PS!. Limited to 16 characters
	Nick string
	// The password associated with the given nick. Blank if
	// the nick is unregistered
	Pass string
	// The server to connect to. PS! main's server is sim.smogon.com
	Server string
	// The port the given server uses. Default should be 8000
	Port string
	// The websocket URL to connect to. Generate automatically from
	// the given settings using Config.GenerateURL
	URL *url.URL
	// The character that indicates that the message received is for the bot
	// to respond to. TODO: add validation for command char
	CommandChar string
	// The rooms the bot is in. Initially loaded from the config file, and
	// updated whenever the bot joins a room.
	Rooms map[string]int64

	/**** Git config ****/
	// Whether or not the bot should listen for webhooks
	EnableHooks bool
	// The port that the bot should listen on for incoming GitHub webhooks
	HookPort int
	// The secret given during the creation of the webhook. Must match the
	// secret on GitHub
	HookSecret string
	// A list of rooms to update when a webhook is received
	HookRooms []string
	// Aliases for .git, of the form alias: user/repo
	GitAliases map[string]string
}

// Reads the bot's config from file and converts it to a Config
// object for use by a Bot.
func GetConfig() Config {
	contents, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		// no config file, so we'll create a new one
		log.Println("No config file found, trying to use config-example" +
			" instead...")
		// read from the example config
		contents, err = ioutil.ReadFile("./config-example.yaml")
		checkError(err)

		// and write it to the new config file
		err = ioutil.WriteFile("./config.yaml", contents, 0644)
		checkError(err)
	}

	// and convert the YAML to a Config object
	var config Config
	err = yaml.Unmarshal(contents, &config)
	checkError(err)

	return config
}

// Generates a websocket URL to use for connecting, based on the given
// parameters.
// The websocket URL has the following format:
//   ws://server:port/showdown/websocket
func (conf *Config) GenerateURL() {
	var err error
	conf.URL, err = url.Parse("ws://" + conf.Server + ":" + conf.Port +
		"/showdown/websocket")
	checkError(err)
}
