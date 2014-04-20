// Payback a file using AGI example in Go
//
// Copyright (C) 2013 - 2014, Lefteris Zafiris <zaf.000@gmail.com>
// This program is free software, distributed under the terms of
// the BSD 3-Clause License. See the LICENSE file
// at the top of the source tree.

package main

import (
	"github.com/zaf/agi"
	"log"
)

const debug = false

func main() {
	//Start a new AGI session
	myAgi, err := agi.Init(nil)
	defer myAgi.Destroy()
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
	err = myAgi.ChannelStatus()
	if err != nil {
		log.Printf("AGI reply error: %v\n", err)
		return
	}
	//Answer channel if not already answered
	if myAgi.Res[0] != "6" {
		err = myAgi.Answer()
		if err != nil || myAgi.Res[0] == "-1" {
			log.Printf("Failed to answer channel: %v\n", err)
			return
		}
	}
	// Playback file
	err = myAgi.StreamFile(file, "any")
	if err != nil || myAgi.Res[0] != "0" {
		log.Printf("Error playing back file: %v\n", err)
	}

HANGUP:
	//Hangup
	myAgi.Hangup("")
}
