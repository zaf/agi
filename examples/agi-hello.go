// AGI 'Hello World' example in Go
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
	// Create a new AGI session and Parse the AGI environment.
	myAgi := agi.New()
	err := myAgi.Init(nil)
	if err != nil {
		log.Printf("Error Parsing AGI environment: %v\n", err)
		return
	}
	if debug {
		// Print to stderr all AGI environment variables that are stored in myAgi.Env map.
		log.Println("AGI environment vars:")
		for key, value := range myAgi.Env {
			log.Printf("%-15s: %s\n", key, value)
		}
	}
	// Print a message on the asterisk console using Verbose. AGI return values are stored in rep, an agi.Reply struct.
	rep, err := myAgi.Verbose("Hello World")
	if err != nil {
		log.Printf("AGI reply error: %v\n", err)
		return
	}
	if debug {
		// Print to stderr the AGI return values. In this case rep.Res is always 1 and rep.Dat is empty.
		log.Printf("AGI command returned: %d %s\n", rep.Res, rep.Dat)
	}
	// Hangup
	myAgi.Hangup()
}
