package main

import (
	"fmt"
	"os"
	"time"
)

const (
	STATE_Uninitialized = iota
	STATE_Initializing
	STATE_Ready
	STATE_Testing
	STATE_Done
)

func StateInit() {

}

func StateReady() {

}

func StateTest() {

}

func StateDone() {

}

func InitiateStateMachine() {
	var s = StatusMsg{"INIT", "TGOtest", "gto0", ""}
	s.Tstamp = time.Now().Format(time.RFC822)

	var r StatusReply
	rc, e := PostStatus(&s, &r)
	if nil != e {
		ulog("PostStatus returned error:  %v\n", e)
		os.Exit(5)
	}

	if rc != 200 {
		ulog("Bad HTTP response code: %d\n", rc)
	}

	if r.ReplyCode != RespOK {
		ulog("Uhura is not happy:  response to status: %d\n", r.ReplyCode)
		os.Exit(1)
	}
	fmt.Printf("WE SUCCESSFULLY CONTACTED UHURA AND GOT A REPLY\n")
}
