// Copyright (C) 2013 - 2014, Lefteris Zafiris <zaf.000@gmail.com>
// This program is free software, distributed under the terms of
// the BSD 3-Clause License. See the LICENSE file
// at the top of the source tree.

// Package agi implements the Asterisk Gateway Interface. All methods return the AGI Protocol error on failure
// and on success the AGI result is stored in Res slice, an element of struct Session.
// 1st slice element holds the numeric result, 2nd element the rest of the results if there are any.
package agi

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	envMin = 20  // Minimun number of AGI environment args
	envMax = 150 // Maximum number of AGI environment args
)

// Session is a struct holding AGI environment vars, responses and the I/O handlers.
type Session struct {
	Env map[string]string //AGI environment variables.
	Res []string          //Result of each AGI command.
	Buf *bufio.ReadWriter //AGI I/O buffer.
}

// Init initializes a new AGI session. If rw is nil the AGI session will use standard input (stdin)
// and standard output (stdout). Returns a pointer to a Session and the error, if any.
func Init(rw *bufio.ReadWriter) (*Session, error) {
	a := new(Session)
	if rw == nil {
		a.Buf = bufio.NewReadWriter(bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout))
	} else {
		a.Buf = rw
	}
	err := a.parseEnv()
	return a, err
}

// Answer answers channel. Result is -1 on channel failure, or 0 if successful.
func (a *Session) Answer() error {
	return a.sendMsg(fmt.Sprintf("ANSWER"))
}

// AsyncagiBreak interrupts Async AGI. Result is always 0.
func (a *Session) AsyncagiBreak() error {
	return a.sendMsg(fmt.Sprintf("ASYNCAGI BREAK"))
}

//ChannelStatus result contains the status of the given channel, if no channel specified
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
func (a *Session) ChannelStatus(channel ...string) error {
	if len(channel) > 0 {
		return a.sendMsg(fmt.Sprintf("CHANNEL STATUS %s", channel[0]))
	}
	return a.sendMsg(fmt.Sprintf("CHANNEL STATUS"))
}

// ControlStreamFile sends audio file on channel and allows the listener to control the stream.
// Optional parameters: skipms, ffchar - Defaults to *, rewchr - Defaults to #, pausechr.
// Result is 0 if playback completes without a digit being pressed, or the ASCII numerical value
// of the digit if one was pressed, or -1 on error or if the channel was disconnected.
func (a *Session) ControlStreamFile(file, escape string, params ...interface{}) error {
	cmd := fmt.Sprintf("%s \"%s\"", file, escape)
	for _, par := range params {
		cmd = fmt.Sprintf("%s %v", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("CONTROL STREAM FILE %s", cmd))
}

// DatabaseDel removes database key/value. Result is 1 if successful, 0 otherwise.
func (a *Session) DatabaseDel(family, key string) error {
	return a.sendMsg(fmt.Sprintf("DATABASE DEL %s %s", family, key))
}

// DatabaseDelTree removes database keytree/value. Result is 1 if successful, 0 otherwise.
func (a *Session) DatabaseDelTree(family, keytree string) error {
	return a.sendMsg(fmt.Sprintf("DATABASE DELTREE %s %s", family, keytree))
}

// DatabaseGet gets database value. Result is 0 if key is not set, 1 if key is set
// and contains the variable in Res[1].
func (a *Session) DatabaseGet(family, key string) error {
	err := a.sendMsg(fmt.Sprintf("DATABASE GET %s %s", family, key))
	if len(a.Res) > 1 {
		a.Res[1] = stripPar(a.Res[1])
	}
	return err
}

// DatabasePut adds/updates database value. Result is 1 if successful, 0 otherwise.
func (a *Session) DatabasePut(family, key, value string) error {
	return a.sendMsg(fmt.Sprintf("DATABASE PUT %s %s %s", family, key, value))
}

// Exec executes a given Application. Result contains whatever the application returns,
// or -2 on failure to find application.
func (a *Session) Exec(app, options string) error {
	return a.sendMsg(fmt.Sprintf("EXEC %s %s", app, options))
}

// GetData prompts for DTMF on a channel. Optional parameters: timeout, maxdigits.
// Result contains the digits received from the channel at the other end.
func (a *Session) GetData(file string, params ...int) error {
	cmd := file
	for _, par := range params {
		cmd = fmt.Sprintf("%s %d", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("GET DATA %s", cmd))
}

// GetFullVariable evaluates a channel expression, if no channel is specified the current channel is used.
// Result is 1 if variablename is set and contains the variable in Res[1].
// Understands complex variable names and builtin variables.
func (a *Session) GetFullVariable(variable string, channel ...string) error {
	var err error
	if len(channel) > 0 {
		err = a.sendMsg(fmt.Sprintf("GET FULL VARIABLE %s %s", variable, channel[0]))
	} else {
		err = a.sendMsg(fmt.Sprintf("GET FULL VARIABLE %s", variable))
	}
	if len(a.Res) > 1 {
		a.Res[1] = stripPar(a.Res[1])
	}
	return err
}

// GetOption streams file, prompts for DTMF with timeout. Optiona parameter: timeout.
// Result contains the digits received from the channel at the other end and the sample ofset.
// In case of failure to playback the result is -1.
func (a *Session) GetOption(filename, escape string, timeout ...int) error {
	var err error
	if len(timeout) > 0 {
		err = a.sendMsg(fmt.Sprintf("GET OPTION %s \"%s\" %d", filename, escape, timeout[0]))
	} else {
		err = a.sendMsg(fmt.Sprintf("GET OPTION %s \"%s\"", filename, escape))
	}
	if len(a.Res) > 1 {
		a.Res[1] = strings.TrimPrefix(a.Res[1], "endpos=")
	}
	return err
}

// GetVariable gets a channel variable. Result is 0 if variablename is not set,
// 1 if variablename is set and contains the variable in Res[1].
func (a *Session) GetVariable(variable string) error {
	err := a.sendMsg(fmt.Sprintf("GET VARIABLE %s", variable))
	if len(a.Res) > 1 {
		a.Res[1] = stripPar(a.Res[1])
	}
	return err
}

// GoSub causes the channel to execute the specified dialplan subroutine, returning to the dialplan
// with execution of a Return().
func (a *Session) GoSub(context, extension, priority, args string) error {
	return a.sendMsg(fmt.Sprintf("GOSUB %s %s %s %s", context, extension, priority, args))
}

// Hangup hangs up a channel, Result is 1 on success, -1 if the given channel was not found.
func (a *Session) Hangup(channel ...string) error {
	if len(channel) > 0 {
		return a.sendMsg(fmt.Sprintf("HANGUP %s", channel[0]))
	}
	return a.sendMsg(fmt.Sprintf("HANGUP"))
}

// Noop does nothing. Result is always 0.
func (a *Session) Noop(params ...interface{}) error {
	var cmd string
	for _, par := range params {
		cmd = fmt.Sprintf("%s %v", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("NOOP%s", cmd))
}

// ReceiveChar receives one character from channels supporting it. Result contains the decimal value of
// the character if one is received, or 0 if the channel does not support text reception.
// Result is -1 only on error/hangup.
func (a *Session) ReceiveChar(timeout int) error {
	return a.sendMsg(fmt.Sprintf("RECEIVE CHAR %d", timeout))
}

// ReceiveText receives text from channels supporting it. Result is -1 for failure
// or 1 for success, and conatins the string in Res[1].
func (a *Session) ReceiveText(timeout int) error {
	err := a.sendMsg(fmt.Sprintf("RECEIVE TEXT %d", timeout))
	if len(a.Res) > 1 {
		a.Res[1] = stripPar(a.Res[1])
	}
	return err
}

// RecordFile records to a given file. The format will specify what kind of file will be recorded.
// The timeout is the maximum record time in milliseconds, or -1 for no timeout.
// Optional parameters: offset samples if provided will seek to the offset without exceeding
// the end of the file. Silence (always prefixed with s=)is the number of seconds of silence allowed
// before the function returns despite the lack of dtmf digits or reaching timeout.
// Negative values of Res[0] mean error, 1 means success.
// Rest of the result is a set of different inconsistent values depending
// on each case, please refer to res_agi.c in asterisk source code for further info.
func (a *Session) RecordFile(file, format, escape string, timeout int, params ...interface{}) error {
	cmd := fmt.Sprintf("%s %s \"%s\" %d", file, format, escape, timeout)
	for _, par := range params {
		cmd = fmt.Sprintf("%s %v", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("RECORD FILE %s", cmd))
}

// SayAlpha says a given character string. Result is 0 if playback completes without a digit
// being pressed, the ASCII numerical value of the digit if one was pressed or -1 on error/hangup.
func (a *Session) SayAlpha(str, escape string) error {
	return a.sendMsg(fmt.Sprintf("SAY ALPHA %s \"%s\"", str, escape))
}

// SayDate says a given date (Unix time format). Result is 0 if playback completes without a digit
// being pressed, the ASCII numerical value of the digit if one was pressed or -1 on error/hangup.
func (a *Session) SayDate(date int64, escape string) error {
	return a.sendMsg(fmt.Sprintf("SAY DATE %d \"%s\"", date, escape))
}

// SayDateTime says a given time (Unix time format). Optional parameters:
// fomrat, the format the time should be said in. See voicemail.conf (defaults to ABdY 'digits/at' IMp).
// timezone, acceptable values can be found in /usr/share/zoneinfo. Defaults to machine default.
// Result is 0 if playback completes without a digit being pressed, the ASCII numerical
// value of the digit if one was pressed or -1 on error/hangup.
func (a *Session) SayDateTime(time int64, escape string, params ...string) error {
	cmd := fmt.Sprintf("%d \"%s\"", time, escape)
	for _, par := range params {
		cmd = fmt.Sprintf("%s %v", cmd, par)
	}
	return a.sendMsg(fmt.Sprintf("SAY DATETIME %s", cmd))
}

// SayDigits says a given digit. Result is 0 if playback completes without a digit being pressed,
// the ASCII numerical value of the digit if one was pressed or -1 on error/hangup.
func (a *Session) SayDigits(digit int, escape string) error {
	return a.sendMsg(fmt.Sprintf("SAY DIGITS %d \"%s\"", digit, escape))
}

// SayNumber says a given number. Optional parameter gender. Result is 0 if playback completes
// without a digit being pressed, the ASCII numerical value of the digit if one was pressed or -1 on error/hangup.
func (a *Session) SayNumber(num int, escape string, gender ...string) error {
	if len(gender) > 0 {
		return a.sendMsg(fmt.Sprintf("SAY NUMBER %d \"%s\" %s", num, escape, gender[0]))
	}
	return a.sendMsg(fmt.Sprintf("SAY NUMBER %d \"%s\"", num, escape))
}

// SayPhonetic says a given character string with phonetics. Result is 0 if playback completes
// without a digit pressed, the ASCII numerical value of the digit if one was pressed, or -1 on error/hangup
func (a *Session) SayPhonetic(str, escape string) error {
	return a.sendMsg(fmt.Sprintf("SAY PHONETIC %s \"%s\"", str, escape))
}

// SayTime says a given time (Unix time format). Result is 0 if playback completes without a digit
// being pressed, or the ASCII numerical value of the digit if one was pressed or -1 on error/hangup.
func (a *Session) SayTime(time int64, escape string) error {
	return a.sendMsg(fmt.Sprintf("SAY TIME %d \"%s\"", time, escape))
}

// SendImage sends images to channels supporting it. Result is 0 if image is sent, or if the channel
// does not support image transmission. Result is -1 only on error/hangup. Image names should not include extensions.
func (a *Session) SendImage(image string) error {
	return a.sendMsg(fmt.Sprintf("SEND IMAGE %s", image))
}

// SendText sends text to channels supporting it. Result is 0 if text is sent, or if the channel
// does not support text transmission. Result is -1 only on error/hangup.
func (a *Session) SendText(text string) error {
	return a.sendMsg(fmt.Sprintf("SEND TEXT \"%s\"", text))
}

// SetAutohangup autohangups channel after a number of seconds. Setting time to 0 will cause the autohangup
// feature to be disabled on this channel. Result is always 0.
func (a *Session) SetAutohangup(time int) error {
	return a.sendMsg(fmt.Sprintf("SET AUTOHANGUP %d", time))
}

// SetCallerid sets callerid for the current channel. Result is always 1.
func (a *Session) SetCallerid(cid string) error {
	return a.sendMsg(fmt.Sprintf("SET CALLERID %s", cid))
}

// SetContext sets channel context. Result is always 0.
func (a *Session) SetContext(context string) error {
	return a.sendMsg(fmt.Sprintf("SET CONTEXT %s", context))
}

// SetExtension changes channel extension. Result is always 0.
func (a *Session) SetExtension(ext string) error {
	return a.sendMsg(fmt.Sprintf("SET EXTENSION %s", ext))
}

// SetMusic enables/disables Music on hold generator by settong opt to "on" or "off".
// Optional parameter: class, if not specified, then the default music on hold class will be used.
// Result is always 0.
func (a *Session) SetMusic(opt string, class ...string) error {
	if len(class) > 0 {
		return a.sendMsg(fmt.Sprintf("SET MUSIC %s %s", opt, class[0]))
	}
	return a.sendMsg(fmt.Sprintf("SET MUSIC %s", opt))
}

// SetPriority sets channel dialplan priority. The priority must be a valid priority or label.
// Result is always 0.
func (a *Session) SetPriority(priority string) error {
	return a.sendMsg(fmt.Sprintf("SET PRIORITY %s", priority))
}

// SetVariable sets a channel variable. Result is always 1.
func (a *Session) SetVariable(variable string, value interface{}) error {
	return a.sendMsg(fmt.Sprintf("SET VARIABLE \"%s\" \"%v\"", variable, value))
}

// SpeechActivateGrammar activates a grammar. Result is 1 on success 0 on error.
func (a *Session) SpeechActivateGrammar(grammar string) error {
	return a.sendMsg(fmt.Sprintf("SPEECH ACTIVATE GRAMMAR %s", grammar))
}

// SpeechCreate creates a speech object. Result is 1 on success 0 on error.
func (a *Session) SpeechCreate(engine string) error {
	return a.sendMsg(fmt.Sprintf("SPEECH CREATE %s", engine))
}

// SpeechDeactivateGrammar deactivates a grammar. Result is 1 on success 0 on error.
func (a *Session) SpeechDeactivateGrammar(grammar string) error {
	return a.sendMsg(fmt.Sprintf("SPEECH DEACTIVATE GRAMMAR %s", grammar))
}

// SpeechDestroy destroys a speech object. Result is 1 on success 0 on error.
func (a *Session) SpeechDestroy() error {
	return a.sendMsg(fmt.Sprintf("SPEECH DESTROY"))
}

// SpeechLoadGrammar loads a grammar. Result is 1 on success 0 on error.
func (a *Session) SpeechLoadGrammar(grammar, path string) error {
	return a.sendMsg(fmt.Sprintf("SPEECH LOAD GRAMMAR %s %s", grammar, path))
}

// SpeechRecognize recognizes speech. Res[0] is 1 onsuccess, 0 in case of error
// In case of success the rest of the result contains a set of different inconsistent values.
// Please refer to res_agi.c in asterisk source code for further info.
func (a *Session) SpeechRecognize(prompt, timeout, offset string) error {
	return a.sendMsg(fmt.Sprintf("SPEECH RECOGNIZE %s %s %s", prompt, timeout, offset))
}

// SpeechSet sets a speech engine setting. Result is 1 on success 0 on error.
func (a *Session) SpeechSet(name, value string) error {
	return a.sendMsg(fmt.Sprintf("SPEECH SET %s %s", name, value))
}

// SpeechUnloadGrammar unloads a grammar. Result is 1 on success 0 on error.
func (a *Session) SpeechUnloadGrammar(grammar string) error {
	return a.sendMsg(fmt.Sprintf("SPEECH UNLOAD GRAMMAR %s", grammar))
}

// StreamFile sends audio file on channel. Optional parameter: sample offset for the playback start position.
// Result is 0 if playback completes without a digit being pressed, the ASCII numerical value
// of the digit if one was pressed, or -1 on error or if the channel was disconnected.
// If musiconhold is playing before calling stream file it will be automatically stopped
// and will not be restarted after completion.
func (a *Session) StreamFile(file, escape string, offset ...int) error {
	var err error
	if len(offset) > 0 {
		err = a.sendMsg(fmt.Sprintf("STREAM FILE %s \"%s\" %d", file, escape, offset[0]))
	} else {
		err = a.sendMsg(fmt.Sprintf("STREAM FILE %s \"%s\"", file, escape))
	}
	if len(a.Res) > 1 {
		a.Res[1] = strings.TrimPrefix(a.Res[1], "endpos=")
	}
	return err
}

// TddMode toggles TDD mode (for the deaf). Result is 1 if successful, or 0 if channel is not TDD-capable.
func (a *Session) TddMode(mode string) error {
	return a.sendMsg(fmt.Sprintf("TDD MODE %s", mode))
}

// Verbose logs a message to the asterisk verbose log.
// Optional variable: level, the verbose level (1-4). Result is always 1.
func (a *Session) Verbose(msg string, level ...int) error {
	if len(level) > 0 {
		return a.sendMsg(fmt.Sprintf("VERBOSE \"%s\" %d", msg, level[0]))
	}
	return a.sendMsg(fmt.Sprintf("VERBOSE \"%s\"", msg))
}

// WaitForDigit waits for a digit to be pressed. Use -1 for the timeout value if you desire
// the call to block indefinitely. Result is -1 on channel failure, 0 if no digit is received
// in the timeout, or the numerical value of the ascii of the digit if one is received.
func (a *Session) WaitForDigit(timeout int) error {
	return a.sendMsg(fmt.Sprintf("WAIT FOR DIGIT %d", timeout))
}

// Destroy clears an AGI session to help GC.
func (a *Session) Destroy() {
	a = nil
}

// sendMsg sends an AGI command and returns the result
func (a *Session) sendMsg(s string) error {
	s = strings.TrimSpace(s)
	fmt.Fprintln(a.Buf, s)
	a.Buf.Flush()
	return a.parseResponse()
}

// parseEnv reads and stores AGI environment.
func (a *Session) parseEnv() error {
	var err error
	a.Env = make(map[string]string)
	for i := 0; i <= envMax; i++ {
		line, err := a.Buf.ReadString('\n')
		if err != nil || line == "\n" {
			break
		}
		inputStr := strings.SplitN(line, ": ", 2)
		if len(inputStr) == 2 {
			inputStr[0] = strings.TrimPrefix(inputStr[0], "agi_")
			inputStr[1] = strings.TrimRight(inputStr[1], "\n")
			a.Env[inputStr[0]] = inputStr[1]
		} else {
			err = fmt.Errorf("erroneous environment input: %v", inputStr)
			a.Env = nil
			return err
		}
	}
	if len(a.Env) < envMin {
		err = fmt.Errorf("incomplete environment with %d env vars", len(a.Env))
		a.Env = nil
	}
	return err
}

// parseResponse reads back and parses AGI repsonse. Returns the protocol error, if any, and the
// AGI response.
func (a *Session) parseResponse() error {
	var err error
	a.Res = nil
	line, _ := a.Buf.ReadString('\n')
	line = strings.TrimRight(line, "\n")
	reply := strings.SplitN(line, " ", 3)
	if len(reply) < 2 {
		if reply[0] == "HANGUP" {
			return fmt.Errorf("client sent a HANGUP request")
		}
		return fmt.Errorf("erroneous or partial agi response: %v", reply)
	}
	switch reply[0] {
	case "200":
		a.Res = reply[1:]
		a.Res[0] = strings.TrimPrefix(a.Res[0], "result=")
	case "510":
		err = fmt.Errorf("invalid or unknown command")
	case "511":
		err = fmt.Errorf("command not permitted on a dead channel")
	case "520":
		err = fmt.Errorf("invalid command syntax")
	case "520-Invalid":
		err = fmt.Errorf("invalid command syntax")
		a.Buf.ReadString('\n')
	default:
		err = fmt.Errorf("erroneous agi response: %v", reply)
	}
	return err
}

// stripPar removes leading and trailing parentheses.
func stripPar(s string) string {
	s = strings.TrimLeft(s, "(")
	s = strings.TrimRight(s, ")")
	return s
}
