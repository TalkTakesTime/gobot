Pokemon Showdown! GoBot
=======================

Another [Pokémon Showdown!][1] bot, this time in Go. Written by TalkTakesTime
as a learning exercise.

  [1]: https://play.pokemonshowdown.com/

WARNING
-------

This bot is still very much in development, and will be in its alpha stages
for possible a long time. Don't expect fast progress.

Installation
------------

This bot runs on [Go][2], Google's open-source language, and was developed
for version 1.4.2, although it has not been tested on any other versions.

It requires the following packages to run:
 - `log` (for logging, surprisingly)
 - `golang.org/x/net/websocket` (for websockets)
 - `strconv` (for int -> string conversion)
 - `math/rand` (for generating the websocket URL)
 - `time` (for random seeds)
 - `fmt` (for other logging?)
Note that the latter three will probably be removed in later updates.

I will assume that you know how to clone a Git repository or otherwise obtain
the source code. To install the dependencies, navigate to the directory
you downloaded this repository to, and run

    go get ./...

To build and start the bot, run

    go build
    ./GoBot

From there, you're on your own!

  [2]: http://golang.org/

License
-------

GoBot is distributed under the terms of the [MIT License][3].

 [3]: https://github.com/TalkTakesTime/GoBot/LICENSE
