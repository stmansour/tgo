package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
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
	STATETerm
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

// AppsAtOrBeyondState will count the total number of apps in the state requested
// return the count along with the total possible
func AppsAtOrBeyondState(state int, testsonly bool) (count, possible int) {
	count = 1 // tgo goes through all states, but we skip it in the loop below
	possible = 1
	for i := 0; i != envMap.ThisApp && i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
		if testsonly && !envMap.Instances[envMap.ThisInst].Apps[i].IsTest {
			continue
		}
		possible++ // this one contributes to the total possible
		if envMap.Instances[envMap.ThisInst].Apps[i].State >= state {
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

// activateCmd execs the supplied instance (only instance index is provided) with the
// supplied cmd argument. It returns the cmd output as a string.
func activateCmd(i int, cmd string) string {
	a := &envMap.Instances[envMap.ThisInst].Apps[i] // convenient handle for the app we're activating
	filename := fmt.Sprintf("../%s/activate.sh", a.Name)
	ulog("os.Stat(%s)\n", filename)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// TODO: Report error to uhura
		ulog("no activation script: %s\n", filename)
		return "error - no activation script"
	}
	// TODO: cd to cdto
	out, err := exec.Command("activate.sh", cmd).Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

// StateInit puts TGO into the INIT state.
// It will spin through all the apps in this instance and do an 'activate.sh start'
// after starting all of the apps and setting all their states to STATEInitializing
// it writes to the channel telling the StateOrchestrator that it's finished and
// it's time to change states.
func StateUnknown() chan int {
	c := make(chan int)
	go func() {
		ulog("Entering StateUnknown\n")
		var a *appDescr
		me := envMap.ThisApp
		var errResult = regexp.MustCompile(`^error .*`)

		// START UP ALL THE APPS
		ulog("Starting all apps\n")
		for i := 0; i != me && i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
			a = &envMap.Instances[envMap.ThisInst].Apps[i]       // shorthand for accessing the app
			a.State = STATEInitializing                          // we're in the Initializing state now
			filename := fmt.Sprintf("../%s/activate.sh", a.Name) // this is the activation script we'll be hitting
			retval := activateCmd(i, "start")                    // try to start it
			lower := strings.ToLower(retval)                     // see how it went
			switch {
			case lower == "ok": // if it started ok...
				ulog("%s returns ok\n", filename) // update the log...
				a.State = STATEReady              // and move to the READY state

			case errResult.MatchString(lower): // regexp:  begins with error
				ulog("%s returns error: %s\n", filename, retval[6:])
				// TODO: if retryable... keep going, if not, report back BLOCKED

			default:
				ulog("*** ERROR: unexpected reply from %s: %s\n", filename, retval)
			}
		}
		c <- 1                                  // we've started each app. we're done
		ulog("StateUnknown: exiting %d\n", <-c) // no cleanup work to do, just ack and exit
	}()

	return c
}

// StateInit sees TGO through the init state.
// It will spin through all the apps in this instance and see if they are in the INITializing
// or READY state. If not, it will wait 10 seconds and try again. If so, it will signal
// the orchestrator that we're done and it can move to the next state
func StateInit() chan int {
	c := make(chan int)
	go func() {
		var errResult = regexp.MustCompile(`^error .*`)
		// This tgo app (me) can move the READY state. Now just wait on the rest of the apps
		me := envMap.ThisApp
		envMap.Instances[envMap.ThisInst].Apps[me].State = STATEReady

		// Check each app to see if it's ready...
		for {
			for i := 0; i != me && i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
				a := &envMap.Instances[envMap.ThisInst].Apps[i]
				filename := fmt.Sprintf("../%s/activate.sh", a.Name) // this is the activation script we'll be hitting
				retval := activateCmd(i, "ready")
				lower := strings.ToLower(retval)
				switch {
				case lower == "ok":
					ulog("%s returns OK\n", filename)
					a.State = STATEReady

				case errResult.MatchString(lower): // regexp:  begins with error
					ulog("%s returns error: %s\n", filename, retval[6:])
					// TODO: if retryable... keep going, if not, report back BLOCKED

				default:
					ulog("*** ERROR: unexpected reply from %s: %s\n", filename, retval)
				}
			}

			count, possible := AppsAtOrBeyondState(STATEInitializing, false)
			ulog("%d of %d apps are in STATEInitializing\n", count, possible)
			if count == possible {
				c <- 0
				break
			}

			time.Sleep(time.Duration(15 * time.Second))
		}

		//do any cleanup work here, wait for acknowledgement before we exit
		ulog("StateInit: exiting %d\n", <-c)
	}()

	return c
}

// Check on all of our apps and look for all of them to be in the READY
// state. When they are, tell the ordhestrator that we are ready to progress to the next state. If
// not try doing an activate stop followed by an activate init.
func StateReady() chan int {
	c := make(chan int)
	go func() {
		ulog("Entering StateReady\n")
		var errResult = regexp.MustCompile(`^error .*`)
		var a *appDescr
		me := envMap.ThisApp

		for {
			for i := 0; i != me && i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
				filename := fmt.Sprintf("../%s/activate.sh", a.Name) // this is the activation script we'll be hitting
				retval := activateCmd(i, "ready")
				lower := strings.ToLower(retval)
				switch {
				case lower == "ok":
					ulog("%s returns OK\n", filename)
					a.State = STATEReady

				case errResult.MatchString(lower): // regexp:  begins with error
					ulog("%s returns error: %s\n", filename, retval[6:])
					// TODO: if retryable... keep going, if not, report back BLOCKED

				default:
					ulog("*** ERROR: unexpected reply from %s: %s\n", filename, retval)
				}
			}
			count, possible := AppsAtOrBeyondState(STATEReady, false)
			ulog("%d of %d apps are in STATETesting\n", count, possible)
			if count == possible {
				c <- 0
				break
			}
			time.Sleep(time.Duration(15 * time.Second))
		}

		//wait for acknowledgement before we exit
		ulog("StateReady: exiting %d\n", <-c)
	}()

	return c
}

// StateTest puts TGO into the TEST state.
func StateTest() chan int {
	c := make(chan int)
	go func() {
		ulog("Entering StateTest\n")
		var errResult = regexp.MustCompile(`^error .*`)
		var a *appDescr
		me := envMap.ThisApp

		// Start all tests...
		for i := 0; i != me && i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
			a = &envMap.Instances[envMap.ThisInst].Apps[i]
			if a.IsTest {
				filename := fmt.Sprintf("../%s/activate.sh", a.Name) // this is the activation script we'll be hitting
				retval := activateCmd(i, "test")
				lower := strings.ToLower(retval)
				a.State = STATETesting
				switch {
				case lower == "ok":
					ulog("%s returns OK\n", filename)

				case errResult.MatchString(lower): // regexp:  begins with error
					ulog("%s returns error: %s\n", filename, retval[6:])
					// TODO: if retryable... keep going, if not, report back BLOCKED

				default:
					ulog("*** ERROR: unexpected reply from %s: %s\n", filename, retval)
				}
			} else {
				a.State = STATETesting
			}
		}

		// This tgo app (me) can move the DONE state. Now just wait on the tests to finish
		envMap.Instances[envMap.ThisInst].Apps[me].State = STATEDone
		for {
			for i := 0; i != me && i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
				a = &envMap.Instances[envMap.ThisInst].Apps[i]
				if a.IsTest {
					filename := fmt.Sprintf("../%s/activate.sh", a.Name) // this is the activation script we'll be hitting
					retval := activateCmd(i, "teststatus")
					lower := strings.ToLower(retval)
					switch {
					case lower == "done":
						ulog("%s returns DONE\n", filename)
						a.State = STATEDone // replace this statement with the real code

					case lower == "testing":
						// nothing to do, let it keep running

					case errResult.MatchString(lower): // regexp:  begins with error
						ulog("%s returns error: %s\n", filename, retval[6:])
						// TODO: if retryable... keep going, if not, report back BLOCKED

					default:
						ulog("*** ERROR: unexpected reply from %s: %s\n", filename, retval)
					}
				}
			}

			count, possible := AppsAtOrBeyondState(STATEDone, true)
			ulog("%d of %d apps are in STATEDone\n", count, possible)
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
func StateOrchestrator(alldone chan int) {
	var r StatusReply
	ulog("Orchestrator: StateUnknown started\n")
	//#################################################################################
	//   UNKNOWN
	//#################################################################################
	c := StateUnknown()
	select {
	case i := <-c:
		ulog("Orchestrator: StateUnknown completed:  %d\n", i)
		c <- 0 // tell the StateInit handler it's ok to exit
	case <-time.After(30 * time.Minute):
		ulog("Orchestrator: StateUnknown has not responded in 30 minutes. Giving up!\n")
		// TODO:  tell uhura that startup has timed out
		os.Exit(1)
	}

	ulog("Orchestrator: StateInit started\n")
	//#################################################################################
	//   INIT
	//#################################################################################
	c = StateInit()
	select {
	case i := <-c:
		ulog("Orchestrator: StateInit completed:  %d\n", i)
		c <- 0 // tell the StateInit handler it's ok to exit
	case <-time.After(30 * time.Minute):
		ulog("Orchestrator: StateInit has not responded in 30 minutes. Giving up!\n")
		// TODO:  tell uhura that startup has timed out
		os.Exit(1)
	}

	//#################################################################################
	//   READY
	// When we enter READY state, there's really
	// nothing to do except wait for UHURA to send
	// us a TESTNOW command. Then we start up the
	// tests.
	//#################################################################################
	ulog("Orchestrator: Entering StateReady\n")
	c = StateReady()

	PostStatusAndGetReply("READY", &r) // tell Uhura we're ready
	ulog("Orchestrator: Posted READY status to uhura. ReplyCode: %d\n", r.ReplyCode)
	ulog("Orchestrator: Calling StateReady\n")
	ulog("Orchestrator: waiting for StateReady to reply\n")
	select {
	case i := <-c:
		ulog("Orchestrator: StateReady completed:  %d\n", i)
		c <- 0 // tell the StateInit handler it's ok to exit
	case <-time.After(15 * time.Minute):
		ulog("Orchestrator: StateReady has not responded in 15 minutes. Giving up!\n")
		// TODO:  tell uhura that startup has timed out
		os.Exit(1)
	}

	//#################################################################################
	//   TEST
	// Before we can begin the test mode, we need to hear back from uhura
	// that we can begin testing.
	//#################################################################################
	ulog("Orchestrator: READY TO TRANSITION TO TEST, read channel Tgo.UhuraComm\n")
	ulog("waiting for Uhura to contact tgo\n")
	select {
	case i := <-Tgo.UhuraComm:
		ulog("Orchestrator: Comms reports uhura has sent command:  %d\n", i)
		if i == cmdTESTNOW {
			ulog("Proceding to state TEST\n")
		} else {
			ulog("Unexpected response: %d.  Not sure what to do, so proceeding...\n", i)
		}
		ulog("Orchestrator: TRANSITION TO TEST, writing to channel Tgo.UhuraComm\n")
		Tgo.UhuraComm <- 0 // tell the HTTP handler it's ok to exit
	case <-time.After(30 * time.Minute):
		ulog("Orchestrator: We have not heard from Uhura in 30 minutes. Giving up!\n")
		// TODO:  tell uhura that startup has timed out
		os.Exit(1)
	}

	PostStatusAndGetReply("TEST", &r) // Tel UHURA we're moving to the TEST state
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

	//#################################################################################
	//   DONE
	//#################################################################################
	PostStatusAndGetReply("DONE", &r) // starting our state machine in the INIT state
	ulog("Posted DONE status to uhura. ReplyCode: %d\n", r.ReplyCode)

	//#################################################################################
	//   TERM
	//#################################################################################
	// TODO:  call all apps and tell them to exit
	//        this is optional because uhura will terminate all instances when it has
	//        completed or timed out.

	ulog("StateOrchestrator exiting\n")

	alldone <- 1 // we're all done

}

// InitiateStateMachine essentially pulls together the mission for this TGO instance
// and sets it into motion.
func InitiateStateMachine(alldone chan int) {
	whoAmI()
	ulog("I am instance %d, my name is %s, I am app index %d\n",
		envMap.ThisInst, envMap.Instances[envMap.ThisInst].InstName, envMap.ThisApp)
	ulog("I will listen for commands on port %d\n",
		envMap.Instances[envMap.ThisInst].Apps[envMap.ThisApp].UPort)
	envMap.Instances[envMap.ThisInst].Apps[envMap.ThisApp].State = STATEInitializing
	var r StatusReply
	go UhuraComms()                   // handle anything that comes from uhura
	PostStatusAndGetReply("INIT", &r) // starting our state machine in the INIT state
	go StateOrchestrator(alldone)     // let the orchestrator handle it from here
}
