// Copyright (C) 2013 - 2015, Lefteris Zafiris <zaf@fastmail.com>
// This program is free software, distributed under the terms of
// the BSD 3-Clause License. See the LICENSE file
// at the top of the source tree.

/*
Package agi implements the Asterisk Gateway Interface (http://www.asterisk.org). All AGI commands are
implemented as methods of the Session struct that holds a copy of the AGI environment variables.
All methods return a Reply struct and the AGI error, if any. The Reply struct contains
the numeric result of the AGI command in Res and if there is any additional data it is stored
as string in the Dat element of the struct.
For example, to create a new AGI session and initialize it:

	myAgi := agi.New()
	err := myAgi.Init(nil)
	if err != nil {
		log.Fatalf("Error Parsing AGI environment: %v\n", err)
	}

To play back a voice prompt using AGI 'Stream file' command:

	rep, err := myAgi.StreamFile("hello-world", "0123456789")
	if err != nil {
		log.Fatalf("AGI reply error: %v\n", err)
	}
	if rep.Res == -1 {
		log.Printf("Error during playback\n")
	}

For more please see the code snippets in the examples folder.
*/
package agi

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
	return a.sendMsg("ANSWER")
}

// AsyncagiBreak interrupts Async AGI. Res is always 0.
func (a *Session) AsyncagiBreak() (Reply, error) {
	return a.sendMsg("ASYNCAGI BREAK")
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
	if channel != nil {
		return a.sendMsg(fmt.Sprintf("CHANNEL STATUS %q", channel[0]))
	}
	return a.sendMsg("CHANNEL STATUS")
}

// ControlStreamFile sends audio file on channel and allows the listener to control the stream.
// Optional parameters: skipms, ffchar - Defaults to *, rewchr - Defaults to #, pausechr.
// Res is 0 if playback completes without a digit being pressed, or the ASCII numerical value
// of the digit if one was pressed, or -1 on error or if the channel was disconnected.
func (a *Session) ControlStreamFile(file, escape string, params ...interface{}) (Reply, error) {
	cmd := fmt.Sprintf("%q %q", file, escape)
	for _, par := range params {
		cmd = fmt.Sprintf("%s \"%v\"", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("CONTROL STREAM FILE %s", cmd))
}

// DatabaseDel removes database key/value. Res is 1 if successful, 0 otherwise.
func (a *Session) DatabaseDel(family, key string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("DATABASE DEL %q %q", family, key))
}

// DatabaseDelTree removes database keytree/value. Res is 1 if successful, 0 otherwise.
func (a *Session) DatabaseDelTree(family string, keytree ...string) (Reply, error) {
	if keytree != nil {
		return a.sendMsg(fmt.Sprintf("DATABASE DELTREE %q %q", family, keytree[0]))
	}
	return a.sendMsg(fmt.Sprintf("DATABASE DELTREE %q", family))
}

// DatabaseGet gets database value. Res is 0 if key is not set, 1 if key is set
// and the value is returned in Dat.
func (a *Session) DatabaseGet(family, key string) (Reply, error) {
	r, err := a.sendMsg(fmt.Sprintf("DATABASE GET %q %q", family, key))
	if r.Dat != "" {
		r.Dat = strings.TrimPrefix(r.Dat, "(")
		r.Dat = strings.TrimSuffix(r.Dat, ")")
	}
	return r, err
}

// DatabasePut adds/updates database value. Res is 1 if successful, 0 otherwise.
func (a *Session) DatabasePut(family, key, value string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("DATABASE PUT %q %q %q", family, key, value))
}

// Exec executes a given application. Res contains whatever the dialplan application returns,
// or -2 on failure to find the application.
func (a *Session) Exec(app, options string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("EXEC %s %q", app, options))
}

// Failure causes asterisk to terminate the AGI session and set the AGISTATUS channel variable to 'FAILURE'.
func (a *Session) Failure() (Reply, error) {
	return a.sendMsg("FAILURE")
}

// GetData prompts for DTMF on a channel. Optional parameters: timeout, maxdigits.
// Res contains the digits received from the channel at the other end.
func (a *Session) GetData(file string, params ...int) (Reply, error) {
	cmd := "\"" + file + "\""
	for _, par := range params {
		cmd = fmt.Sprintf("%s \"%d\"", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("GET DATA %s", cmd))
}

// GetFullVariable evaluates a channel expression, if no channel is specified the current channel is used.
// Res is 1 if variable is set and the value is returned in Dat.
// Understands complex variable names and build in variables.
func (a *Session) GetFullVariable(variable string, channel ...string) (Reply, error) {
	var r Reply
	var err error
	if channel != nil {
		r, err = a.sendMsg(fmt.Sprintf("GET FULL VARIABLE %q %q", variable, channel[0]))
	} else {
		r, err = a.sendMsg(fmt.Sprintf("GET FULL VARIABLE %q", variable))
	}
	if r.Dat != "" {
		r.Dat = strings.TrimPrefix(r.Dat, "(")
		r.Dat = strings.TrimSuffix(r.Dat, ")")
	}
	return r, err
}

// GetOption streams file, prompts for DTMF with timeout. Optional parameter: timeout.
// Res contains the digits received from the channel at the other end and Dat
// contains the sample ofset. In case of failure to playback Res is -1.
func (a *Session) GetOption(filename, escape string, timeout ...int) (Reply, error) {
	var r Reply
	var err error
	if timeout != nil {
		r, err = a.sendMsg(fmt.Sprintf("GET OPTION %q %q %d", filename, escape, timeout[0]))
	} else {
		r, err = a.sendMsg(fmt.Sprintf("GET OPTION %q %q", filename, escape))
	}
	if r.Dat != "" {
		r.Dat = strings.TrimPrefix(r.Dat, "endpos=")
	}
	return r, err
}

// GetVariable gets a channel variable. Res is 0 if variable is not set,
// 1 if variable is set and Dat contains the value.
func (a *Session) GetVariable(variable string) (Reply, error) {
	r, err := a.sendMsg(fmt.Sprintf("GET VARIABLE %q", variable))
	if r.Dat != "" {
		r.Dat = strings.TrimPrefix(r.Dat, "(")
		r.Dat = strings.TrimSuffix(r.Dat, ")")
	}
	return r, err
}

// GoSub causes the channel to execute the specified dialplan subroutine, returning to the dialplan
// with execution of a Return().
func (a *Session) GoSub(context, extension, priority, args string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("GOSUB %q %q %q %q", context, extension, priority, args))
}

// Hangup hangs up a channel, Res is 1 on success, -1 if the given channel was not found.
func (a *Session) Hangup(channel ...string) (Reply, error) {
	var r Reply
	var err error
	if channel != nil {
		r, err = a.sendMsg(fmt.Sprintf("HANGUP %q", channel[0]))
	} else {
		r, err = a.sendMsg("HANGUP")
	}
	//a.buf.ReadBytes(10) // Read 'HANGUP' command from asterisk
	return r, err
}

// Noop does nothing. Res is always 0.
func (a *Session) Noop(params ...interface{}) (Reply, error) {
	var cmd string
	for _, par := range params {
		cmd = fmt.Sprintf("%s \"%v\"", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("NOOP %s", cmd))
}

// RawCommand sends a user defined command. Use of this is generally discouraged.
// Useful only for debugging, testing and maybe compatibility with newer/altered versions of the AGI
// protocol.
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
	r, err := a.sendMsg(fmt.Sprintf("RECEIVE TEXT \"%d\"", timeout))
	if r.Dat != "" {
		r.Dat = strings.TrimPrefix(r.Dat, "(")
		r.Dat = strings.TrimSuffix(r.Dat, ")")
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
	cmd := fmt.Sprintf("%q %q %q %d", file, format, escape, timeout)
	for _, par := range params {
		cmd = fmt.Sprintf("%s \"%v\"", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("RECORD FILE %s", cmd))
}

// SayAlpha says a given character string. Res is 0 if playback completes without a digit
// being pressed, the ASCII numerical value of the digit if one was pressed or -1 on error/hang-up.
func (a *Session) SayAlpha(str, escape string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SAY ALPHA %q %q", str, escape))
}

// SayDate says a given date (Unix time format). Res is 0 if playback completes without a digit
// being pressed, the ASCII numerical value of the digit if one was pressed or -1 on error/hang-up.
func (a *Session) SayDate(date int64, escape string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SAY DATE \"%d\" %q", date, escape))
}

// SayDateTime says a given time (Unix time format). Optional parameters:
// format, the format the time should be said in. See voicemail.conf (defaults to ABdY 'digits/at' IMp).
// timezone, acceptable values can be found in /usr/share/zoneinfo. Defaults to machine default.
// Res is 0 if playback completes without a digit being pressed, the ASCII numerical
// value of the digit if one was pressed or -1 on error/hang-up.
func (a *Session) SayDateTime(time int64, escape string, params ...string) (Reply, error) {
	cmd := fmt.Sprintf("\"%d\" %q", time, escape)
	for _, par := range params {
		cmd = fmt.Sprintf("%s \"%v\"", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("SAY DATETIME %s", cmd))
}

// SayDigits says a given digit. Res is 0 if playback completes without a digit being pressed,
// the ASCII numerical value of the digit if one was pressed or -1 on error/hang-up.
func (a *Session) SayDigits(digit int, escape string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SAY DIGITS \"%d\" %q", digit, escape))
}

// SayNumber says a given number. Optional parameter gender. Res is 0 if playback completes
// without a digit being pressed, the ASCII numerical value of the digit if one was pressed or -1 on error/hang-up.
func (a *Session) SayNumber(num int, escape string, gender ...string) (Reply, error) {
	if gender != nil {
		return a.sendMsg(fmt.Sprintf("SAY NUMBER \"%d\" %q %q", num, escape, gender[0]))
	}
	return a.sendMsg(fmt.Sprintf("SAY NUMBER \"%d\" %q", num, escape))
}

// SayPhonetic says a given character string with phonetics. Res is 0 if playback completes
// without a digit pressed, the ASCII numerical value of the digit if one was pressed, or -1 on error/hang-up
func (a *Session) SayPhonetic(str, escape string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SAY PHONETIC %q %q", str, escape))
}

// SayTime says a given time (Unix time format). Res is 0 if playback completes without a digit
// being pressed, or the ASCII numerical value of the digit if one was pressed or -1 on error/hang-up.
func (a *Session) SayTime(time int64, escape string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SAY TIME \"%d\" %q", time, escape))
}

// SendImage sends images to channels supporting it. Res is 0 if image is sent, or if the channel
// does not support image transmission. Result is -1 only on error/hang-up. Image names should not include extensions.
func (a *Session) SendImage(image string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SEND IMAGE %q", image))
}

// SendText sends text to channels supporting it. Res is 0 if text is sent, or if the channel
// does not support text transmission. Result is -1 only on error/hang-up.
func (a *Session) SendText(text string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SEND TEXT %q", text))
}

// SetAutohangup autohang-ups channel after a number of seconds. Setting time to 0 will cause the autohang-up
// feature to be disabled on this channel. Res is always 0.
func (a *Session) SetAutohangup(time int) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SET AUTOHANGUP \"%d\"", time))
}

// SetCallerid sets callerid for the current channel. Res is always 1.
func (a *Session) SetCallerid(cid string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SET CALLERID %q", cid))
}

// SetContext sets channel context. Res is always 0.
func (a *Session) SetContext(context string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SET CONTEXT %q", context))
}

// SetExtension changes channel extension. Res is always 0.
func (a *Session) SetExtension(ext string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SET EXTENSION %q", ext))
}

// SetMusic enables/disables Music on hold generator by setting opt to "on" or "off".
// Optional parameter: class, if not specified, then the default music on hold class will be used.
// Res is always 0.
func (a *Session) SetMusic(opt string, class ...string) (Reply, error) {
	if class != nil {
		return a.sendMsg(fmt.Sprintf("SET MUSIC %q %q", opt, class[0]))
	}
	return a.sendMsg(fmt.Sprintf("SET MUSIC %q", opt))
}

// SetPriority sets channel dialplan priority. The priority must be a valid priority or label.
// Res is always 0.
func (a *Session) SetPriority(priority string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SET PRIORITY %q", priority))
}

// SetVariable sets a channel variable. Res is always 1.
func (a *Session) SetVariable(variable string, value interface{}) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SET VARIABLE %q \"%v\"", variable, value))
}

// SpeechActivateGrammar activates a grammar. Res is 1 on success 0 on error.
func (a *Session) SpeechActivateGrammar(grammar string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH ACTIVATE GRAMMAR %q", grammar))
}

// SpeechCreate creates a speech object. Res is 1 on success 0 on error.
func (a *Session) SpeechCreate(engine string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH CREATE %q", engine))
}

// SpeechDeactivateGrammar deactivates a grammar. Res is 1 on success 0 on error.
func (a *Session) SpeechDeactivateGrammar(grammar string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH DEACTIVATE GRAMMAR %q", grammar))
}

// SpeechDestroy destroys a speech object. Res is 1 on success 0 on error.
func (a *Session) SpeechDestroy() (Reply, error) {
	return a.sendMsg("SPEECH DESTROY")
}

// SpeechLoadGrammar loads a grammar. Res is 1 on success 0 on error.
func (a *Session) SpeechLoadGrammar(grammar, path string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH LOAD GRAMMAR %q %q", grammar, path))
}

// SpeechRecognize recognizes speech. Res is 1 onsuccess, 0 in case of error
// In case of success Dat contains a set of different inconsistent values.
// Please refer to res_agi.c in asterisk source code for further info.
func (a *Session) SpeechRecognize(prompt, timeout, offset string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH RECOGNIZE %q %q %q", prompt, timeout, offset))
}

// SpeechSet sets a speech engine setting. Res is 1 on success 0 on error.
func (a *Session) SpeechSet(name, value string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH SET %q %q", name, value))
}

// SpeechUnloadGrammar unloads a grammar. Result is 1 on success 0 on error.
func (a *Session) SpeechUnloadGrammar(grammar string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("SPEECH UNLOAD GRAMMAR %q", grammar))
}

// StreamFile sends audio file on channel. Optional parameter: sample offset for the playback start position.
// Res is 0 if playback completes without a digit being pressed, the ASCII numerical value
// of the digit if one was pressed, or -1 on error or if the channel was disconnected.
// If musiconhold is playing before calling stream file it will be automatically stopped
// and will not be restarted after completion.
func (a *Session) StreamFile(file, escape string, offset ...int) (Reply, error) {
	var r Reply
	var err error
	if offset != nil {
		r, err = a.sendMsg(fmt.Sprintf("STREAM FILE %q %q \"%d\"", file, escape, offset[0]))
	} else {
		r, err = a.sendMsg(fmt.Sprintf("STREAM FILE %q %q", file, escape))
	}
	if r.Dat != "" {
		r.Dat = strings.TrimPrefix(r.Dat, "endpos=")
	}
	return r, err
}

// TddMode toggles TDD mode (for the deaf). Res is 1 if successful, or 0 if channel is not TDD-capable.
func (a *Session) TddMode(mode string) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("TDD MODE %q", mode))
}

// Verbose logs a message to the asterisk verbose log.
// Optional variable: level, the verbose level (1-4). Res is always 1.
func (a *Session) Verbose(msg interface{}, level ...int) (Reply, error) {
	if level != nil {
		return a.sendMsg(fmt.Sprintf("VERBOSE \"%v\" %d", msg, level[0]))
	}
	return a.sendMsg(fmt.Sprintf("VERBOSE \"%v\"", msg))
}

// WaitForDigit waits for a digit to be pressed. Use -1 for the timeout value if you desire
// the call to block indefinitely. Res is -1 on channel failure, 0 if no digit is received
// in the timeout, or the ASCII numerical value of the digit if one is received.
func (a *Session) WaitForDigit(timeout int) (Reply, error) {
	return a.sendMsg(fmt.Sprintf("WAIT FOR DIGIT %d", timeout))
}
