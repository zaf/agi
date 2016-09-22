// Copyright (C) 2013 - 2015, Lefteris Zafiris <zaf@fastmail.com>
// This program is free software, distributed under the terms of
// the BSD 3-Clause License. See the LICENSE file
// at the top of the source tree.

//Package agi implements the Asterisk Gateway Interface (http://www.asterisk.org).
package agi

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	envMin = 18  // Minimum number of AGI environment args
	envMax = 150 // Maximum number of AGI environment args
)

// parseEnv reads and stores AGI environment.
func (a *Session) parseEnv() error {
	var err error
	var line []byte
	for i := 0; i <= envMax; i++ {
		line, err = a.buf.ReadBytes(10)
		if err != nil || len(line) <= len("\r\n") {
			break
		}
		// Strip trailing newline
		line = line[:len(line)-1]
		ind := bytes.IndexByte(line, ':')
		// "agi_type" is the shortest length key, "agi_network_script" the longest, anything outside these boundaries is invalid.
		if ind < len("agi_type") || ind > len("agi_network_script") || ind == len(line)-1 {
			err = fmt.Errorf("malformed environment input: %s", string(line))
			a.Env = nil
			return err
		}
		key := string(line[len("agi_"):ind])
		ind += len(": ")
		value := string(line[ind:])
		a.Env[key] = value
	}
	if len(a.Env) < envMin {
		err = fmt.Errorf("incomplete environment with only %d env vars", len(a.Env))
		a.Env = nil
	}
	return err
}

// sendMsg sends an AGI command and returns the result.
func (a *Session) sendMsg(s string) (Reply, error) {
	// Make sure there wasn't any data received, usually a HANGUP request from asterisk.
	if i := a.buf.Reader.Buffered(); i != 0 {
		line, _ := a.buf.ReadBytes(10)
		return Reply{}, fmt.Errorf(string(line[:len(line)-1]))
	}
	s = strings.Replace(s, "\r", " ", -1)
	s = strings.Replace(s, "\n", " ", -1)
	if _, err := a.buf.WriteString(s + "\n"); err != nil {
		return Reply{}, err
	}
	if err := a.buf.Flush(); err != nil {
		return Reply{}, err
	}
	return a.parseResponse()
}

// parseResponse reads back and parses AGI response. Returns the Reply and the protocol error, if any.
func (a *Session) parseResponse() (Reply, error) {
	r := Reply{}
	line, err := a.buf.ReadBytes(10)
	if err != nil {
		return r, err
	}
	// Strip trailing newline
	line = line[:len(line)-1]
	ind := bytes.IndexByte(line, ' ')
	if ind <= 0 || ind == len(line)-1 {
		// Line doesn't match /^\w+\s.+$/
		if bytes.Equal(line, []byte("HANGUP")) {
			err = errors.New("HANGUP")
		} else {
			err = fmt.Errorf("malformed or partial agi response: %s", string(line))
		}
		return r, err
	}
	switch string(line[:ind]) {
	case "200":
		eqInd := bytes.IndexByte(line, '=')
		if eqInd == len("200 result") && eqInd < len(line)-1 {
			// If line matches /^200\s\w{7}=.*$/ strip the "200 result=" prefix.
			line = line[eqInd+1:]
			spInd := bytes.IndexByte(line, ' ')
			if spInd < 0 {
				// Line matches /^\w$/
				r.Res, err = strconv.Atoi(string(line))
				if err != nil {
					err = fmt.Errorf("failed to parse AGI 200 reply: %v", err)
				}
				break
			} else if spInd > 0 && spInd < len(line)-1 {
				// Line matches /^\w+\s.+$/
				r.Res, err = strconv.Atoi(string(line[:spInd]))
				if err != nil {
					err = fmt.Errorf("failed to parse AGI 200 reply: %v", err)
				}
				// Strip leading space and save additional returned data.
				r.Dat = string(line[spInd+1:])
				break
			}
		}
		err = fmt.Errorf("malformed 200 response: %s", string(line))
	case "510":
		err = errors.New("invalid or unknown command")
	case "511":
		err = errors.New("command not permitted on a dead channel")
	case "520":
		err = errors.New("invalid command syntax")
	case "520-Invalid":
		err = errors.New("invalid command syntax")
		a.buf.ReadBytes(10) // Read Command syntax doc.
	default:
		err = fmt.Errorf("malformed or partial agi response: %s", string(line))
	}
	return r, err
}
