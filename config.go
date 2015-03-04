package gobot

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
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
	// the given settings using Config.GenerateUrl
	Url string
}

// Reads the bot's config from file and converts it to a Config
// object for use by a Bot.
func GetConfig() Config {
	contents, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		// no config file, so we'll create a new one
		log.Println("No config file found, trying to use config-example instead...")
		// read from the example config
		contents, err = ioutil.ReadFile("./config-example.yaml")
		if err != nil {
			log.Fatal(err)
		}
		// and write it to the new config file
		err = ioutil.WriteFile("./config.yaml", contents, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	// and convert the YAML to a Config object
	var config Config
	err = yaml.Unmarshal(contents, &config)
	if err != nil {
		// someone probably messed with config-example.yaml
		// replace it with a correct one from the repo
		log.Fatal(err)
	}

	return config
}

// Generates a websocket URL to use for connecting, based on the given
// parameters.
// The websocket URL has the following format:
//   ws://server:port/showdown/websocket
func (conf *Config) GenerateUrl() {
	conf.Url = "ws://" + conf.Server + ":" + conf.Port +
		"/showdown/websocket"
}
