package main

import (
	"strings"
	"testing"
)

func TestParseLine_welcome(t *testing.T) {

	line1 := ":barjavel.freenode.net 001 graham_king :Welcome to the freenode Internet Relay Chat Network graham_king"
	line, err := ParseLine(line1)

	if err != nil {
		t.Error("ParseLine error: ", err)
	}

	if line.Command != "001" {
		t.Error("Command incorrect")
	}
	if line.Host != "barjavel.freenode.net" {
		t.Error("Host incorrect")
	}
}

func TestParseLine_privmsg(t *testing.T) {
	line1 := ":rnowak!~rnowak@q.ovron.com PRIVMSG #linode :totally"
	line, err := ParseLine(line1)

	if err != nil {
		t.Error("ParseLine error: ", err)
	}

	if line.Command != "PRIVMSG" {
		t.Error("Command incorrect. Got", line.Command)
	}
	if line.Host != "~rnowak@q.ovron.com" {
		t.Error("Host incorrect. Got", line.Host)
	}
	if line.User != "rnowak" {
		t.Error("User incorrect. Got", line.User)
	}
	if line.Channel != "#linode" {
		t.Error("Channel incorrect. Got", line.Channel)
	}
	if line.Content != "totally" {
		t.Error("Content incorrect. Got", line.Content)
	}
}

func TestParseLine_list(t *testing.T) {
	line1 := ":oxygen.oftc.net 322 graham_king #linode 412 :Linode Community Support | http://www.linode.com/ | Linodes in Asia-Pacific! - http://bit.ly/ooBzhV"
	line, err := ParseLine(line1)

	if err != nil {
		t.Error("ParseLine error: ", err)
	}

	if line.Command != "322" {
		t.Error("Command incorrect. Got", line.Command)
	}
	if line.Host != "oxygen.oftc.net" {
		t.Error("Host incorrect. Got", line.Host)
	}
	if line.Channel != "#linode" {
		t.Error("Channel incorrect. Got", line.Channel)
	}
	if !strings.Contains(line.Content, "Community Support") {
		t.Error("Content incorrect. Got", line.Content)
	}
	if line.Args[2] != "412" {
		t.Error("Args incorrect. Got", line.Args)
	}
}

func TestParseLine_away(t *testing.T) {
	line1 := ":hybrid7.debian.local 301 graham_king graham :Not here"
	line, err := ParseLine(line1)

	if err != nil {
		t.Error("ParseLine error: ", err)
	}

	if line.Command != "301" {
		t.Error("Command incorrect. Got", line.Command)
	}
	if line.Channel != "graham" {
		t.Error("Channel incorrect. Got", line.Channel)
	}
	if line.User != "graham" {
		t.Error("User incorrect. Got", line.User)
	}
}
