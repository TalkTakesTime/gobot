package gobot

import (
	"github.com/TalkTakesTime/hookserve/hookserve" // credits to phayes for the original
	"github.com/tonnerre/golang-pretty"
	"time"
)

func (bot *Bot) CreateHook() {
	bot.hookServer = hookserve.NewServer()
	bot.hookServer.Port = bot.config.HookPort
	bot.hookServer.Secret = bot.config.HookSecret
	bot.hookServer.GoListenAndServe()

	go bot.ListenForHooks()
}

func (bot *Bot) ListenForHooks() {
	for {
		select {
		case event := <-bot.hookServer.Events:
			pretty.Log(event)
			switch event.Type {
			case "push":
				bot.HandlePushHook(event)
			case "pull_request":
				bot.HandlePullHook(event)
			default:
				// do nothing for now
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (bot *Bot) HandlePushHook(event hookserve.Event) {
	msg := "**" + event.Owner + "/" + event.Repo + "** -- " + event.By +
		" pushed a new commit to " + event.Branch + ": " + event.Message +
		" (" + event.URL + ")"
	bot.QueueMessage(msg, "techcode")
}

func (bot *Bot) HandlePullHook(event hookserve.Event) {
	msg := "**" + event.Owner + "/" + event.Repo + "** -- " + event.By +
		" opened a new pull request: " + event.Message +
		" (" + event.URL + ")"
	bot.QueueMessage(msg, "techcode")
}
