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

type writeConn struct {
	buf []byte
}

func (c *writeConn) Write(p []byte) (int, error) {
	c.buf = append(c.buf, p...)
	return len(p), nil
}

// Test AGI environment parsing
func TestAgiEnv(t *testing.T) {
	var a Session
	a.Buf = bufio.NewReadWriter(
		bufio.NewReader(bytes.NewReader(genEnv())),
		nil,
	)
	err := a.parseEnv()
	if err != nil {
		t.Error("parseEnv failed")
	}
	if len(a.Env) != 25 {
		t.Error("Error parsing complete AGI environment var list.")
	}
	if a.Env["arg_1"] != "foo" {
		t.Error("Error parsing arg1")
	}
	if a.Env["arg_2"] != "bar" {
		t.Error("Error parsing arg2")
	}
	if a.Env["arg_3"] != "roo" {
		t.Error("Error parsing arg3")
	}
}

// Test AGI repsonse parsing
func TestRes(t *testing.T) {
	var a Session
	data := genRes()
	a.Buf = bufio.NewReadWriter(
		bufio.NewReader(bytes.NewReader(data)),
		nil,
	)
	err := a.parseResponse()
	if err != nil {
		t.Error("Error parsing AGI 200 response.")
	}
	if len(a.Res) > 1 {
		t.Error("Error parsing AGI 200 response.")
	}
	if a.Res[0] != "1" {
		t.Error("Error parsing AGI 200 response.")
	}
	err = a.parseResponse()
	if len(a.Res) != 2 {
		t.Error("Error parsing AGI complex 200 response.")
	}

	err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing AGI 510 response.")
	}
	err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing AGI 511 response.")
	}
	err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing AGI 520 response.")
	}
	err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing AGI 520 response containing usage details.")
	}
	err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing a partial AGI response.")
	}
	err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing an empty AGI response.")
	}
	err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing an empty AGI response.")
	}
	err = a.parseResponse()
	if err == nil {
		t.Error("No error after parsing an erroneous AGI response.")
	}
}

// Test the generation of AGI commands
func TestCmd(t *testing.T) {
	wc := new(writeConn)
	data := genEnv()
	data = append(data, "200 result=1 endpos=1234\n"...)
	a, err := Init(
		bufio.NewReadWriter(
			bufio.NewReader(bytes.NewReader(data)),
			bufio.NewWriter(io.Writer(wc)),
		),
	)

	if err != nil {
		t.Error("Failed to initialize new AGI session")
	}
	err = a.GetOption("echo", "any")
	if err != nil {
		t.Error("Failed to parse AGI responce")
	}
	if wc.buf == nil {
		t.Error("Failed to send AGI command")
	}
	if string(wc.buf) != "GET OPTION echo \"any\"\n" {
		t.Error("Failed to sent properly formatted AGI command")
	}
	if len(a.Res) < 2 {
		t.Error("Failed to store the full response")
	}
	if a.Res[0] != "1" {
		t.Error("Failed to get the right numeric result")
	}
	if a.Res[1] != "1234" {
		t.Error("Failed to properly parse the rest of the response")
	}
}

// Benchmark AGI session initialisation
func BenchmarkParseEnv(b *testing.B) {
	data := genEnv()
	for i := 0; i < b.N; i++ {
		Init(
			bufio.NewReadWriter(
				bufio.NewReader(bytes.NewReader(data)),
				nil,
			),
		)
	}
}

// Benchmark AGI response parsing
func BenchmarkParseRes(b *testing.B) {
	var a Session
	data := genRes()
	for i := 0; i < b.N; i++ {
		a.Buf = bufio.NewReadWriter(
			bufio.NewReader(bytes.NewReader(data)),
			nil,
		)
		for k := 0; k < 10; k++ {
			a.parseResponse()
		}
	}
}

// Generate AGI environment data
func genEnv() []byte {
	var agiData []byte
	agiData = append(agiData, "agi_network: yes\n"...)
	agiData = append(agiData, "agi_network_script: foo?\n"...)
	agiData = append(agiData, "agi_request: agi://127.0.0.1/foo?\n"...)
	agiData = append(agiData, "agi_channel: SIP/1234-00000000\n"...)
	agiData = append(agiData, "agi_language: en\n"...)
	agiData = append(agiData, "agi_type: SIP\n"...)
	agiData = append(agiData, "agi_uniqueid: 1397044468.0\n"...)
	agiData = append(agiData, "agi_version: 0.1\n"...)
	agiData = append(agiData, "agi_callerid: 1001\n"...)
	agiData = append(agiData, "agi_calleridname: 1001\n"...)
	agiData = append(agiData, "agi_callingpres: 67\n"...)
	agiData = append(agiData, "agi_callingani2: 0\n"...)
	agiData = append(agiData, "agi_callington: 0\n"...)
	agiData = append(agiData, "agi_callingtns: 0\n"...)
	agiData = append(agiData, "agi_dnid: 123456\n"...)
	agiData = append(agiData, "agi_rdnis: unknown\n"...)
	agiData = append(agiData, "agi_context: default\n"...)
	agiData = append(agiData, "agi_extension: 123456\n"...)
	agiData = append(agiData, "agi_priority: 1\n"...)
	agiData = append(agiData, "agi_enhanced: 0.0\n"...)
	agiData = append(agiData, "agi_accountcode: \n"...)
	agiData = append(agiData, "agi_threadid: -1289290944\n"...)
	agiData = append(agiData, "agi_arg_1: foo\n"...)
	agiData = append(agiData, "agi_arg_2: bar\n"...)
	agiData = append(agiData, "agi_arg_3: roo\n\n"...)
	return agiData
}

// Generate AGI Responses
func genRes() []byte {
	var res []byte
	res = append(res, "200 result=1\n"...)
	res = append(res, "200 result=1 (speech) endpos=1234 results=foo bar\n"...)
	res = append(res, "510 Invalid or unknown command\n"...)
	res = append(res, "511 Command Not Permitted on a dead channel\n"...)
	res = append(res, "520 Invalid command syntax.  Proper usage not available.\n"...)
	res = append(res, "520-Invalid command syntax.  Proper usage follows:\nAnswers channel if not already in answer state. Returns -1 on channel failure, or 0 if successful.\n"...)
	res = append(res, "200\n"...)
	res = append(res, "\n\n"...)
	res = append(res, "some random reply that we are not supposed to get\n"...)
	return res
}
