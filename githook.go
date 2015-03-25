/*
 * This file handles GitHub webhooks. For more information on GitHub webhooks
 * go to https://developer.github.com/webhooks/, and for information on their
 * payloads, go to https://developer.github.com/v3/activity/events/types.
 * For information on the hookserve module which deals with receiving and
 * parsing the hooks, see https://github.com/TalkTakesTime/hookserve
 *
 * Note that the bot will panic if the port given in config.yaml is already
 * in use, so be careful.
 *
 * Copyright 2015 (c) Ben Frengley (TalkTakesTime)
 */

package gobot

import (
	"errors"
	"fmt"
	"github.com/TalkTakesTime/hookserve/hookserve" // credits to phayes for the original
	"github.com/tonnerre/golang-pretty"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	GitIOBase   = "http://git.io/"
	GitIOCreate = "http://git.io/create"

	// template for push messages
	// [repo] user pushed number new commits? to branch: URL
	PushTemplate = "[%s] %s pushed %s new commit%s to %s: %s"
	// template for commit messages
	// repo/branch SHA user: commit message
	CommitTemplate = "%s/%s %s %s: %s"
	// template for pull request messages
	// [repo] user action pull request #number: message (upstream...base) URL
	PullReqTemplate = "[%s] %s %s pull request #%s: %s (%s...%s) %s"
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

	plural := ""
	if event.Size > 1 {
		plural = "s"
	}

	msg := fmt.Sprintf(PushTemplate, FormatRepo(event.Repo),
		FormatName(event.By), FormatSize(event.Size), plural,
		FormatBranch(event.Branch), FormatURL(shortURL))
	// send to all githook rooms
	// for _, r := range bot.config.HookRooms {
	// 	bot.QueueMessage(msg, r)
	// }

	// add messages for individual commits too
	for i := 0; i < event.Size; i++ {
		msgParts := strings.Split(event.Commits[i].Message, "\n")

		msg += "<br />" + fmt.Sprintf(CommitTemplate, FormatRepo(event.Repo),
			FormatBranch(event.Branch), FormatSHA(event.Commits[i].SHA[:7]),
			FormatName(event.Commits[i].By), msgParts[0])
	}

	msg = "!htmlbox " + msg
	for _, r := range bot.config.HookRooms {
		bot.QueueMessage(msg, r)
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

	if event.Action == "synchronize" {
		event.Action = "synchronized"
	}

	msg := fmt.Sprintf(PullReqTemplate, FormatRepo(event.BaseRepo),
		FormatName(event.By), event.Action, event.Number, msgParts[0],
		FormatBranch(event.BaseBranch), FormatBranch(event.Branch),
		FormatURL(shortURL))

	for _, r := range bot.config.HookRooms {
		bot.QueueMessage("!htmlbox "+msg, r)
	}
}

// FormatRepo formats a repo name for !htmlbox using #FF00FF.
func FormatRepo(repo string) string {
	return fmt.Sprintf("<font color=\"#FF00FF\">%s</font>", html.EscapeString(repo))
}

// FormatBranch formats a branch for !htmlbox using #9C009C.
func FormatBranch(branch string) string {
	return fmt.Sprintf("<font color=\"#9C009C\">%s</font>", html.EscapeString(branch))
}

// FormatName formats a name for !htmlbox using #7F7F7F. Note that it uses a
// different colour to the IRC version due to PS' background colour.
func FormatName(name string) string {
	return fmt.Sprintf("<font color=\"#6F6F6F\">%s</font>", html.EscapeString(name))
}

// FormatURL formats a URL for !htmlbox.
func FormatURL(url string) string {
	return fmt.Sprintf("<a href=\"%s\">%s</a>", html.EscapeString(url), html.EscapeString(url))
}

// FormatSHA formats a commit SHA for !htmlbox using #7F7F7F.
func FormatSHA(sha string) string {
	return fmt.Sprintf("<font color=\"#7F7F7F\">%s</font>", html.EscapeString(sha))
}

// FormatSize formats an event size for !htmlbox using <strong>.
func FormatSize(size int) string {
	return fmt.Sprintf("<strong>%d</strong>", size)
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
