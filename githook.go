/*
 * This file handles GitHub webhooks. For more information on GitHub webhooks
 * go to https://developer.github.com/webhooks/, and for information on their
 * payloads, go to https://developer.github.com/v3/activity/events/types.
 * For information on the hookserve module which deals with receiving and
 * parsing the hooks, see https://github.com/TalkTakesTime/hookserve
 *
 * Copyright 2015 (c) Ben Frengley (TalkTakesTime)
 */

package gobot

import (
	"errors"
	"github.com/TalkTakesTime/hookserve/hookserve" // credits to phayes for the original
	"github.com/tonnerre/golang-pretty"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	GitIOBase   = "http://git.io/"
	GitIOCreate = "http://git.io/create"
)

var (
	ErrShortenURL = errors.New("could not shorten the URL")
)

// Error encountered when a URL can't be shortened using a URL shortener
// such as git.io or goo.gl
type ShortURLError struct {
	URL string // the URL trying to be shortened
	Err error
}

func (e *ShortURLError) Error() string {
	return e.URL + ": " + e.Err.Error()
}

// Generates the GitHub webhook receiver and starts a goroutine to deal
// with received events
func (bot *Bot) CreateHook() {
	bot.hookServer = hookserve.NewServer()
	bot.hookServer.Port = bot.config.HookPort
	bot.hookServer.Secret = bot.config.HookSecret
	bot.hookServer.GoListenAndServe()

	go bot.ListenForHooks()
}

// Listens for GitHub webhook events and delegates them to handlers, such as
// `HandlePushHook`. Currently only push and pull_request events are supported
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

// Sends messages to all relevant rooms updating them when a push event
// is received. Tells how many commits were pushed, and gives a description
// of each individual commit, as given in the commit message
func (bot *Bot) HandlePushHook(event hookserve.Event) {
	// we don't care about 0 commit pushes
	if event.Size == 0 {
		return
	}

	// attempt to shorten the URL using git.io
	shortURL, err := ShortenURL(event.URL)
	if err != nil {
		shortURL = event.URL
	}

	msg := "[" + event.Repo + "] **" + event.By + "** pushed " +
		strconv.Itoa(event.Size) + " commit(s) to " + event.Branch + " (" +
		shortURL + ")"
	// send to all githook rooms
	for _, r := range bot.config.HookRooms {
		bot.QueueMessage(msg, r)
	}

	// add messages for individual commits too
	for i := 0; i < event.Size; i++ {
		msgParts := strings.Split(event.Commits[i].Message, "\n")

		msg = event.Repo + "/" + event.Branch + " | " + event.Commits[i].SHA[:7] +
			" | **" + event.Commits[i].By + "**: " + msgParts[0]
		for _, r := range bot.config.HookRooms {
			bot.QueueMessage(msg, r)
		}
	}
}

// Sends messages to all relevant rooms updating them when a pull_request
// event is received. Still in beta
func (bot *Bot) HandlePullHook(event hookserve.Event) {
	// attempt to shorten the URL using git.io
	shortURL, err := ShortenURL(event.URL)
	if err != nil {
		shortURL = event.URL
	}

	msgParts := strings.Split(event.Message, "\n")

	msg := "[" + event.BaseRepo + "] **" + event.By + "** "
	switch event.Action {
	case "opened":
		msg += "opened a new "
	case "reopened":
		msg += "reopened a "
	case "closed":
		msg += "closed a "
	case "synchronize":
		msg += "synchronized a "
	}
	msg += "pull request: " + msgParts[0] + " __" +
		event.BaseBranch + "..." + event.Branch + "__ " + " (" + shortURL + ")"

	for _, r := range bot.config.HookRooms {
		bot.QueueMessage(msg, r)
	}
}

// Utility to shorten a URL using http://git.io/
// Returns an empty string and ErrShortenURL if something goes wrong,
// otherwise returns the shortened URL and nil
func ShortenURL(longURL string) (string, error) {
	response, err := http.PostForm(GitIOCreate, url.Values{
		"url": []string{longURL},
	})
	if err != nil {
		return "", &ShortURLError{longURL, ErrShortenURL}
	}

	extension, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return "", &ShortURLError{longURL, ErrShortenURL}
	}

	return GitIOBase + string(extension), nil
}
