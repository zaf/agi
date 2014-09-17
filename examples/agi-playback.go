// Payback a file using AGI example in Go.
//
// Copyright (C) 2013 - 2014, Lefteris Zafiris <zaf.000@gmail.com>
// This program is free software, distributed under the terms of
// the BSD 3-Clause License. See the LICENSE file
// at the top of the source tree.

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/zaf/agi"
)

const debug = false

func main() {
	var file string
	var rep agi.Reply
	// Create a new AGI session and Parse the AGI environment.
	myAgi := agi.New()
	err := myAgi.Init(nil)
	if err != nil {
		log.Fatalf("Error Parsing AGI environment: %v\n", err)
	}
	if debug {
		// Print to stderr all AGI environment variables that are stored in myAgi.Env map.
		log.Println("AGI environment vars:")
		for key, value := range myAgi.Env {
			log.Printf("%-15s: %s\n", key, value)
		}
	}
	// Handle Hangup from the asterisk server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)
	go handleHangup(sigChan)

	// Check passed arguments. The filename of the file to be played back is supposed to be passed
	// as the first argument to the AGI script.
	if myAgi.Env["arg_1"] == "" {
		log.Println("No arguments passed, exiting...")
		goto HANGUP
	}
	file = myAgi.Env["arg_1"]
	// Chech channel status.
	rep, err = myAgi.ChannelStatus()
	if err != nil {
		log.Fatalf("AGI reply error: %v\n", err)
	}
	// Answer channel if not already answered.
	if rep.Res != 6 {
		rep, err = myAgi.Answer()
		if err != nil || rep.Res == -1 {
			log.Fatalf("Failed to answer channel: %v\n", err)
		}
	}
	// Playback file
	rep, err = myAgi.StreamFile(file, "1234567890*#")
	if err != nil {
		log.Fatalf("AGI reply error: %v\n", err)
	}
	if rep.Res == -1 {
		log.Printf("Error playing back file: %s\n", file)
	}

HANGUP:
	myAgi.Hangup()
}

func handleHangup(sch <-chan os.Signal) {
	signal := <-sch
	log.Printf("Received %v, exiting...\n", signal)
	os.Exit(1)
}
