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
	//Start a new AGI session
	myAgi, err := agi.Init(nil)
	defer myAgi.Destroy()
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
	//Print a message on asterisk console
	err = myAgi.Verbose("Hello World")
	if err != nil {
		log.Printf("AGI reply error: %v\n", err)
		return
	}
	if debug {
		//Print the response
		log.Printf("AGI command returned: %v\n", myAgi.Res)
	}
	//Hangup
	myAgi.Hangup()
}
