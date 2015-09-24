package main

import (
	"fmt"
	"os"
	"time"
)

// STATEUninitialized - STATEDone represent the states through which
// moves to accomplish its mission -- starting one or more apps and
// managing them through startup, testing, and shutdown.
const (
	STATEUninitialized = iota
	STATEInitializing
	STATEReady
	STATETesting
	STATEDone
)

const (
	cmdTESTNOW = iota // tells Tgo to initiate testing
	cmdSTOP
)

type appDescr struct {
	UID    string
	Name   string
	Repo   string
	UPort  int
	IsTest bool
	State  int
	RunCmd string
}

type instDescr struct {
	InstName string
	OS       string
	HostName string
	Apps     []appDescr
}

type envDescr struct {
	EnvName   string
	UhuraURL  string
	UhuraPort int
	ThisInst  int
	ThisApp   int // not in uhura's def. This is tgo's index within the Apps array
	State     int
	Instances []instDescr
}

var envMap envDescr

func dPrintStatusReply(r *StatusReply) {
	if Tgo.Debug {
		ulog("Status Reply = %+v\n", *r)
	}
	if Tgo.DebugToScreen {
		fmt.Printf("Status Reply = %+v\n", *r)
	}
}

// AppsInState will count the total number of apps in the state requested
// return the count along with the total possible
func AppsInState(state int, testsonly bool) (count, possible int) {
	count = 0
	possible = 0
	for i := 0; i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
		if testsonly && !envMap.Instances[envMap.ThisInst].Apps[i].IsTest {
			continue
		}
		possible++ // this one contributes to the total possible
		if envMap.Instances[envMap.ThisInst].Apps[i].State == state {
			count++
		}
	}
	return count, possible
}

// PostStatusAndGetReply does exactly as the title suggests.
// TODO: probably need to add some error handling for common
// http error types where we can retry.
func PostStatusAndGetReply(state string, r *StatusReply) {
	s := StatusMsg{state,
		envMap.Instances[envMap.ThisInst].InstName,
		envMap.Instances[envMap.ThisInst].Apps[envMap.ThisApp].UID,
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

// StateInit puts TGO into the INIT state.
// It will spin through all the apps in this instance
// Start them up. When they're all in the READY state, signal
// the orchestrator that we're done and it can move to the next state
func StateInit() chan int {
	c := make(chan int)
	go func() {
		ulog("Entering StateInit\n")
		var a *appDescr
		me := envMap.ThisApp
		for i := 0; i != me && i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
			a = &envMap.Instances[envMap.ThisInst].Apps[i]
			// TODO:  call their activate script to START
			a.State = STATEInitializing
		}

		// This tgo app (me) can move the READY state. Now just wait on the rest of the apps
		envMap.Instances[envMap.ThisInst].Apps[me].State = STATEReady
		for {
			for i := 0; i != me && i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
				a = &envMap.Instances[envMap.ThisInst].Apps[i]
				// TODO:  call their activate script
				a.State = STATEReady // this is a fake statement, just to get the code going
			}

			count, possible := AppsInState(STATEReady, false)
			ulog("%d of %d apps are in STATEReady\n", count, possible)
			if count == possible {
				c <- 0
				break
			}

			time.Sleep(time.Duration(10 * time.Second))
		}

		//do any cleanup work here, wait for acknowledgement before we exit
		ulog("StateInit: exiting %d\n", <-c)
	}()

	return c
}

// // Check on all of our apps.  When all the apps are in the READY
// // state we tell the ordhestrator that we are ready to progress to the next state
// func StateReady() chan int {
// 	c := make(chan int)
// 	go func() {
// 		ulog("Entering StateReady\n")
// 		var a *appDescr
// 		me := envMap.ThisApp
// 		for i := 0; i != me && i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
// 			a = &envMap.Instances[envMap.ThisInst].Apps[i]
// 			// TODO:  call their activate script
// 			a.State = STATETesting // replace this statement with the real code
// 		}

// 		// We'll put tgo into testing mode (since it is waiting on test apps )
// 		envMap.Instances[envMap.ThisInst].Apps[me].State = STATETesting
// 		for {
// 			for i := 0; i != me && i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
// 				a = &envMap.Instances[envMap.ThisInst].Apps[i]
// 				// TODO:  call their activate script
// 				a.State = STATEReady // this is a fake statement, just to get the code going
// 			}

// 			count, possible := AppsInState(STATETesting, true)
// 			ulog("%d of %d apps are in STATETesting\n", count, possible)
// 			if count == possible {
// 				c <- 0
// 				break
// 			}

// 			time.Sleep(time.Duration(10 * time.Second))
// 		}

// 		//do any cleanup work here, wait for acknowledgement before we exit
// 		ulog("StateReady: exiting %d\n", <-c)
// 	}()

// 	return c
// }

// StateTest puts TGO into the TEST state.
func StateTest() chan int {
	c := make(chan int)
	go func() {
		ulog("Entering StateTest\n")
		var a *appDescr
		me := envMap.ThisApp
		for i := 0; i != me && i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
			a = &envMap.Instances[envMap.ThisInst].Apps[i]
			// TODO:  call their activate script
			a.State = STATEDone // replace this statement with the real code
		}

		// This tgo app (me) can move the DONE state. Now just wait on the tests to finish
		envMap.Instances[envMap.ThisInst].Apps[me].State = STATEDone
		for {
			for i := 0; i != me && i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
				a = &envMap.Instances[envMap.ThisInst].Apps[i]
				// TODO:  call their activate script
				a.State = STATEDone // this is a fake statement, just to get the code going
			}

			count, possible := AppsInState(STATEDone, true)
			ulog("%d of %d apps are in STATETesting\n", count, possible)
			if count == possible {
				c <- 0
				break
			}

			time.Sleep(time.Duration(10 * time.Second))
		}

		//do any cleanup work here, wait for acknowledgement before we exit
		ulog("StateTest: exiting %d\n", <-c)
	}()

	return c
}

// StateDone puts TGO into the DONE state. This may not be necessary
func StateDone() {
	// nothing to do at the moment
}

// StateOrchestrator manages the states through which TGO
// progresses. It decides when we need to switch states and makes
// the change.
func StateOrchestrator() {
	var r StatusReply
	ulog("Orchestrator: StateInit started\n")
	//############################################
	//   INIT
	//############################################
	c := StateInit()
	select {
	case i := <-c:
		ulog("Orchestrator: StateInit completed:  %d\n", i)
		c <- 0 // tell the StateInit handler it's ok to exit
	case <-time.After(30 * time.Minute):
		ulog("Orchestrator: StateInit has not responded in 30 minutes. Giving up!\n")
		// TODO:  tell uhura that startup has timed out
		os.Exit(1)
	}

	//############################################
	//   READY
	// When we enter READY state, there's really
	// nothing to do except wait for UHURA to send
	// us a TESTNOW command. Then we start up the
	// tests.
	//############################################
	// ulog("Orchestrator: Entering READY state\n")
	// PostStatusAndGetReply("READY", &r) // starting our state machine in the INIT state
	// ulog("Orchestrator: Posted READY status to uhura. ReplyCode: %d\n", r.ReplyCode)
	// ulog("Orchestrator: Calling StateReady\n")
	// c = StateReady()
	// ulog("Orchestrator: waiting for StateReady to reply\n")
	// select {
	// case i := <-c:
	// 	ulog("Orchestrator: StateReady completed:  %d\n", i)
	// 	c <- 0 // tell the StateInit handler it's ok to exit
	// case <-time.After(5 * time.Minute):
	// 	ulog("Orchestrator: StateReady has not responded in 5 minutes. Giving up!\n")
	// 	// TODO:  tell uhura that startup has timed out
	// 	os.Exit(1)
	// }

	//############################################
	//   TEST
	//############################################
	// ulog("Orchestrator: READY TO TRANSITION TO TEST, read channel Tgo.UhuraComm\n")
	// // Before we can begin the test mode, we need to hear back from uhura
	// // that we can begin testing.
	// select {
	// case i := <-Tgo.UhuraComm:
	// 	ulog("Orchestrator: Comms reports uhura has sent command:  %d\n", i)
	// 	if i == cmdTESTNOW {
	// 		ulog("Proceding to state TEST\n")
	// 	} else {
	// 		ulog("Unexpected response: %d.  Not sure what to do, so proceeding...\n", i)
	// 	}
	// 	ulog("Orchestrator: TRANSITION TO TEST, writing to channel Tgo.UhuraComm\n")
	// 	Tgo.UhuraComm <- 0 // tell the HTTP handler it's ok to exit
	// case <-time.After(30 * time.Minute):
	// 	ulog("Orchestrator: We have not heard from Uhura in 30 minutes. Giving up!\n")
	// 	// TODO:  tell uhura that startup has timed out
	// 	os.Exit(1)
	// }

	PostStatusAndGetReply("TEST", &r) // starting our state machine in the INIT state
	ulog("Posted TEST status to uhura. ReplyCode: %d\n", r.ReplyCode)
	c = StateTest()
	select {
	case i := <-c:
		ulog("Orchestrator: StateTest completed:  %d\n", i)
		c <- 0 // tell the StateInit handler it's ok to exit
	case <-time.After(30 * time.Minute):
		ulog("Orchestrator: StateTest has not responded in 30 minutes. Giving up!\n")
		// TODO:  tell uhura that startup has timed out
		os.Exit(1)
	}

	//############################################
	//   DONE
	//############################################
	PostStatusAndGetReply("DONE", &r) // starting our state machine in the INIT state
	ulog("Posted DONE status to uhura. ReplyCode: %d\n", r.ReplyCode)

}

// InitiateStateMachine essentially pulls together the mission for this TGO instance
// and sets it into motion.
func InitiateStateMachine() {
	whoAmI()
	ulog("I am instance %d, my name is %s, I am app index %d\n",
		envMap.ThisInst, envMap.Instances[envMap.ThisInst].InstName, envMap.ThisApp)
	ulog("I will listen for commands on port %d\n",
		envMap.Instances[envMap.ThisInst].Apps[envMap.ThisApp].UPort)
	envMap.Instances[envMap.ThisInst].Apps[envMap.ThisApp].State = STATEInitializing
	var r StatusReply
	go UhuraComms()                   // handle anything that comes from uhura
	PostStatusAndGetReply("INIT", &r) // starting our state machine in the INIT state
	StateOrchestrator()               // let the orchestrator handle it from here
}
