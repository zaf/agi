// FastAGI 'Hello World' example in Go
//
// Copyright (C) 2013 - 2015, Lefteris Zafiris <zaf@fastmail.com>
// This program is free software, distributed under the terms of
// the BSD 3-Clause License. See the LICENSE file
// at the top of the source tree.

package main

import (
	"bufio"
	"log"
	"net"

	"github.com/zaf/agi"
)

const debug = false

func main() {
	// Create a listener on port 4573 and start a new goroutine for each connection.
	ln, err := net.Listen("tcp", ":4573")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go connHandle(conn)
	}
}

func connHandle(c net.Conn) {
	defer func() {
		c.Close()
		if err := recover(); err != nil {
			log.Println("Session terminated:", err)
		}
	}()
	// Create a new FastAGI session and Parse the AGI environment.
	myAgi := agi.New()
	rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
	err := myAgi.Init(rw)
	checkErr(err)
	if debug {
		// Print to stderr all AGI environment variables that are stored in myAgi.Env map.
		log.Println("AGI environment vars:")
		for key, value := range myAgi.Env {
			log.Printf("%-15s: %s\n", key, value)
		}
	}
	// Print a message on the asterisk console using Verbose. AGI return values are stored in rep, an agi.Reply struct.
	rep, err := myAgi.Verbose("Hello World")
	checkErr(err)
	if debug {
		// Print to stderr the AGI return values. In this case rep.Res is always 1 and rep.Dat is empty.
		log.Printf("AGI command returned: %d %s\n", rep.Res, rep.Dat)
	}
	return
}

//Check for AGI Protocol errors or hangups
func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}
