// FastAGI 'Hello World' example in Go
//
// Copyright (C) 2013 - 2014, Lefteris Zafiris <zaf.000@gmail.com>
// This program is free software, distributed under the terms of
// the BSD 3-Clause License. See the LICENSE file
// at the top of the source tree.

package main

import (
	"bufio"
	"flag"
	"github.com/zaf/agi"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
)

var (
	debug  = flag.Bool("debug", false, "Print debug information on stderr")
	listen = flag.String("listen", "127.0.0.1", "Listening address")
	port   = flag.String("port", "4573", "Listening server port")
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	shutdown := false

	addr := net.JoinHostPort(*listen, *port)
	log.Printf("Starting FastAGI server on %v\n", addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()
	wg := new(sync.WaitGroup)
	go func() {
		for !shutdown {
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
	signal := <-c
	log.Printf("Received %v, Waiting for remaining sessions to end and exit.\n", signal)
	shutdown = true
	wg.Wait()
}

func agiConnHandle(client net.Conn, wg *sync.WaitGroup) {
	//Create a new AGI session
	rw := bufio.NewReadWriter(bufio.NewReader(client), bufio.NewWriter(client))
	myAgi, err := agi.Init(rw)
	defer func() {
		if *debug {
			log.Printf("Closing connection from %v", client.RemoteAddr())
		}
		client.Close()
		myAgi.Destroy()
		wg.Done()
	}()
	if err != nil {
		log.Printf("Error Parsing AGI environment: %v\n", err)
		return
	}
	if *debug {
		//Print AGI environment
		log.Println("AGI environment vars:")
		for key, value := range myAgi.Env {
			log.Printf("%-15s: %s\n", key, value)
		}
	}
	//Print a message on asterisk console
	err = myAgi.Verbose("Hello World", "1")
	if err != nil {
		log.Printf("AGI reply error: %v\n", err)
		return
	}
	if *debug {
		//Print the response
		log.Printf("AGI command returned: %v\n", myAgi.Res)
	}
	//Hangup
	myAgi.Hangup("")
	return
}
