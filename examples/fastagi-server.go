// An example of implementing an Fast AGI server in Go
//
// A request formed like the following:
// agi(agi://127.0.0.1/playback?file=foo)
// plays back file 'foo' to the user
//
// Copyright (C) 2013 - 2015, Lefteris Zafiris <zaf.000@gmail.com>
// This program is free software, distributed under the terms of
// the BSD 3-Clause License. See the LICENSE file
// at the top of the source tree.

package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"

	"github.com/zaf/agi"
)

var (
	debug     = flag.Bool("debug", false, "Print debug information on stderr")
	listen    = flag.String("listen", "127.0.0.1", "Listening address")
	port      = flag.String("port", "4573", "Listening server port")
	listeners = flag.Int("pool", 4, "Pool size of Listeners")
)

func main() {
	flag.Parse()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	var shutdown int32

	addr := net.JoinHostPort(*listen, *port)
	log.Printf("Starting FastAGI server on %v\n", addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()
	wg := new(sync.WaitGroup)
	for i := 0; i < *listeners; i++ {
		go func() {
			for atomic.LoadInt32(&shutdown) == 0 {
				conn, err := listener.Accept()
				if err != nil {
					log.Println(err)
					continue
				}
				if *debug {
					log.Printf("Connected: %v <-> %v\n", conn.LocalAddr(), conn.RemoteAddr())
				}
				wg.Add(1)
				go agiConnHandle(conn, wg)
			}
		}()
	}
	signal := <-c
	log.Printf("Received %v, Waiting for remaining sessions to end and exit.\n", signal)
	atomic.StoreInt32(&shutdown, 1)
	wg.Wait()
}

func agiConnHandle(client net.Conn, wg *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Session terminated:", err)
		}
		if *debug {
			log.Printf("Closing connection from %v", client.RemoteAddr())
		}
		client.Close()
		wg.Done()
	}()
	// Create a new AGI session
	myAgi := agi.New()
	rw := bufio.NewReadWriter(bufio.NewReader(client), bufio.NewWriter(client))
	err := myAgi.Init(rw)
	checkErr(err)
	var file string
	var rep agi.Reply
	if *debug {
		// Print AGI environment
		log.Println("AGI environment vars:")
		for key, value := range myAgi.Env {
			log.Printf("%-15s: %s\n", key, value)
		}
	}
	// Parse AGI recuest
	_, query := parseAgiReq(myAgi.Env["request"])
	if query["file"] == nil {
		if *debug {
			log.Println("No arguments passed, exiting")
		}
		goto HANGUP
	}
	file = query["file"][0]
	// Chech channel status
	rep, err = myAgi.ChannelStatus()
	checkErr(err)
	// Answer channel if not already answered
	if rep.Res != 6 {
		rep, err = myAgi.Answer()
		checkErr(err)
		if rep.Res == -1 {
			log.Printf("Failed to answer channel\n")
			return
		}
	}
	// Playback file
	rep, err = myAgi.StreamFile(file, "1234567890#*")
	checkErr(err)
	if rep.Res == -1 {
		log.Printf("Failed to playback file: %s\n", file)
	}
HANGUP:
	myAgi.Hangup()
	return
}

// Parse AGI reguest return path and query params
func parseAgiReq(request string) (string, url.Values) {
	req, _ := url.Parse(request)
	query, _ := url.ParseQuery(req.RawQuery)
	return req.Path, query
}

//Check for AGI Protocol errors or hangups
func checkErr(e error) {
	if e != nil {
		if e.Error() == "HANGUP" {
			panic("Client Hangup")
		}
		panic("AGI error: " + e.Error())
	}
}
