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

type AppDescr struct {
	UID    string
	Name   string
	Repo   string
	UPort  int
	IsTest bool
	State  int
	RunCmd string
}

type InstDescr struct {
	InstName string
	OS       string
	HostName string
	Apps     []AppDescr
}

type EnvDescr struct {
	EnvName   string
	UhuraURL  string
	UhuraPort int
	ThisInst  int
	ThisApp   int // not in uhura's def. This is tgo's index within the Apps array
	State     int
	Instances []InstDescr
}

var EnvMap EnvDescr

func dPrintStatusReply(r *StatusReply) {
	if Tgo.Debug {
		ulog("Status Reply = %+v\n", *r)
	}
	if Tgo.DebugToScreen {
		fmt.Printf("Status Reply = %+v\n", *r)
	}
}

func PostStatusAndGetReply(state string, r *StatusReply) {
	s := StatusMsg{state,
		EnvMap.Instances[EnvMap.ThisInst].InstName,
		EnvMap.Instances[EnvMap.ThisInst].Apps[EnvMap.ThisApp].UID,
		time.Now().Format(time.RFC822)}

	rc, e := PostStatus(&s, r)
	if nil != e {
		ulog("PostStatus returned error:  %v\n", e)
		os.Exit(5)
	}

	if rc != 200 {
		ulog("Bad HTTP response code: %d\n", rc)
		os.Exit(3)
	}

	if r.ReplyCode != RespOK {
		ulog("Uhura is not happy:  response to status: %d\n", r.ReplyCode)
		dPrintStatusReply(r)
		os.Exit(4)
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
	WhoAmI()
	fmt.Printf("I am instance %d, my name is %s, I am app index %d\n",
		EnvMap.ThisInst, EnvMap.Instances[EnvMap.ThisInst].InstName, EnvMap.ThisApp)
	fmt.Printf("I will listen for commands on port %d\n",
		EnvMap.Instances[EnvMap.ThisInst].Apps[EnvMap.ThisApp].UPort)

	var r StatusReply
	PostStatusAndGetReply("INIT", &r)
	fmt.Printf("I successfully contacted uhura at %s and got a reply\n", EnvMap.UhuraURL)

}
