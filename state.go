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

func dPrintStatusReply(r *StatusReply) {
	if Tgo.Debug {
		ulog("Status Reply = %+v\n", *r)
	}
	if Tgo.DebugToScreen {
		fmt.Printf("Status Reply = %+v\n", *r)
	}
}

func PostStatusAndGetReply(state string, r *StatusReply) {
	s := StatusMsg{state, "TGOtest", "tgo0", time.Now().Format(time.RFC822)}
	rc, e := PostStatus(&s, r)
	if nil != e {
		ulog("PostStatus returned error:  %v\n", e)
		os.Exit(5)
	}

	if rc != 200 {
		ulog("Bad HTTP response code: %d\n", rc)
	}

	if r.ReplyCode != RespOK {
		ulog("Uhura is not happy:  response to status: %d\n", r.ReplyCode)
		dPrintStatusReply(r)
		os.Exit(1)
	}
}

func StateInit() {

}

func StateReady() {

}

func StateTest() {

}

func StateDone() {

}

func InitiateStateMachine() {
	var r StatusReply
	PostStatusAndGetReply("INIT", &r)
	fmt.Printf("WE SUCCESSFULLY CONTACTED UHURA AND GOT A REPLY\n")
}
