// Payback a file using AGI example in Go
//
// Copyright (C) 2013 - 2014, Lefteris Zafiris <zaf.000@gmail.com>
// This program is free software, distributed under the terms of
// the BSD 3-Clause License. See the LICENSE file
// at the top of the source tree.

package main

import (
	"log"

	"github.com/zaf/agi"
)

const debug = false

func main() {
	//Start a new AGI session
	myAgi, err := agi.Init(nil)
	var rep agi.Reply
	var file string
	if err != nil {
		log.Printf("Error Parsing AGI environment: %v\n", err)
		return
	}
	if debug {
		//Print AGI environment
		log.Println("AGI environment vars:")
		for key, value := range myAgi.Env {
			log.Printf("%-15s: %s\n", key, value)
		}
	}
	// Check passed arguments
	if myAgi.Env["arg_1"] == "" {
		log.Println("No arguments passed, exiting...")
		goto HANGUP
	}
	file = myAgi.Env["arg_1"]
	// Chech channel status
	rep, err = myAgi.ChannelStatus()
	if err != nil {
		log.Printf("AGI reply error: %v\n", err)
		return
	}
	//Answer channel if not already answered
	if rep.Res != 6 {
		rep, err = myAgi.Answer()
		if err != nil || rep.Res == -1 {
			log.Printf("Failed to answer channel: %v\n", err)
			return
		}
	}
	// Playback file
	rep, err = myAgi.StreamFile(file, "1234567890*#")
	if err != nil {
		log.Printf("Error playing back file: %v\n", err)
	}

HANGUP:
	//Hangup
	myAgi.Hangup()
}
