// Pokemon Showdown! bot in Go, written by TalkTakesTime

package main

import "github.com/TalkTakesTime/gobot"

func main() {
	config := gobot.GetConfig()
	config.GenerateUrl()

	psBot := gobot.CreateBot(config)
	psBot.Start()
}
