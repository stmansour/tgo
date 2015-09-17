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

// Count the total number of apps in the state requested
// return the count
func AppsInState(state int) int {
	count := 0
	for i := 0; i < len(EnvMap.Instances[EnvMap.ThisInst].Apps); i++ {
		if EnvMap.Instances[EnvMap.ThisInst].Apps[i].State == state {
			count++
		}
	}
	return count
}

// Spin through all the apps in this instance
// Start them up. When they're all in the READY state, signal
// the orchestrator that we're done and it can move to the READY state
func StateInit() chan int {
	c := make(chan int)
	go func() {
		ulog("Entering StateInit\n")
		var a *AppDescr
		me := EnvMap.ThisApp
		for i := 0; i != me && i < len(EnvMap.Instances[EnvMap.ThisInst].Apps); i++ {
			a = &EnvMap.Instances[EnvMap.ThisInst].Apps[i]
			a.State = STATE_Initializing
			// TODO:  call their activate script
		}

		// we're ready and now just waiting on the rest of the apps
		EnvMap.Instances[EnvMap.ThisInst].Apps[me].State = STATE_Ready

		// TODO: wait for everybody to initialize
		if len(EnvMap.Instances[EnvMap.ThisInst].Apps) == AppsInState(STATE_Ready) {
			c <- 0
		} else {
			c <- 1
		}

		//do any cleanup work here, wait for acknowledgement before we exit
		ulog("Exiting StateInit: %d\n", <-c)
	}()
	return c
}

func StateReady() {

}

func StateTest() {

}

func StateDone() {

}

func StateOrchestrator() {
	ulog("Orchestrator: StateInit started\n")
	c := StateInit()
	select {
	case i := <-c:
		ulog("Orchestrator: StateInit completed:  %d\n", i)
		c <- 0 // tell the StateInit handler it's ok to exit
	case <-time.After(5 * time.Minute):
		ulog("Orchestrator: StateInit has not responded in 5 minutes. Giving up!\n")
		// TODO:  tell uhura that startup has timed out
	}

	// TODO: the rest of the states...

}

func InitiateStateMachine() {
	WhoAmI()

	fmt.Printf("I am instance %d, my name is %s, I am app index %d\n",
		EnvMap.ThisInst, EnvMap.Instances[EnvMap.ThisInst].InstName, EnvMap.ThisApp)
	fmt.Printf("I will listen for commands on port %d\n",
		EnvMap.Instances[EnvMap.ThisInst].Apps[EnvMap.ThisApp].UPort)
	EnvMap.Instances[EnvMap.ThisInst].Apps[EnvMap.ThisApp].State = STATE_Initializing
	var r StatusReply
	PostStatusAndGetReply("INIT", &r) // starting our state machine in the INIT state
	fmt.Printf("I successfully contacted uhura at %s and got a reply\n", EnvMap.UhuraURL)
	StateOrchestrator() // let the orchestrator handle it from here
}
