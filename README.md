Pokémon Showdown! GoBot
=======================

Another [Pokémon Showdown!][1] bot, this time in Go. Written by TalkTakesTime
as a learning exercise.

  [1]: https://play.pokemonshowdown.com/

WARNING
-------

This bot is still very much in development, and will be in its alpha stages
for possibly a long time. Don't expect fast progress.

IMPORTANT: login without a password is currently unsupported. It will be fixed
as soon as possible.

Installation
------------

This bot runs on [Go][2], Google's open-source language, and was developed
for version 1.4.2, although it has not been tested on any other versions.

It requires the following packages to run:
  - `encoding/json` -- for logging in
  - `github.com/tonnerre/golang-pretty` -- for pretty printing
  - `golang.org/x/net/websocket` -- for websockets
  - `gopkg.in/yaml.v2` -- for parsing the config
  - `io/ioutil` -- for reading files and http responses
  - `log` -- for logging
  - `net/http` -- for logging in
  - `net/url` -- for logging in
  - `regexp` -- for the PS standard toId function
  - `strings` -- for message parsing
  - `time` -- for sleeping etc

I will assume that you know how to clone a Git repository or otherwise obtain
the source code (hint: `go get github.com/TalkTakesTime/gobot` works). To
install the dependencies, navigate to the directory you downloaded this
repository to, and run

    go get .
    go build

To build and start the bot, run

    cd main
    go build
    ./main

From there, you're on your own!

  [2]: http://golang.org/

License
-------

GoBot is distributed under the terms of the [MIT License][3].

 [3]: https://github.com/TalkTakesTime/GoBot/LICENSE
