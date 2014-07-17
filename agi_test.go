// Copyright (C) 2013 - 2014, Lefteris Zafiris <zaf.000@gmail.com>
// This program is free software, distributed under the terms of
// the BSD 3-Clause License. See the LICENSE file
// at the top of the source tree.

package agi

import (
	"bufio"
	"bytes"
	"io"
	"testing"
)

// AGI environment data
var env = []byte(`agi_network: yes
agi_network_script: foo?
agi_request: agi://127.0.0.1/foo?
agi_channel: SIP/1234-00000000
agi_language: en
agi_type: SIP
agi_uniqueid: 1397044468.0
agi_version: 0.1
agi_callerid: 1001
agi_calleridname: 1001
agi_callingpres: 67
agi_callingani2: 0
agi_callington: 0
agi_callingtns: 0
agi_dnid: 123456
agi_rdnis: unknown
agi_context: default
agi_extension: 123456
agi_priority: 1
agi_enhanced: 0.0
agi_accountcode: 0
agi_threadid: -1289290944
agi_arg_1: argument1
agi_arg_2: argument 2
agi_arg_3: 3

`)

// AGI Responses
var rep = []byte(`200 result=1
200 result=1 (speech) endpos=1234 results=foo bar
510 Invalid or unknown command
511 Command Not Permitted on a dead channel
520 Invalid command syntax.  Proper usage not available.
520-Invalid command syntax.  Proper usage follows:
Answers channel if not already in answer state. Returns -1 on channel failure, or 0 if successful.
200


some random reply that we are not supposed to get
HANGUP
`)

// Test AGI environment parsing
func TestAgiEnv(t *testing.T) {
	a := New()
	a.buf = bufio.NewReadWriter(
		bufio.NewReader(bytes.NewReader(env)),
		nil,
	)
	err := a.parseEnv()
	if err != nil {
		t.Errorf("parseEnv failed: %v", err)
	}
	if len(a.Env) != 25 {
		t.Errorf("Error parsing complete AGI environment var list. Expected length: 25, reported: %d", len(a.Env))
	}
	if a.Env["arg_1"] != "argument1" {
		t.Errorf("Error parsing arg1. Expecting: argument1, got: %s", a.Env["arg_1"])
	}
	if a.Env["arg_2"] != "argument 2" {
		t.Errorf("Error parsing arg2. Expecting: argument 2, got: %s", a.Env["arg_2"])
	}
	if a.Env["arg_3"] != "3" {
		t.Errorf("Error parsing arg3. Expecting: 3, got: %s", a.Env["arg_3"])
	}
}

// Test AGI repsonse parsing
func TestRes(t *testing.T) {
	a := New()
	a.buf = bufio.NewReadWriter(
		bufio.NewReader(bytes.NewReader(rep)),
		nil,
	)
	r, err := a.parseResponse()
	if err != nil {
		t.Errorf("Error parsing AGI 200 response: %v", err)
	}
	if r.Dat != "" {
		t.Errorf("Error parsing AGI 200 response. Got unexpected data: %d", r.Dat)
	}
	if r.Res != 1 {
		t.Errorf("Error parsing AGI 200 response. Expecting: 1, got: %d", r.Res)
	}
	r, err = a.parseResponse()
	if r.Dat != "(speech) endpos=1234 results=foo bar" {
		t.Errorf("Error parsing AGI complex 200 response: %v", r)
	}

	_, err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing AGI 510 response.")
	}
	_, err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing AGI 511 response.")
	}
	_, err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing AGI 520 response.")
	}
	_, err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing AGI 520 response containing usage details.")
	}
	_, err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing a partial AGI response.")
	}
	_, err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing an empty AGI response.")
	}
	_, err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing an empty AGI response.")
	}
	_, err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing an erroneous AGI response.")
	}
	_, err = a.parseResponse()
	if err == nil {
		t.Error("Failed to detect a HANGUP reguest.")
	}
}

// Test the generation of AGI commands
func TestCmd(t *testing.T) {
	var r Reply
	var b []byte
	buf := bytes.NewBuffer(b)
	a := New()
	data := append(env, "200 result=1 endpos=1234\n"...)
	err := a.Init(
		bufio.NewReadWriter(
			bufio.NewReader(bytes.NewReader(data)),
			bufio.NewWriter(io.Writer(buf)),
		),
	)

	if err != nil {
		t.Errorf("Failed to initialize new AGI session: %v", err)
	}
	r, err = a.GetOption("echo", "any")
	if err != nil {
		t.Errorf("Failed to parse AGI responce: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("Failed to send AGI command")
	}
	str, _ := buf.ReadString(10)
	if str != "GET OPTION echo \"any\"\n" {
		t.Errorf("Failed to sent properly formatted AGI command: %s", str)
	}
	if r.Res != 1 {
		t.Errorf("Failed to get the right numeric result. Expecting: 1, got: %d", r.Res)
	}
	if r.Dat != "1234" {
		t.Errorf("Failed to properly parse the rest of the response. Expecting: 1234, got: %s", r.Dat)
	}
}

// Benchmark AGI session initialisation
func BenchmarkParseEnv(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a := New()
		a.Init(
			bufio.NewReadWriter(
				bufio.NewReader(bytes.NewReader(env)),
				nil,
			),
		)
	}
}

// Benchmark AGI response parsing
func BenchmarkParseRes(b *testing.B) {
	a := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.buf = bufio.NewReadWriter(
			bufio.NewReader(bytes.NewReader(rep)),
			nil,
		)
		for k := 0; k < 10; k++ {
			a.parseResponse()
		}
	}
}
