// A set of tests for AGI in Go
//
// Copyright (C) 2013 - 2014, Lefteris Zafiris <zaf.000@gmail.com>
// This program is free software, distributed under the terms of
// the BSD 3-Clause License. See the LICENSE file
// at the top of the source tree.
//
// Based on agi-test.agi from asterisk source tree.
// Can be used both as a standalone AGI app or a FastAGI server.
// When called directly from the  asterisk dialplan it behaves like an ordinary AGI app.
// When started on its own outside asterisk it spawns as a FastAGI server and
// listens for connections on 127.0.0.1:4573
// This works with asterisk 1.4 and later, earlier versions didn't set any environment vars.

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/zaf/agi"
)

func main() {
	// Get Asterisk Environment Variables. These are set by asterisk when
	// directly excetues an AGI application and they are not available in FastAGI sessions.
	astEnv := getAstEnv()

	// We check the availability of the Environment Variables to determine if the program
	// should behave as AGI app or FastAGI server.
	if astEnv["AST_AGI_DIR"] != "" {
		// Started as a standalone AGI app by asterisk.
		spawnAgi(nil)
	} else {
		// If started as a FastAGI server create a listener on port 4573
		// and start a new goroutine for each connection.
		ln, err := net.Listen("tcp", ":4573")
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Listening for FastAGI connections on 127.0.0.1:4573\n")
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			go spawnAgi(conn)
		}
	}
}

//Parse and store asterisk Environment Variables.
func getAstEnv() map[string]string {
	var env = map[string]string{
		"AST_CONFIG_DIR":  os.Getenv("AST_CONFIG_DIR"),
		"AST_CONFIG_FILE": os.Getenv("AST_CONFIG_FILE"),
		"AST_MODULE_DIR":  os.Getenv("AST_MODULE_DIR"),
		"AST_SPOOL_DIR":   os.Getenv("AST_SPOOL_DIR"),
		"AST_MONITOR_DIR": os.Getenv("AST_MONITOR_DIR"),
		"AST_VAR_DIR":     os.Getenv("AST_VAR_DIR"),
		"AST_DATA_DIR":    os.Getenv("AST_DATA_DIR"),
		"AST_LOG_DIR":     os.Getenv("AST_LOG_DIR"),
		"AST_AGI_DIR":     os.Getenv("AST_AGI_DIR"),
		"AST_KEY_DIR":     os.Getenv("AST_KEY_DIR"),
		"AST_RUN_DIR":     os.Getenv("AST_RUN_DIR"),
	}
	return env
}

// Start the AGI or FastAGI session.
func spawnAgi(c net.Conn) {
	myAgi := agi.New()
	var err error
	if c != nil {
		// Create a new FastAGI session.
		rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
		err = myAgi.Init(rw)
		defer c.Close()
	} else {
		// Create a new AGI session.
		err = myAgi.Init(nil)
	}
	if err != nil {
		log.Printf("Error Parsing AGI environment: %v\n", err)
		return
	}
	testAgi(myAgi)
	return
}

// Run the actual tests.
func testAgi(sess *agi.Session) {
	var tests, pass int
	var err error
	var r agi.Reply

	// For each test we Diesplay a message on the asterisk console with verbose,
	// send an AGI command and store the return values in r.
	sess.Verbose("Testing answer...")
	r, err = sess.Answer()
	// We check if there was an error or if the return values were not the expected ones.
	if err != nil || r.Res != 0 {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing channelstatus...")
	r, err = sess.ChannelStatus()
	if err != nil || r.Res != 6 {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing databaseput...")
	r, err = sess.DatabasePut("test", "my_key", "true")
	if err != nil || r.Res != 1 {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing databaseget...")
	r, err = sess.DatabaseGet("test", "my_key")
	if err != nil || r.Res != 1 || r.Dat != "true" {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing databasedel...")
	r, err = sess.DatabaseDel("test", "my_key")
	if err != nil || r.Res != 1 {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing databasedeltree...")
	r, err = sess.DatabaseDelTree("test")
	if err != nil || r.Res != 1 {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing streamfile...")
	r, err = sess.StreamFile("beep", "")
	if err != nil || r.Res != 0 {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing sendtext...")
	r, err = sess.SendText("Hello World")
	if err != nil || r.Res != 0 {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing sendimage...")
	r, err = sess.SendImage("asterisk-image")
	if err != nil || r.Res != 0 {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing saynumber...")
	r, err = sess.SayNumber(192837465, "")
	if err != nil || r.Res != 0 {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing wait for digit...")
	r, err = sess.WaitForDigit(3000)
	if err != nil || r.Res == -1 {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing record...")
	r, err = sess.RecordFile("/tmp/testagi", "alaw", "1234567890*#", 3000)
	if err != nil || r.Res == -1 {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing record playback...")
	r, err = sess.StreamFile("/tmp/testagi", "")
	if err != nil || r.Res != 0 {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing set variable...")
	r, err = sess.SetVariable("testagi", "foo")
	if err != nil || r.Res != 1 {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing get variable...")
	r, err = sess.GetVariable("testagi")
	if err != nil || r.Res != 1 || r.Dat != "foo" {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing get full variable...")
	r, err = sess.GetFullVariable("${testagi}")
	if err != nil || r.Res != 1 || r.Dat != "foo" {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("Testing exec...")
	r, err = sess.Exec("Wait", "3")
	if err != nil || r.Res != 0 {
		sess.Verbose("Failed.")
	} else {
		pass++
	}
	tests++

	sess.Verbose("================== Complete ======================")
	sess.Verbose(fmt.Sprintf("%d tests completed, %d passed, %d failed", tests, pass, tests-pass))
	sess.Verbose("==================================================")
	return
}
