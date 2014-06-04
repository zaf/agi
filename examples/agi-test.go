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
	"github.com/zaf/agi"
	"log"
	"net"
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
	sess.Verbose("1.  Testing streamfile...")
	sess.StreamFile("beep", "")
	tests++
	if sess.Res == nil || sess.Res[0] != "0" {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}

	sess.Verbose("2.  Testing sendtext...")
	sess.SendText("Hello World")
	tests++
	if sess.Res == nil || sess.Res[0] != "0" {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}

	sess.Verbose("3.  Testing sendimage...")
	sess.SendImage("asterisk-image")
	tests++
	if sess.Res == nil || sess.Res[0] != "0" {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}

	sess.Verbose("4.  Testing saynumber...")
	sess.SayNumber(192837465, "")
	tests++
	if sess.Res == nil || sess.Res[0] != "0" {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}

	sess.Verbose("5.  Testing waitdtmf...")
	sess.WaitForDigit(3000)
	tests++
	if sess.Res == nil || sess.Res[0] != "0" {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}

	sess.Verbose("6.  Testing redord...")
	sess.RecordFile("/tmp/testagi", "alaw", "1234567890*#", 3000)
	tests++
	if sess.Res == nil || sess.Res[0] != "0" {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}

	sess.Verbose("7.  Testing record playback...")
	sess.StreamFile("/tmp/testagi", "")
	tests++
	if sess.Res == nil || sess.Res[0] != "0" {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}

	sess.Verbose("8.  Testing set variable...")
	sess.SetVariable("testagi", "foo")
	tests++
	if sess.Res == nil || sess.Res[0] != "1" {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}

	sess.Verbose("9.  Testing get full variable...")
	sess.GetFullVariable("testagi")
	tests++
	if sess.Res == nil || sess.Res[0] != "1" {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}

	sess.Verbose("10. Testing exec...")
	sess.Exec("Wait", "3")
	tests++
	if sess.Res == nil || sess.Res[0] != "0" {
		sess.Verbose("Failed.")
		fail++
	} else {
		pass++
	}

	sess.Verbose("================== Complete ======================")
	sess.Verbose(fmt.Sprintf("%d tests completed, %d passed, %d failed", tests, pass, fail))
	sess.Verbose("==================================================")
	return
}
