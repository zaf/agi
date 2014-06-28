// FastAGI 'Hello World' example in Go
//
// Copyright (C) 2013 - 2014, Lefteris Zafiris <zaf.000@gmail.com>
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

func main() {
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
	//Create a new AGI session
	rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
	myAgi, err := agi.Init(rw)
	defer func() {
		c.Close()
		myAgi.Destroy()
	}()
	if err != nil {
		log.Printf("Error Parsing AGI environment: %v\n", err)
		return
	}
	//Print a message on asterisk console
	err = myAgi.Verbose("Hello World")
	if err != nil {
		log.Printf("AGI reply error: %v\n", err)
		return
	}
	//Hangup
	myAgi.Hangup()
	return
}
