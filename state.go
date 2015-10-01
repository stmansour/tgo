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
	ThisApp   int // not in uhura's def. This is tgo's index within the Apps array. Tgo looks it up
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
	for i := 0; i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
		if i == envMap.ThisApp || (testsonly && !envMap.Instances[envMap.ThisInst].Apps[i].IsTest) {
			continue
		}
		possible++ // this one contributes to the total possible
		if envMap.Instances[envMap.ThisInst].Apps[i].State >= state {
			count++
		}
	}
	return count, possible
}

// // PostExtendedStatusAndGetReply does exactly as the title suggests.
// // TODO: probably need to add some error handling for common
// // http error types where we can retry.
// func qqPostExtendedStatusAndGetReply(iapp int, state string, r *StatusReply) {
// 	s := StatusMsg{state,
// 		envMap.Instances[envMap.ThisInst].InstName,
// 		envMap.Instances[envMap.ThisInst].Apps[iapp].UID,
// 		time.Now().Format(time.RFC822)}

// 	rc, e := PostStatus(&s, r)
// 	if nil != e {
// 		ulog("PostStatus returned error:  %v\n", e)
// 		os.Exit(5)
// 	}

// 	if rc != 200 {
// 		ulog("Bad HTTP response code: %d\n", rc)
// 		os.Exit(3)
// 	}

// 	if r.ReplyCode != RespOK {
// 		ulog("Uhura is not happy:  response to status: %d\n", r.ReplyCode)
// 		dPrintStatusReply(r)
// 		os.Exit(4)
// 	}
// }

// PostExtendedStatusAndGetReply does exactly as the title suggests.
// TODO: probably need to add some error handling for common
// http error types where we can retry.
func PostExtendedStatusAndGetReply(iapp int, state string, r *StatusReply, mapname *string, m *map[string]string) {
	var s StatusMsgExt
	s.State = state
	s.InstName = envMap.Instances[envMap.ThisInst].InstName
	s.UID = envMap.Instances[envMap.ThisInst].Apps[iapp].UID
	s.Tstamp = time.Now().Format(time.RFC822)
	if m != nil && mapname != nil {
		s.KV.Name = *mapname
		for k, v := range *m {
			s.KV.KVs = append(s.KV.KVs, KeyVal{k, v})
			// ulog("EXTENDED STATUS DATA:  %s: %s\n", k, v)
		}
	}

	rc, e := PostStatusExt(&s, r)
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
/*

prog -h {{InstName.prop}} -p {{InstName.AppName.prop}}

*/
func activateCmd(i int, cmd string) string {
	var err error
	var out []byte
	var rsp string

	a := &envMap.Instances[envMap.ThisInst].Apps[i] // convenient handle for the app we're activating
	dirname := fmt.Sprintf("../%s", a.Name)
	if err := os.Chdir(dirname); err != nil {
		ulog("could not cd to %s:  %v\n", dirname, err)
	}

	// To launch apps, we will handle the special case where a run command is supplied.
	// This essentially replaces 'activate.sh start'.  If no command is supplied, then
	// just use 'activate.sh start'.  Also updated for the test command. So this can be
	// used to replace 'activate.sh test'
	if a.RunCmd != "" && // special handling if there's a RunCmd and...
		((cmd == "start" && !a.IsTest) || // either we're starting an app
			(cmd == "test" && a.IsTest)) { // or we're starting a test
		cmd := envDescrSub(a.RunCmd)
		ca := strings.Split(cmd, " ")
		ulog("os.Stat(%s/%s)\n", dirname, ca[0])
		if _, err = os.Stat(ca[0]); os.IsNotExist(err) {
			// TODO: Report error to uhura
			ulog("no such application %s in %s\n", ca[0], dirname)
			return "error - run command points to non-existent app"
		}
		// Not sure what to do with the running command. We'll start it in a go function
		mycmd := fmt.Sprintf("./%s", ca[0])
		c := exec.Command(mycmd, ca[1:]...)
		err := c.Start()
		if err != nil {
			rsp = fmt.Sprintf("*** Error running %s %v -- error = %v\n", mycmd, ca[1:], err)
			ulog(rsp)
		} else {
			rsp = "OK"
			ulog("Successfully started: %s\n", cmd)
		}
		go func() {
			err = c.Wait()
			ulog("Finished: command %s\n         err returned = %v\n", cmd, err)
		}()
	} else {
		ulog("os.Stat(%s/activate.sh)\n", dirname)
		if _, err = os.Stat("activate.sh"); os.IsNotExist(err) {
			// TODO: Report error to uhura
			ulog("no activation script in: %s\n", dirname)
			return "error - no activation script"
		}
		out, err = exec.Command("./activate.sh", cmd).Output()
		if err != nil {
			log.Fatal(err)
		}
		rsp = string(out)
	}
	os.Chdir("../tgo")
	return rsp
}

// actionAllApps calls the activate.sh script for all Apps (excluding tgo itself)
// If the result is "OK" then it automatically sends uhura the status for each app.
func actionAllApps(actCmd string, expect string, stateval int, status string) {
	me := envMap.ThisApp
	var errResult = regexp.MustCompile(`^error .*`)
	for i := 0; i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
		a := &envMap.Instances[envMap.ThisInst].Apps[i] // shorter notation
		if i == me || a.State >= stateval {             // skip tgo, and any app already at or beyond reqested state
			continue
		}
		filename := fmt.Sprintf("../%s/activate.sh", a.Name) // this is the activation script we'll be hitting
		retval := activateCmd(i, actCmd)                     // run the command
		lower := strings.ToLower(retval)                     // see how it went
		lower = strings.TrimRight(lower, "\n\r")             // remove CR, LF
		switch {
		case lower == expect: // if it started ok...
			ulog("%s %s returns %s\n", filename, actCmd, expect) // update the log...
			a.State = stateval                                   // and move to the Init state
			var r StatusReply
			PostExtendedStatusAndGetReply(i, status, &r, nil, nil)
			// TODO: look at this reply and act on it if necessary
		case errResult.MatchString(lower): // regexp:  begins with error
			ulog("%s returns error: %s\n", filename, retval[6:])
			// TODO: if retryable... keep going, if not, report back BLOCKED
		default:
			ulog("*** ERROR: unexpected reply from %s: %s\n", filename, retval)
		}
	}
}

// StateInit puts TGO into the INIT state.
// 'activate.sh start' all apps
// set all their states to STATEInitializing
func StateUnknown() chan int {
	c := make(chan int)
	go func() {
		ulog("Entering StateUnknown\n")
		ulog("Starting all apps\n")
		actionAllApps("start", "ok", STATEInitializing, "INIT")
		c <- 1                                  // we've started each app. we're done
		ulog("StateUnknown: exiting %d\n", <-c) // no cleanup work to do, just ack and exit
	}()

	return c
}

// StateInit sees TGO through the init state.
// 'activate.sh ready' all apps.
// For each app that's ready, move it to the INIT state.
// If all apps are not in the ready state, it will wait 15 seconds and try again.
// It will stay in this mode forever until all the apps are in the init state or beyond
func StateInit() chan int {
	c := make(chan int)
	go func() {
		envMap.Instances[envMap.ThisInst].Apps[envMap.ThisApp].State = STATEReady // tgo is READY, just waiting on apps now
		for {
			actionAllApps("ready", "ok", STATEInitializing, "INIT")           // activate.sh ready
			count, possible := AppsAtOrBeyondState(STATEInitializing, false)  // how many are ready or init
			ulog("%d of %d apps are in STATEInitializing\n", count, possible) // log results
			if count == possible {                                            // if all are at least in the init state move on
				c <- 0 // tell StateOrchestrator we're done
				break  // bust out of the loop
			}
			time.Sleep(time.Duration(15 * time.Second)) // if any of the apps are still UNKNOWN wait and try again
		}
		ulog("StateInit: exiting %d\n", <-c) //do any cleanup work before this point
	}()
	return c
}

// 'activate.sh ready' all apps.  They will probably already be in the READY state,
// but this is the final check. If there were slow starters during the INIT phase
// they may need the time.
func StateReady() chan int {
	c := make(chan int)
	go func() {
		ulog("Entering StateReady\n")
		for {
			actionAllApps("ready", "ok", STATEReady, "READY") // activate.sh ready
			count, possible := AppsAtOrBeyondState(STATEReady, false)
			ulog("%d of %d apps are in STATETesting\n", count, possible)
			if count == possible {
				c <- 0
				break
			}
			time.Sleep(time.Duration(15 * time.Second))
		}
		ulog("StateReady: exiting %d\n", <-c) //do any cleanup work before this point
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
		for i := 0; i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
			if i == me {
				continue
			}
			a = &envMap.Instances[envMap.ThisInst].Apps[i]
			if a.IsTest {
				filename := fmt.Sprintf("../%s/activate.sh", a.Name) // this is the activation script we'll be hitting
				retval := activateCmd(i, "test")
				lower := strings.ToLower(retval)
				lower = strings.TrimRight(lower, "\n\r") // remove CR, LF
				a.State = STATETesting
				switch {
				case lower == "ok":
					ulog("%s returns OK\n", filename)
					a.State = STATETesting
					var r StatusReply
					PostExtendedStatusAndGetReply(i, "TEST", &r, nil, nil)

				case errResult.MatchString(lower): // regexp:  begins with error
					ulog("%s returns error: %s\n", filename, retval[6:])
					// TODO: if retryable... keep going, if not, report back BLOCKED

				default:
					ulog("*** ERROR: unexpected reply to 'test' command from %s: %s\n", filename, retval)
				}
			} else {
				a.State = STATETesting
				var r StatusReply
				PostExtendedStatusAndGetReply(i, "TEST", &r, nil, nil)
			}
		}

		// This tgo app (me) can move the DONE state. Now just wait on the tests to finish
		envMap.Instances[envMap.ThisInst].Apps[me].State = STATEDone
		for {
			for i := 0; i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
				if i == me {
					continue
				}
				a = &envMap.Instances[envMap.ThisInst].Apps[i]
				if a.IsTest {
					filename := fmt.Sprintf("../%s/activate.sh", a.Name) // this is the activation script we'll be hitting
					retval := activateCmd(i, "teststatus")
					lower := strings.ToLower(retval)
					lower = strings.TrimRight(lower, "\n\r") // remove CR, LF
					switch {
					case lower == "done":
						ulog("%s returns DONE\n", filename)
						// get the test results
						retval = activateCmd(i, "testresults")
						m := make(map[string]string)
						m["testresults"] = retval
						var r StatusReply
						kvname := "Test Results"
						PostExtendedStatusAndGetReply(i, "DONE", &r, &kvname, &m)
						a.State = STATEDone // all done with this test

					case lower == "testing":
						// nothing to do, let it keep running

					case errResult.MatchString(lower): // regexp:  begins with error
						ulog("%s returns error: %s\n", filename, retval[6:])
						// TODO: if retryable... keep going, if not, report back BLOCKED

					default:
						ulog("*** ERROR: unexpected reply to 'teststatus' command from %s: %s\n", filename, retval)
					}
				}
			}

			count, possible := AppsAtOrBeyondState(STATEDone, true)
			ulog("%d of %d apps are in STATEDone\n", count, possible)
			if count == possible {
				// mark the apps as in the DONE state now...
				for i := 0; i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
					if i == me {
						continue
					}
					a = &envMap.Instances[envMap.ThisInst].Apps[i]
					if !a.IsTest {
						a.State = STATEDone
						var r StatusReply
						PostExtendedStatusAndGetReply(i, "DONE", &r, nil, nil)
					}
				}
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

	ulog("Orchestrator: Posted READY status to uhura. ReplyCode: %d\n", r.ReplyCode)
	PostExtendedStatusAndGetReply(envMap.ThisApp, "READY", &r, nil, nil) // tell Uhura we're ready
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

	PostExtendedStatusAndGetReply(envMap.ThisApp, "TEST", &r, nil, nil) // Tel UHURA we're moving to the TEST state
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
	PostExtendedStatusAndGetReply(envMap.ThisApp, "DONE", &r, nil, nil) // starting our state machine in the INIT state
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
	//whoAmI()
	ulog("I am instance %d, my name is %s, I am app index %d\n",
		envMap.ThisInst, envMap.Instances[envMap.ThisInst].InstName, envMap.ThisApp)
	ulog("I will listen for commands on port %d\n",
		envMap.Instances[envMap.ThisInst].Apps[envMap.ThisApp].UPort)
	envMap.Instances[envMap.ThisInst].Apps[envMap.ThisApp].State = STATEInitializing
	var r StatusReply
	go UhuraComms()                                                     // handle anything that comes from uhura
	PostExtendedStatusAndGetReply(envMap.ThisApp, "INIT", &r, nil, nil) // starting our state machine in the INIT state
	go StateOrchestrator(alldone)                                       // let the orchestrator handle it from here
}
