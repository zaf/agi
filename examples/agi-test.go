// A set of tests for AGI in Go
//
// Copyright (C) 2013 - 2014, Lefteris Zafiris <zaf.000@gmail.com>
// This program is free software, distributed under the terms of
// the BSD 3-Clause License. See the LICENSE file
// at the top of the source tree.
//
// Based on agi-test.agi from asterisk source tree.
// Can be used both as standalone AGI app or a FastAGI server
// if called with the flag '-spawn_fagi'

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/zaf/agi"
)

var listen = flag.Bool("spawn_fagi", false, "Spawn as a FastAGI server")

func main() {
	flag.Parse()
	if *listen {
		//If called as a FastAGI server
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
			go spawnAgi(conn)
		}
	} else {
		//If called as standalone AGI app
		spawnAgi(nil)
	}
}

func spawnAgi(c net.Conn) {
	var myAgi *agi.Session
	var err error
	if c != nil {
		//Create a new FastAGI session
		rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
		myAgi, err = agi.Init(rw)
		defer func() {
			c.Close()
			myAgi.Destroy()
		}()
	} else {
		//Create a new AGI session
		myAgi, err = agi.Init(nil)
		defer myAgi.Destroy()
	}
	if err != nil {
		log.Printf("Error Parsing AGI environment: %v\n", err)
		return
	}
	testAgi(myAgi)
	return
}

func testAgi(sess *agi.Session) {
	//Perform some tests
	var tests, pass, fail int
	var err error
	var r agi.Reply

	sess.Verbose("Testing answer...")
	r, err = sess.Answer()
	if err != nil || r.Res != 0 {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing channelstatus...")
	r, err = sess.ChannelStatus()
	if err != nil || r.Res != 6 {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing databaseput...")
	r, err = sess.DatabasePut("test", "my_key", "true")
	if err != nil || r.Res != 1 {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing databaseget...")
	r, err = sess.DatabaseGet("test", "my_key")
	if err != nil || r.Res != 1 || r.Dat != "true" {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing databasedel...")
	r, err = sess.DatabaseDel("test", "my_key")
	if err != nil || r.Res != 1 {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing databasedeltree...")
	r, err = sess.DatabaseDelTree("test")
	if err != nil || r.Res != 1 {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing streamfile...")
	r, err = sess.StreamFile("beep", "")
	if err != nil || r.Res != 0 {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing sendtext...")
	r, err = sess.SendText("Hello World")
	if err != nil || r.Res != 0 {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing sendimage...")
	r, err = sess.SendImage("asterisk-image")
	if err != nil || r.Res != 0 {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing saynumber...")
	r, err = sess.SayNumber(192837465, "")
	if err != nil || r.Res != 0 {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing wait for digit...")
	r, err = sess.WaitForDigit(3000)
	if err != nil || r.Res == -1 {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing record...")
	r, err = sess.RecordFile("/tmp/testagi", "alaw", "1234567890*#", 3000)
	if err != nil || r.Res == -1 {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing record playback...")
	r, err = sess.StreamFile("/tmp/testagi", "")
	if err != nil || r.Res != 0 {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing set variable...")
	r, err = sess.SetVariable("testagi", "foo")
	if err != nil || r.Res != 1 {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing get variable...")
	r, err = sess.GetVariable("testagi")
	if err != nil || r.Res != 1 || r.Dat != "foo" {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing get full variable...")
	r, err = sess.GetFullVariable("${testagi}")
	if err != nil || r.Res != 1 || r.Dat != "foo" {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing exec...")
	r, err = sess.Exec("Wait", "3")
	if err != nil || r.Res != 0 {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}
	tests++

	sess.Verbose("================== Complete ======================")
	sess.Verbose(fmt.Sprintf("%d tests completed, %d passed, %d failed", tests, pass, fail))
	sess.Verbose("==================================================")
	return
}
