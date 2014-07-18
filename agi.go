// Copyright (C) 2013 - 2014, Lefteris Zafiris <zaf.000@gmail.com>
// This program is free software, distributed under the terms of
// the BSD 3-Clause License. See the LICENSE file
// at the top of the source tree.

// Package agi implements the Asterisk Gateway Interface (http://www.asterisk.org). All AGI commands are
// implemented as methods of the Session struct that holds a copy of the AGI environment variables.
// All methods return a Reply struct and the AGI error, if any. The Reply struct contains
// the numeric result of the AGI command in Res and if there is any additional data it is stored
// as string in the Dat element of the struct.
package agi

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	envMin = 18  // Minimun number of AGI environment args
	envMax = 150 // Maximum number of AGI environment args
)

// Session is a struct holding AGI environment vars and the I/O handlers.
type Session struct {
	Env map[string]string //AGI environment variables.
	buf *bufio.ReadWriter //AGI I/O buffer.
}

// Reply is a struct that holds the return values of each AGI command.
type Reply struct {
	Res int    //Numeric result of the AGI command.
	Dat string //Additional returned data.
}

// New creates a new Session and returns a pointer to it.
func New() *Session {
	a := new(Session)
	a.Env = make(map[string]string, envMin+5)
	return a
}

// Init initializes a new AGI session. If rw is nil the AGI session will use standard input (stdin)
// and output (stdout) for a standalone AGI application. It reads and stores the AGI environment
// variables in Env. Returns an error if the parsing of the AGI environment was unsuccessful.
//
// For example, we create a new AGI session, initialize it and print the Env variables
// by using:
//
//	myAgi := agi.New()
//	err := myAgi.Init(nil)
//	if err != nil {
//		log.Fatal("Error Parsing AGI environment: %v\n", err)
//	}
//	for key, value := range myAgi.Env {
//		log.Printf("%-15s: %s\n", key, value)
//	}
//
func (a *Session) Init(rw *bufio.ReadWriter) error {
	if rw == nil {
		a.buf = bufio.NewReadWriter(bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout))
	} else {
		a.buf = rw
	}
	err := a.parseEnv()
	return err
}

// Answer answers channel. Res is -1 on channel failure, or 0 if successful.
func (a *Session) Answer() (Reply, error) {
	return a.sendMsg(fmt.Sprintf("ANSWER"))
}

// AsyncagiBreak interrupts Async AGI. Res is always 0.
func (a *Session) AsyncagiBreak() (Reply, error) {
	return a.sendMsg(fmt.Sprintf("ASYNCAGI BREAK"))
}

//ChannelStatus Res contains the status of the given channel, if no channel specified
// checks the current channel.
// Result values:
//     0 - Channel is down and available.
//     1 - Channel is down, but reserved.
//     2 - Channel is off hook.
//     3 - Digits (or equivalent) have been dialed.
//     4 - Line is ringing.
//     5 - Remote end is ringing.
//     6 - Line is up.
//     7 - Line is busy.
func (a *Session) ChannelStatus(channel ...string) (Reply, error) {
	if len(channel) > 0 {
		return a.sendMsg(fmt.Sprintf("CHANNEL STATUS %s", channel[0]))
	}
	return a.sendMsg(fmt.Sprintf("CHANNEL STATUS"))
}

// ControlStreamFile sends audio file on channel and allows the listener to control the stream.
// Optional parameters: skipms, ffchar - Defaults to *, rewchr - Defaults to #, pausechr.
// Res is 0 if playback completes without a digit being pressed, or the ASCII numerical value
// of the digit if one was pressed, or -1 on error or if the channel was disconnected.
func (a *Session) ControlStreamFile(file, escape string, params ...interface{}) (Reply, error) {
	cmd := fmt.Sprintf("%s \"%s\"", file, escape)
	for _, par := range params {
		cmd = fmt.Sprintf("%s %v", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("CONTROL STREAM FILE %s", cmd))
}

// DatabaseDel removes database key/value. Res is 1 if successful, 0 otherwise.
func (a *Session) DatabaseDel(family, key string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("DATABASE DEL %s %s", family, key))
}

// DatabaseDelTree removes database keytree/value. Res is 1 if successful, 0 otherwise.
func (a *Session) DatabaseDelTree(family string, keytree ...string) (Reply, error) {
	if len(keytree) > 0 {
		return a.sendMsg(fmt.Sprintf("DATABASE DELTREE %s %s", family, keytree[0]))
	}
	return a.sendMsg(fmt.Sprintf("DATABASE DELTREE %s", family))
}

// DatabaseGet gets database value. Res is 0 if key is not set, 1 if key is set
// and the value is returned in Dat.
func (a *Session) DatabaseGet(family, key string) (Reply, error) {
	r, err := a.sendMsg(fmt.Sprintf("DATABASE GET %s %s", family, key))
	if r.Dat != "" {
		r.Dat = strings.Trim(r.Dat, "()")
	}
	return r, err
}

// DatabasePut adds/updates database value. Res is 1 if successful, 0 otherwise.
func (a *Session) DatabasePut(family, key, value string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("DATABASE PUT %s %s %s", family, key, value))
}

// Exec executes a given application. Res contains whatever the dialplan application returns,
// or -2 on failure to find the application.
func (a *Session) Exec(app, options string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("EXEC %s %s", app, options))
}

// GetData prompts for DTMF on a channel. Optional parameters: timeout, maxdigits.
// Res contains the digits received from the channel at the other end.
func (a *Session) GetData(file string, params ...int) (Reply, error) {
	cmd := file
	for _, par := range params {
		cmd = fmt.Sprintf("%s %d", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("GET DATA %s", cmd))
}

// GetFullVariable evaluates a channel expression, if no channel is specified the current channel is used.
// Res is 1 if variable is set and the value is returned in Dat.
// Understands complex variable names and build in variables.
func (a *Session) GetFullVariable(variable string, channel ...string) (Reply, error) {
	var r Reply
	var err error
	if len(channel) > 0 {
		r, err = a.sendMsg(fmt.Sprintf("GET FULL VARIABLE %s %s", variable, channel[0]))
	} else {
		r, err = a.sendMsg(fmt.Sprintf("GET FULL VARIABLE %s", variable))
	}
	if r.Dat != "" {
		r.Dat = strings.Trim(r.Dat, "()")
	}
	return r, err
}

// GetOption streams file, prompts for DTMF with timeout. Optional parameter: timeout.
// Res contains the digits received from the channel at the other end and Dat
// contains the sample ofset. In case of failure to playback Res is -1.
func (a *Session) GetOption(filename, escape string, timeout ...int) (Reply, error) {
	var r Reply
	var err error
	if len(timeout) > 0 {
		r, err = a.sendMsg(fmt.Sprintf("GET OPTION %s \"%s\" %d", filename, escape, timeout[0]))
	} else {
		r, err = a.sendMsg(fmt.Sprintf("GET OPTION %s \"%s\"", filename, escape))
	}
	if r.Dat != "" {
		r.Dat = strings.TrimPrefix(r.Dat, "endpos=")
	}
	return r, err
}

// GetVariable gets a channel variable. Res is 0 if variable is not set,
// 1 if variable is set and Dat contains the value.
func (a *Session) GetVariable(variable string) (Reply, error) {
	r, err := a.sendMsg(fmt.Sprintf("GET VARIABLE %s", variable))
	if r.Dat != "" {
		r.Dat = strings.Trim(r.Dat, "()")
	}
	return r, err
}

// GoSub causes the channel to execute the specified dialplan subroutine, returning to the dialplan
// with execution of a Return().
func (a *Session) GoSub(context, extension, priority, args string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("GOSUB %s %s %s %s", context, extension, priority, args))
}

// Hangup hangs up a channel, Res is 1 on success, -1 if the given channel was not found.
func (a *Session) Hangup(channel ...string) (Reply, error) {
	var r Reply
	var err error
	if len(channel) > 0 {
		r, err = a.sendMsg(fmt.Sprintf("HANGUP %s", channel[0]))
	} else {
		r, err = a.sendMsg(fmt.Sprintf("HANGUP"))
	}
	a.buf.ReadBytes(10) //Read 'HANGUP' command from asterisk
	return r, err
}

// Noop does nothing. Res is always 0.
func (a *Session) Noop(params ...interface{}) (Reply, error) {
	var cmd string
	for _, par := range params {
		cmd = fmt.Sprintf("%s %v", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("NOOP%s", cmd))
}

// RawCommand sends a user defined command. Use of this is generally discouraged.
// Useful only for debugging, testing and maybe compatibility with very old versions of asterisk.
func (a *Session) RawCommand(params ...interface{}) (Reply, error) {
	var cmd string
	for _, par := range params {
		cmd = fmt.Sprintf("%s %v", cmd, par)
	}
	return a.sendMsg(cmd)
}

// ReceiveChar receives one character from channels supporting it. Res contains the decimal value of
// the character if one is received, or 0 if the channel does not support text reception.
// Result is -1 only on error/hang-up.
func (a *Session) ReceiveChar(timeout int) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("RECEIVE CHAR %d", timeout))
}

// ReceiveText receives text from channels supporting it. Res is -1 for failure
// or 1 for success, and Dat contains the string.
func (a *Session) ReceiveText(timeout int) (Reply, error) {
	r, err := a.sendMsg(fmt.Sprintf("RECEIVE TEXT %d", timeout))
	if r.Dat != "" {
		r.Dat = strings.Trim(r.Dat, "()")
	}
	return r, err
}

// RecordFile records to a given file. The format will specify what kind of file will be recorded.
// The timeout is the maximum record time in milliseconds, or -1 for no timeout.
// Optional parameters: offset samples if provided will seek to the offset without exceeding
// the end of the file. Silence (always prefixed with s=)is the number of seconds of silence allowed
// before the function returns despite the lack of DTMF digits or reaching timeout.
// Negative values of Res mean error. Res equals 0 if recording was successful without any DTMF
// digit received, or the ASCII numerical value of the digit if one was received.
// Dat contains a set of different inconsistent return values depending on each case,
// please refer to res_agi.c in asterisk source code for further info.
func (a *Session) RecordFile(file, format, escape string, timeout int, params ...interface{}) (Reply, error) {
	cmd := fmt.Sprintf("%s %s \"%s\" %d", file, format, escape, timeout)
	for _, par := range params {
		cmd = fmt.Sprintf("%s %v", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("RECORD FILE %s", cmd))
}

// SayAlpha says a given character string. Res is 0 if playback completes without a digit
// being pressed, the ASCII numerical value of the digit if one was pressed or -1 on error/hang-up.
func (a *Session) SayAlpha(str, escape string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SAY ALPHA %s \"%s\"", str, escape))
}

// SayDate says a given date (Unix time format). Res is 0 if playback completes without a digit
// being pressed, the ASCII numerical value of the digit if one was pressed or -1 on error/hang-up.
func (a *Session) SayDate(date int64, escape string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SAY DATE %d \"%s\"", date, escape))
}

// SayDateTime says a given time (Unix time format). Optional parameters:
// fomrat, the format the time should be said in. See voicemail.conf (defaults to ABdY 'digits/at' IMp).
// timezone, acceptable values can be found in /usr/share/zoneinfo. Defaults to machine default.
// Res is 0 if playback completes without a digit being pressed, the ASCII numerical
// value of the digit if one was pressed or -1 on error/hang-up.
func (a *Session) SayDateTime(time int64, escape string, params ...string) (Reply, error) {
	cmd := fmt.Sprintf("%d \"%s\"", time, escape)
	for _, par := range params {
		cmd = fmt.Sprintf("%s %v", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("SAY DATETIME %s", cmd))
}

// SayDigits says a given digit. Res is 0 if playback completes without a digit being pressed,
// the ASCII numerical value of the digit if one was pressed or -1 on error/hang-up.
func (a *Session) SayDigits(digit int, escape string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SAY DIGITS %d \"%s\"", digit, escape))
}

// SayNumber says a given number. Optional parameter gender. Res is 0 if playback completes
// without a digit being pressed, the ASCII numerical value of the digit if one was pressed or -1 on error/hang-up.
func (a *Session) SayNumber(num int, escape string, gender ...string) (Reply, error) {
	if len(gender) > 0 {
		return a.sendMsg(fmt.Sprintf("SAY NUMBER %d \"%s\" %s", num, escape, gender[0]))
	}
	return a.sendMsg(fmt.Sprintf("SAY NUMBER %d \"%s\"", num, escape))
}

// SayPhonetic says a given character string with phonetics. Res is 0 if playback completes
// without a digit pressed, the ASCII numerical value of the digit if one was pressed, or -1 on error/hang-up
func (a *Session) SayPhonetic(str, escape string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SAY PHONETIC %s \"%s\"", str, escape))
}

// SayTime says a given time (Unix time format). Res is 0 if playback completes without a digit
// being pressed, or the ASCII numerical value of the digit if one was pressed or -1 on error/hang-up.
func (a *Session) SayTime(time int64, escape string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SAY TIME %d \"%s\"", time, escape))
}

// SendImage sends images to channels supporting it. Res is 0 if image is sent, or if the channel
// does not support image transmission. Result is -1 only on error/hang-up. Image names should not include extensions.
func (a *Session) SendImage(image string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SEND IMAGE %s", image))
}

// SendText sends text to channels supporting it. Res is 0 if text is sent, or if the channel
// does not support text transmission. Result is -1 only on error/hang-up.
func (a *Session) SendText(text string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SEND TEXT \"%s\"", text))
}

// SetAutohangup autohang-ups channel after a number of seconds. Setting time to 0 will cause the autohang-up
// feature to be disabled on this channel. Res is always 0.
func (a *Session) SetAutohangup(time int) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SET AUTOHANGUP %d", time))
}

// SetCallerid sets callerid for the current channel. Res is always 1.
func (a *Session) SetCallerid(cid string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SET CALLERID %s", cid))
}

// SetContext sets channel context. Res is always 0.
func (a *Session) SetContext(context string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SET CONTEXT %s", context))
}

// SetExtension changes channel extension. Res is always 0.
func (a *Session) SetExtension(ext string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SET EXTENSION %s", ext))
}

// SetMusic enables/disables Music on hold generator by setting opt to "on" or "off".
// Optional parameter: class, if not specified, then the default music on hold class will be used.
// Res is always 0.
func (a *Session) SetMusic(opt string, class ...string) (Reply, error) {
	if len(class) > 0 {
		return a.sendMsg(fmt.Sprintf("SET MUSIC %s %s", opt, class[0]))
	}
	return a.sendMsg(fmt.Sprintf("SET MUSIC %s", opt))
}

// SetPriority sets channel dialplan priority. The priority must be a valid priority or label.
// Res is always 0.
func (a *Session) SetPriority(priority string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SET PRIORITY %s", priority))
}

// SetVariable sets a channel variable. Res is always 1.
func (a *Session) SetVariable(variable string, value interface{}) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SET VARIABLE \"%s\" \"%v\"", variable, value))
}

// SpeechActivateGrammar activates a grammar. Res is 1 on success 0 on error.
func (a *Session) SpeechActivateGrammar(grammar string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH ACTIVATE GRAMMAR %s", grammar))
}

// SpeechCreate creates a speech object. Res is 1 on success 0 on error.
func (a *Session) SpeechCreate(engine string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH CREATE %s", engine))
}

// SpeechDeactivateGrammar deactivates a grammar. Res is 1 on success 0 on error.
func (a *Session) SpeechDeactivateGrammar(grammar string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH DEACTIVATE GRAMMAR %s", grammar))
}

// SpeechDestroy destroys a speech object. Res is 1 on success 0 on error.
func (a *Session) SpeechDestroy() (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH DESTROY"))
}

// SpeechLoadGrammar loads a grammar. Res is 1 on success 0 on error.
func (a *Session) SpeechLoadGrammar(grammar, path string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH LOAD GRAMMAR %s %s", grammar, path))
}

// SpeechRecognize recognizes speech. Res is 1 onsuccess, 0 in case of error
// In case of success Dat contains a set of different inconsistent values.
// Please refer to res_agi.c in asterisk source code for further info.
func (a *Session) SpeechRecognize(prompt, timeout, offset string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH RECOGNIZE %s %s %s", prompt, timeout, offset))
}

// SpeechSet sets a speech engine setting. Res is 1 on success 0 on error.
func (a *Session) SpeechSet(name, value string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH SET %s %s", name, value))
}

// SpeechUnloadGrammar unloads a grammar. Result is 1 on success 0 on error.
func (a *Session) SpeechUnloadGrammar(grammar string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH UNLOAD GRAMMAR %s", grammar))
}

// StreamFile sends audio file on channel. Optional parameter: sample offset for the playback start position.
// Res is 0 if playback completes without a digit being pressed, the ASCII numerical value
// of the digit if one was pressed, or -1 on error or if the channel was disconnected.
// If musiconhold is playing before calling stream file it will be automatically stopped
// and will not be restarted after completion.
func (a *Session) StreamFile(file, escape string, offset ...int) (Reply, error) {
	var r Reply
	var err error
	if len(offset) > 0 {
		r, err = a.sendMsg(fmt.Sprintf("STREAM FILE %s \"%s\" %d", file, escape, offset[0]))
	} else {
		r, err = a.sendMsg(fmt.Sprintf("STREAM FILE %s \"%s\"", file, escape))
	}
	if r.Dat != "" {
		r.Dat = strings.TrimPrefix(r.Dat, "endpos=")
	}
	return r, err
}

// TddMode toggles TDD mode (for the deaf). Res is 1 if successful, or 0 if channel is not TDD-capable.
func (a *Session) TddMode(mode string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("TDD MODE %s", mode))
}

// Verbose logs a message to the asterisk verbose log.
// Optional variable: level, the verbose level (1-4). Res is always 1.
func (a *Session) Verbose(msg string, level ...int) (Reply, error) {
	if len(level) > 0 {
		return a.sendMsg(fmt.Sprintf("VERBOSE \"%s\" %d", msg, level[0]))
	}
	return a.sendMsg(fmt.Sprintf("VERBOSE \"%s\"", msg))
}

// WaitForDigit waits for a digit to be pressed. Use -1 for the timeout value if you desire
// the call to block indefinitely. Res is -1 on channel failure, 0 if no digit is received
// in the timeout, or the ASCII numerical value of the digit if one is received.
func (a *Session) WaitForDigit(timeout int) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("WAIT FOR DIGIT %d", timeout))
}

// parseEnv reads and stores AGI environment.
func (a *Session) parseEnv() error {
	var err error
	for i := 0; i <= envMax; i++ {
		line, err := a.buf.ReadBytes(10)
		if err != nil || len(line) <= len("\r\n") {
			break
		}
		//Strip trailing newline
		line = line[:len(line)-1]
		ind := bytes.IndexByte(line, ':')
		//"agi_type" is the shortest length key, "agi_network_script" the longest, anything ouside these boundaries is invalid.
		if ind < len("agi_type") || ind > len("agi_network_script") || ind == len(line)-1 {
			//line doesn't match: /^.{8,18}:.+$/
			err = fmt.Errorf("malformed environment input: %s", string(line))
			a.Env = nil
			return err
		}
		//Strip 'agi_' prefix from key.
		key := string(line[len("agi_"):ind])
		//Strip leading colon and space from value.
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

// sendMsg sends an AGI command and returns the result. In case of error Reply.Res is set to -99
// for people that wont bother doing proper error checking. :(
func (a *Session) sendMsg(s string) (Reply, error) {
	s = strings.TrimSpace(s)
	if _, err := fmt.Fprintln(a.buf, s); err != nil {
		return Reply{-99, ""}, err
	}
	if err := a.buf.Flush(); err != nil {
		return Reply{-99, ""}, err
	}
	return a.parseResponse()
}

// parseResponse reads back and parses AGI repsonse. Returns the Reply and the protocol error, if any.
// In case of error Res is again set to -99 for the aforementioned reason.
func (a *Session) parseResponse() (Reply, error) {
	r := Reply{-99, ""}
	line, err := a.buf.ReadBytes(10)
	if err != nil {
		return r, err
	}
	//Strip trailing newline
	line = line[:len(line)-1]
	ind := bytes.IndexByte(line, ' ')
	if ind <= 0 || ind == len(line)-1 {
		//line doesnt match /^\w+\s.+$/
		if string(line) == "HANGUP" {
			err = fmt.Errorf("client sent a HANGUP request")
		} else {
			err = fmt.Errorf("malformed or partial agi response: %s", string(line))
		}
		return r, err
	}
	switch string(line[:ind]) {
	case "200":
		eqInd := bytes.IndexByte(line, '=')
		//Check if line matches /^200\s\w{7}=.*$/
		if eqInd == len("200 result") && eqInd < len(line)-1 {
			//Strip the "200 result=" prefix.
			line = line[eqInd+1:]
			spInd := bytes.IndexByte(line, ' ')
			if spInd < 0 {
				//line matches /^\w$/
				r.Res, err = strconv.Atoi(string(line))
				if err != nil {
					err = fmt.Errorf("failed to parse AGI 200 reply: %v", err)
					r.Res = -99
				}
			} else if spInd > 0 && spInd < len(line)-1 {
				//line matches /^\w+\s.+$/
				r.Res, err = strconv.Atoi(string(line[:spInd]))
				if err != nil {
					err = fmt.Errorf("failed to parse AGI 200 reply: %v", err)
					r.Res = -99
				}
				//Strip leading space and save additional returned data.
				r.Dat = string(line[spInd+1:])
			}
			break
		}
		err = fmt.Errorf("malformed 200 response: %s", string(line))
	case "510":
		err = fmt.Errorf("invalid or unknown command")
	case "511":
		err = fmt.Errorf("command not permitted on a dead channel")
	case "520":
		err = fmt.Errorf("invalid command syntax")
	case "520-Invalid":
		err = fmt.Errorf("invalid command syntax")
		a.buf.ReadBytes(10) //Read Command syntax doc.
	default:
		err = fmt.Errorf("malformed or partial agi response: %s", string(line))
	}
	return r, err
}
