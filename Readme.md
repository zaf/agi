# agi
--
    import "github.com/zaf/agi"

Package agi implements the Asterisk Gateway Interface (http://www.asterisk.org).
All AGI commands are implemented as methods of the Session struct that holds a
copy of the AGI environment variables. All methods return a Reply struct and the
AGI error, if any. The Reply struct contains the numeric result of the AGI
command in Res and if there is any additional data it is stored as string in the
Dat element of the struct. For example, to create a new AGI session and
initialize it:

```go
myAgi := agi.New()
err := myAgi.Init(nil)
if err != nil {
	log.Fatalf("Error Parsing AGI environment: %v\n", err)
}
```

To play back a voice prompt using AGI 'Stream file' command:

```go
rep, err := myAgi.StreamFile("hello-world", "0123456789")
if err != nil {
	log.Fatalf("AGI reply error: %v\n", err)
}
if rep.Res == -1 {
	log.Printf("Error during playback\n")
}
```

For more please see the code snippets in the examples folder.

Package agi implements the Asterisk Gateway Interface (http://www.asterisk.org).

## Usage

#### type Reply

```go
type Reply struct {
	Res int    //Numeric result of the AGI command.
	Dat string //Additional returned data.
}
```

Reply is a struct that holds the return values of each AGI command.

#### type Session

```go
type Session struct {
	Env map[string]string //AGI environment variables.
}
```

Session is a struct holding AGI environment vars and the I/O handlers.

#### func  New

```go
func New() *Session
```
New creates a new Session and returns a pointer to it.

#### func (*Session) Answer

```go
func (a *Session) Answer() (Reply, error)
```
Answer answers channel. Res is -1 on channel failure, or 0 if successful.

#### func (*Session) AsyncagiBreak

```go
func (a *Session) AsyncagiBreak() (Reply, error)
```
AsyncagiBreak interrupts Async AGI. Res is always 0.

#### func (*Session) ChannelStatus

```go
func (a *Session) ChannelStatus(channel ...string) (Reply, error)
```
ChannelStatus Res contains the status of the given channel, if no channel
specified checks the current channel. Result values:

    0 - Channel is down and available.
    1 - Channel is down, but reserved.
    2 - Channel is off hook.
    3 - Digits (or equivalent) have been dialed.
    4 - Line is ringing.
    5 - Remote end is ringing.
    6 - Line is up.
    7 - Line is busy.

#### func (*Session) ControlStreamFile

```go
func (a *Session) ControlStreamFile(file, escape string, params ...interface{}) (Reply, error)
```
ControlStreamFile sends audio file on channel and allows the listener to control
the stream. Optional parameters: skipms, ffchar - Defaults to *, rewchr -
Defaults to #, pausechr. Res is 0 if playback completes without a digit being
pressed, or the ASCII numerical value of the digit if one was pressed, or -1 on
error or if the channel was disconnected.

#### func (*Session) DatabaseDel

```go
func (a *Session) DatabaseDel(family, key string) (Reply, error)
```
DatabaseDel removes database key/value. Res is 1 if successful, 0 otherwise.

#### func (*Session) DatabaseDelTree

```go
func (a *Session) DatabaseDelTree(family string, keytree ...string) (Reply, error)
```
DatabaseDelTree removes database keytree/value. Res is 1 if successful, 0
otherwise.

#### func (*Session) DatabaseGet

```go
func (a *Session) DatabaseGet(family, key string) (Reply, error)
```
DatabaseGet gets database value. Res is 0 if key is not set, 1 if key is set and
the value is returned in Dat.

#### func (*Session) DatabasePut

```go
func (a *Session) DatabasePut(family, key, value string) (Reply, error)
```
DatabasePut adds/updates database value. Res is 1 if successful, 0 otherwise.

#### func (*Session) Exec

```go
func (a *Session) Exec(app, options string) (Reply, error)
```
Exec executes a given application. Res contains whatever the dialplan
application returns, or -2 on failure to find the application.

#### func (*Session) Failure

```go
func (a *Session) Failure() (Reply, error)
```
Failure causes asterisk to terminate the AGI session and set the AGISTATUS
channel variable to 'FAILURE'.

#### func (*Session) GetData

```go
func (a *Session) GetData(file string, params ...int) (Reply, error)
```
GetData prompts for DTMF on a channel. Optional parameters: timeout, maxdigits.
Res contains the digits received from the channel at the other end.

#### func (*Session) GetFullVariable

```go
func (a *Session) GetFullVariable(variable string, channel ...string) (Reply, error)
```
GetFullVariable evaluates a channel expression, if no channel is specified the
current channel is used. Res is 1 if variable is set and the value is returned
in Dat. Understands complex variable names and build in variables.

#### func (*Session) GetOption

```go
func (a *Session) GetOption(filename, escape string, timeout ...int) (Reply, error)
```
GetOption streams file, prompts for DTMF with timeout. Optional parameter:
timeout. Res contains the digits received from the channel at the other end and
Dat contains the sample ofset. In case of failure to playback Res is -1.

#### func (*Session) GetVariable

```go
func (a *Session) GetVariable(variable string) (Reply, error)
```
GetVariable gets a channel variable. Res is 0 if variable is not set, 1 if
variable is set and Dat contains the value.

#### func (*Session) GoSub

```go
func (a *Session) GoSub(context, extension, priority, args string) (Reply, error)
```
GoSub causes the channel to execute the specified dialplan subroutine, returning
to the dialplan with execution of a Return().

#### func (*Session) Hangup

```go
func (a *Session) Hangup(channel ...string) (Reply, error)
```
Hangup hangs up a channel, Res is 1 on success, -1 if the given channel was not
found.

#### func (*Session) Init

```go
func (a *Session) Init(rw *bufio.ReadWriter) error
```
Init initializes a new AGI session. If rw is nil the AGI session will use
standard input (stdin) and output (stdout) for a standalone AGI application. It
reads and stores the AGI environment variables in Env. Returns an error if the
parsing of the AGI environment was unsuccessful.

#### func (*Session) Noop

```go
func (a *Session) Noop(params ...interface{}) (Reply, error)
```
Noop does nothing. Res is always 0.

#### func (*Session) RawCommand

```go
func (a *Session) RawCommand(params ...interface{}) (Reply, error)
```
RawCommand sends a user defined command. Use of this is generally discouraged.
Useful only for debugging, testing and maybe compatibility with newer/altered
versions of the AGI protocol.

#### func (*Session) ReceiveChar

```go
func (a *Session) ReceiveChar(timeout int) (Reply, error)
```
ReceiveChar receives one character from channels supporting it. Res contains the
decimal value of the character if one is received, or 0 if the channel does not
support text reception. Result is -1 only on error/hang-up.

#### func (*Session) ReceiveText

```go
func (a *Session) ReceiveText(timeout int) (Reply, error)
```
ReceiveText receives text from channels supporting it. Res is -1 for failure or
1 for success, and Dat contains the string.

#### func (*Session) RecordFile

```go
func (a *Session) RecordFile(file, format, escape string, timeout int, params ...interface{}) (Reply, error)
```
RecordFile records to a given file. The format will specify what kind of file
will be recorded. The timeout is the maximum record time in milliseconds, or -1
for no timeout. Optional parameters: offset samples if provided will seek to the
offset without exceeding the end of the file. Silence (always prefixed with
s=)is the number of seconds of silence allowed before the function returns
despite the lack of DTMF digits or reaching timeout. Negative values of Res mean
error. Res equals 0 if recording was successful without any DTMF digit received,
or the ASCII numerical value of the digit if one was received. Dat contains a
set of different inconsistent return values depending on each case, please refer
to res_agi.c in asterisk source code for further info.

#### func (*Session) SayAlpha

```go
func (a *Session) SayAlpha(str, escape string) (Reply, error)
```
SayAlpha says a given character string. Res is 0 if playback completes without a
digit being pressed, the ASCII numerical value of the digit if one was pressed
or -1 on error/hang-up.

#### func (*Session) SayDate

```go
func (a *Session) SayDate(date int64, escape string) (Reply, error)
```
SayDate says a given date (Unix time format). Res is 0 if playback completes
without a digit being pressed, the ASCII numerical value of the digit if one was
pressed or -1 on error/hang-up.

#### func (*Session) SayDateTime

```go
func (a *Session) SayDateTime(time int64, escape string, params ...string) (Reply, error)
```
SayDateTime says a given time (Unix time format). Optional parameters: format,
the format the time should be said in. See voicemail.conf (defaults to ABdY
'digits/at' IMp). timezone, acceptable values can be found in
/usr/share/zoneinfo. Defaults to machine default. Res is 0 if playback completes
without a digit being pressed, the ASCII numerical value of the digit if one was
pressed or -1 on error/hang-up.

#### func (*Session) SayDigits

```go
func (a *Session) SayDigits(digit int, escape string) (Reply, error)
```
SayDigits says a given digit. Res is 0 if playback completes without a digit
being pressed, the ASCII numerical value of the digit if one was pressed or -1
on error/hang-up.

#### func (*Session) SayNumber

```go
func (a *Session) SayNumber(num int, escape string, gender ...string) (Reply, error)
```
SayNumber says a given number. Optional parameter gender. Res is 0 if playback
completes without a digit being pressed, the ASCII numerical value of the digit
if one was pressed or -1 on error/hang-up.

#### func (*Session) SayPhonetic

```go
func (a *Session) SayPhonetic(str, escape string) (Reply, error)
```
SayPhonetic says a given character string with phonetics. Res is 0 if playback
completes without a digit pressed, the ASCII numerical value of the digit if one
was pressed, or -1 on error/hang-up

#### func (*Session) SayTime

```go
func (a *Session) SayTime(time int64, escape string) (Reply, error)
```
SayTime says a given time (Unix time format). Res is 0 if playback completes
without a digit being pressed, or the ASCII numerical value of the digit if one
was pressed or -1 on error/hang-up.

#### func (*Session) SendImage

```go
func (a *Session) SendImage(image string) (Reply, error)
```
SendImage sends images to channels supporting it. Res is 0 if image is sent, or
if the channel does not support image transmission. Result is -1 only on
error/hang-up. Image names should not include extensions.

#### func (*Session) SendText

```go
func (a *Session) SendText(text string) (Reply, error)
```
SendText sends text to channels supporting it. Res is 0 if text is sent, or if
the channel does not support text transmission. Result is -1 only on
error/hang-up.

#### func (*Session) SetAutohangup

```go
func (a *Session) SetAutohangup(time int) (Reply, error)
```
SetAutohangup autohang-ups channel after a number of seconds. Setting time to 0
will cause the autohang-up feature to be disabled on this channel. Res is always
0.

#### func (*Session) SetCallerid

```go
func (a *Session) SetCallerid(cid string) (Reply, error)
```
SetCallerid sets callerid for the current channel. Res is always 1.

#### func (*Session) SetContext

```go
func (a *Session) SetContext(context string) (Reply, error)
```
SetContext sets channel context. Res is always 0.

#### func (*Session) SetExtension

```go
func (a *Session) SetExtension(ext string) (Reply, error)
```
SetExtension changes channel extension. Res is always 0.

#### func (*Session) SetMusic

```go
func (a *Session) SetMusic(opt string, class ...string) (Reply, error)
```
SetMusic enables/disables Music on hold generator by setting opt to "on" or
"off". Optional parameter: class, if not specified, then the default music on
hold class will be used. Res is always 0.

#### func (*Session) SetPriority

```go
func (a *Session) SetPriority(priority string) (Reply, error)
```
SetPriority sets channel dialplan priority. The priority must be a valid
priority or label. Res is always 0.

#### func (*Session) SetVariable

```go
func (a *Session) SetVariable(variable string, value interface{}) (Reply, error)
```
SetVariable sets a channel variable. Res is always 1.

#### func (*Session) SpeechActivateGrammar

```go
func (a *Session) SpeechActivateGrammar(grammar string) (Reply, error)
```
SpeechActivateGrammar activates a grammar. Res is 1 on success 0 on error.

#### func (*Session) SpeechCreate

```go
func (a *Session) SpeechCreate(engine string) (Reply, error)
```
SpeechCreate creates a speech object. Res is 1 on success 0 on error.

#### func (*Session) SpeechDeactivateGrammar

```go
func (a *Session) SpeechDeactivateGrammar(grammar string) (Reply, error)
```
SpeechDeactivateGrammar deactivates a grammar. Res is 1 on success 0 on error.

#### func (*Session) SpeechDestroy

```go
func (a *Session) SpeechDestroy() (Reply, error)
```
SpeechDestroy destroys a speech object. Res is 1 on success 0 on error.

#### func (*Session) SpeechLoadGrammar

```go
func (a *Session) SpeechLoadGrammar(grammar, path string) (Reply, error)
```
SpeechLoadGrammar loads a grammar. Res is 1 on success 0 on error.

#### func (*Session) SpeechRecognize

```go
func (a *Session) SpeechRecognize(prompt, timeout, offset string) (Reply, error)
```
SpeechRecognize recognizes speech. Res is 1 onsuccess, 0 in case of error In
case of success Dat contains a set of different inconsistent values. Please
refer to res_agi.c in asterisk source code for further info.

#### func (*Session) SpeechSet

```go
func (a *Session) SpeechSet(name, value string) (Reply, error)
```
SpeechSet sets a speech engine setting. Res is 1 on success 0 on error.

#### func (*Session) SpeechUnloadGrammar

```go
func (a *Session) SpeechUnloadGrammar(grammar string) (Reply, error)
```
SpeechUnloadGrammar unloads a grammar. Result is 1 on success 0 on error.

#### func (*Session) StreamFile

```go
func (a *Session) StreamFile(file, escape string, offset ...int) (Reply, error)
```
StreamFile sends audio file on channel. Optional parameter: sample offset for
the playback start position. Res is 0 if playback completes without a digit
being pressed, the ASCII numerical value of the digit if one was pressed, or -1
on error or if the channel was disconnected. If musiconhold is playing before
calling stream file it will be automatically stopped and will not be restarted
after completion.

#### func (*Session) TddMode

```go
func (a *Session) TddMode(mode string) (Reply, error)
```
TddMode toggles TDD mode (for the deaf). Res is 1 if successful, or 0 if channel
is not TDD-capable.

#### func (*Session) Verbose

```go
func (a *Session) Verbose(msg interface{}, level ...int) (Reply, error)
```
Verbose logs a message to the asterisk verbose log. Optional variable: level,
the verbose level (1-4). Res is always 1.

#### func (*Session) WaitForDigit

```go
func (a *Session) WaitForDigit(timeout int) (Reply, error)
```
WaitForDigit waits for a digit to be pressed. Use -1 for the timeout value if
you desire the call to block indefinitely. Res is -1 on channel failure, 0 if no
digit is received in the timeout, or the ASCII numerical value of the digit if
one is received.
