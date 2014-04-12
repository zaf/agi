# agi
--

	To install: go get github.com/zaf/agi

	import "github.com/zaf/agi"

Package agi implements the Aterisk Gateway Interface. All methods return the AGI
Protocol error on failure and on success the AGI result is stored in Res slice,
an element of struct Session. 1st slice element holds the numeric result, 2nd
element the rest of the results if there are any.

## Usage

#### type Session

```go
type Session struct {
	Env map[string]string //AGI environment variables.
	Res []string          //Result of each AGI command.
	Buf *bufio.ReadWriter //AGI I/O buffer.
}
```

Session is a struct holding AGI environment vars, responses and the I/O
handlers.

#### func  Init

```go
func Init(rw *bufio.ReadWriter) (*Session, error)
```
Init initializes a new AGI session. If rw is nil the AGI session will use
standard input (stdin) and standard output (stdout). Returns a pointer to a
Session and the error, if any.

#### func (*Session) Answer

```go
func (a *Session) Answer() error
```
Answer answers channel. Result is -1 on channel failure, or 0 if successful.

#### func (*Session) AsyncagiBreak

```go
func (a *Session) AsyncagiBreak() error
```
AsyncagiBreak interrupts Async AGI. Result is always 0.

#### func (*Session) ChannelStatus

```go
func (a *Session) ChannelStatus(channel string) error
```
ChannelStatus result contains the status of the connected channel. Result
values:

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
func (a *Session) ControlStreamFile(file, escape, skipms, ffchar, rewchr, pausechr string) error
```
ControlStreamFile sends audio file on channel and allows the listener to control
the stream. Result is 0 if playback completes without a digit being pressed, or
the ASCII numerical value of the digit if one was pressed, or -1 on error or if
the channel was disconnected.

#### func (*Session) DatabaseDel

```go
func (a *Session) DatabaseDel(family, key string) error
```
DatabaseDel removes database key/value. Result is 1 if successful, 0 otherwise.

#### func (*Session) DatabaseDelTree

```go
func (a *Session) DatabaseDelTree(family, keytree string) error
```
DatabaseDelTree removes database keytree/value. Result is 1 if successful, 0
otherwise.

#### func (*Session) DatabaseGet

```go
func (a *Session) DatabaseGet(family, key string) error
```
DatabaseGet gets database value. Result is 0 if key is not set, 1 if key is set
and contains the variable in Res[1].

#### func (*Session) DatabasePut

```go
func (a *Session) DatabasePut(family, key, value string) error
```
DatabasePut adds/updates database value. Result is 1 if successful, 0 otherwise.

#### func (*Session) Destroy

```go
func (a *Session) Destroy()
```
Destroy clears an AGI session to help GC.

#### func (*Session) Exec

```go
func (a *Session) Exec(app, options string) error
```
Exec executes a given Application. Result contains whatever the application
returns, or -2 on failure to find application.

#### func (*Session) GetData

```go
func (a *Session) GetData(file, timeout, maxdigits string) error
```
GetData prompts for DTMF on a channel. Result contains the digits received from
the channel at the other end.

#### func (*Session) GetFullVariable

```go
func (a *Session) GetFullVariable(variable, channel string) error
```
GetFullVariable evaluates a channel expression. Result is 1 if variablename is
set and contains the variable in Res[1]. Understands complex variable names and
builtin variables.

#### func (*Session) GetOption

```go
func (a *Session) GetOption(filename, escape, timeout string) error
```
GetOption streams file, prompts for DTMF with timeout. Result contains the
digits received from the channel at the other end and the sample ofset. In case
of failure to playback the result is -1.

#### func (*Session) GetVariable

```go
func (a *Session) GetVariable(variable string) error
```
GetVariable gets a channel variable. Result is 0 if variablename is not set, 1
if variablename is set and contains the variable in Res[1].

#### func (*Session) GoSub

```go
func (a *Session) GoSub(context, extension, priority, args string) error
```
GoSub causes the channel to execute the specified dialplan subroutine, returning
to the dialplan with execution of a Return().

#### func (*Session) Hangup

```go
func (a *Session) Hangup(channel string) error
```
Hangup hangs up a channel, Result is 1 on success, -1 if the given channel was
not found.

#### func (*Session) Noop

```go
func (a *Session) Noop() error
```
Noop does nothing. Result is always 0.

#### func (*Session) ReceiveChar

```go
func (a *Session) ReceiveChar(timeout string) error
```
ReceiveChar receives one character from channels supporting it. Result contains
the decimal value of the character if one is received, or 0 if the channel does
not support text reception. Result is -1 only on error/hangup.

#### func (*Session) ReceiveText

```go
func (a *Session) ReceiveText(timeout string) error
```
ReceiveText receives text from channels supporting it. Result is -1 for failure
or 1 for success, and conatins the string in Res[1].

#### func (*Session) RecordFile

```go
func (a *Session) RecordFile(file, format, escape, timeout, samples, beep, silence string) error
```
RecordFile records to a given file. The format will specify what kind of file
will be recorded. The timeout is the maximum record time in milliseconds, or -1
for no timeout, offset samples is optional, and, if provided, will seek to the
offset without exceeding the end of the file. silence is the number of seconds
of silence allowed before the function returns despite the lack of dtmf digits
or reaching timeout. Negative values of Res[0] mean error, 1 means success. Rest
of the result is a set of different inconsistent values depending on each case,
please refer to res_agi.c in asterisk source code for further info.

#### func (*Session) SayAlpha

```go
func (a *Session) SayAlpha(str, escape string) error
```
SayAlpha says a given character string. Result is 0 if playback completes
without a digit being pressed, the ASCII numerical value of the digit if one was
pressed or -1 on error/hangup.

#### func (*Session) SayDate

```go
func (a *Session) SayDate(date, escape string) error
```
SayDate says a given date. Result is 0 if playback completes without a digit
being pressed, the ASCII numerical value of the digit if one was pressed or -1
on error/hangup.

#### func (*Session) SayDateTime

```go
func (a *Session) SayDateTime(time, escape, format, tz string) error
```
SayDateTime says a given time as specified by the format given. Result is 0 if
playback completes without a digit being pressed, the ASCII numerical value of
the digit if one was pressed or -1 on error/hangup.

#### func (*Session) SayDigits

```go
func (a *Session) SayDigits(number, escape string) error
```
SayDigits says a given digit string. Result is 0 if playback completes without a
digit being pressed, the ASCII numerical value of the digit if one was pressed
or -1 on error/hangup.

#### func (*Session) SayNumber

```go
func (a *Session) SayNumber(num, escape, gender string) error
```
SayNumber says a given number. Result is 0 if playback completes without a digit
being pressed, the ASCII numerical value of the digit if one was pressed or -1
on error/hangup.

#### func (*Session) SayPhonetic

```go
func (a *Session) SayPhonetic(str, escape string) error
```
SayPhonetic says a given character string with phonetics. Result is 0 if
playback completes without a digit pressed, the ASCII numerical value of the
digit if one was pressed, or -1 on error/hangup

#### func (*Session) SayTime

```go
func (a *Session) SayTime(time, escape string) error
```
SayTime says a given time. Result is 0 if playback completes without a digit
being pressed, or the ASCII numerical value of the digit if one was pressed or
-1 on error/hangup.

#### func (*Session) SendImage

```go
func (a *Session) SendImage(image string) error
```
SendImage sends images to channels supporting it. Result is 0 if image is sent,
or if the channel does not support image transmission. Result is -1 only on
error/hangup. Image names should not include extensions.

#### func (*Session) SendText

```go
func (a *Session) SendText(text string) error
```
SendText sends text to channels supporting it. Result is 0 if text is sent, or
if the channel does not support text transmission. Result is -1 only on
error/hangup.

#### func (*Session) SetAutohangup

```go
func (a *Session) SetAutohangup(time string) error
```
SetAutohangup autohangups channel in some time. Setting time to 0 will cause the
autohangup feature to be disabled on this channel. Result is always 0.

#### func (*Session) SetCallerid

```go
func (a *Session) SetCallerid(cid string) error
```
SetCallerid sets callerid for the current channel. Result is always 1.

#### func (*Session) SetContext

```go
func (a *Session) SetContext(context string) error
```
SetContext sets channel context. Result is always 0.

#### func (*Session) SetExtension

```go
func (a *Session) SetExtension(ext string) error
```
SetExtension changes channel extension. Result is always 0.

#### func (*Session) SetMusic

```go
func (a *Session) SetMusic(opt, class string) error
```
SetMusic enables/disables Music on hold generator. If class is not specified,
then the default music on hold class will be used. Result is always 0.

#### func (*Session) SetPriority

```go
func (a *Session) SetPriority(priority string) error
```
SetPriority sets channel dialplan priority. The priority must be a valid
priority or label. Result is always 0.

#### func (*Session) SetVariable

```go
func (a *Session) SetVariable(variable, value string) error
```
SetVariable sets a channel variable. Result is always 1.

#### func (*Session) SpeechActivateGrammar

```go
func (a *Session) SpeechActivateGrammar(grammar string) error
```
SpeechActivateGrammar activates a grammar. Result is 1 on success 0 on error.

#### func (*Session) SpeechCreate

```go
func (a *Session) SpeechCreate(engine string) error
```
SpeechCreate creates a speech object. Result is 1 on success 0 on error.

#### func (*Session) SpeechDeactivateGrammar

```go
func (a *Session) SpeechDeactivateGrammar(grammar string) error
```
SpeechDeactivateGrammar deactivates a grammar. Result is 1 on success 0 on
error.

#### func (*Session) SpeechDestroy

```go
func (a *Session) SpeechDestroy() error
```
SpeechDestroy destroys a speech object. Result is 1 on success 0 on error.

#### func (*Session) SpeechLoadGrammar

```go
func (a *Session) SpeechLoadGrammar(grammar, path string) error
```
SpeechLoadGrammar loads a grammar. Result is 1 on success 0 on error.

#### func (*Session) SpeechRecognize

```go
func (a *Session) SpeechRecognize(prompt, timeout, offset string) error
```
SpeechRecognize recognizes speech. Res[0] is 1 onsuccess, 0 in case of error In
case of success the rest of the result contains a set of different inconsistent
values. Please refer to res_agi.c in asterisk source code for further info.

#### func (*Session) SpeechSet

```go
func (a *Session) SpeechSet(name, value string) error
```
SpeechSet sets a speech engine setting. Result is 1 on success 0 on error.

#### func (*Session) SpeechUnloadGrammar

```go
func (a *Session) SpeechUnloadGrammar(grammar string) error
```
SpeechUnloadGrammar unloads a grammar. Result is 1 on success 0 on error.

#### func (*Session) StreamFile

```go
func (a *Session) StreamFile(file, escape, offset string) error
```
StreamFile sends audio file on channel. Result is 0 if playback completes
without a digit being pressed, the ASCII numerical value of the digit if one was
pressed, or -1 on error or if the channel was disconnected. If musiconhold is
playing before calling stream file it will be automatically stopped and will not
be restarted after completion.

#### func (*Session) TddMode

```go
func (a *Session) TddMode(mode string) error
```
TddMode toggles TDD mode (for the deaf). Result is 1 if successful, or 0 if
channel is not TDD-capable.

#### func (*Session) Verbose

```go
func (a *Session) Verbose(msg, level string) error
```
Verbose logs a message to the asterisk verbose log. level is the verbose level
(1-4). Result is always 1.

#### func (*Session) WaitForDigit

```go
func (a *Session) WaitForDigit(timeout string) error
```
WaitForDigit waits for a digit to be pressed. Use -1 for the timeout value if
you desire the call to block indefinitely. Result is -1 on channel failure, 0 if
no digit is received in the timeout, or the numerical value of the ascii of the
digit if one is received.
