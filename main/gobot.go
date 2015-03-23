/*
 * Pokemon Showdown! bot in Go, written by TalkTakesTime
 *
 * Usage:
 *   go build [-o output]
 *   ./main [-log=filename] (or ./output, if -o was used)
 *
 * Copyright 2015 (c) Ben Frengley (TalkTakesTime)
 */

package main

import (
	"flag"
	"github.com/TalkTakesTime/gobot"
	"log"
	"os"
)

var (
	logFile = flag.String("log", "", "the file to store the output logs in")
)

func main() {
	flag.Parse()

	// if a logfile is given, send output there instead of stdout
	if *logFile != "" {
		file := ChangeLogFile(*logFile)
		defer file.Close()
	}

	config := gobot.GetConfig()
	config.GenerateURL()

	psBot := gobot.CreateBot(config)
	psBot.Start()
}

// Changes logging from stdout to the given file. If the file doesn't
// exist, it is created. Returns a pointer to the file to allow defer
// to be used to close it at the end of the main function.
func ChangeLogFile(filename string) *os.File {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE,
		0666)
	if err != nil {
		log.Fatalf("could not use file %s for output", filename)
	}

	log.SetOutput(file)
	log.Println(">>> BEGIN LOGGING <<<")
	return file
}
