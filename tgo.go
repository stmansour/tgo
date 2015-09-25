package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

//  This is the TGO program. The Task Group Organizer.
//  this process is responsible for launching and
//  managing the apps on its system and communicating
//  with uhura.

//  Duties:
//  1. locate the phonehome file, contact uhura with
//     a status message, confirming that it is now in
//     the INIT phase.

// TGOApp is the structure definit of data used globally by this
// application. It holds many basic but critical values
// needed by most of the application.
type TGOApp struct {
	State         int
	LogFile       *os.File
	UhuraComm     chan int // communications from Uhura
	Port          int      // What port are we listening on
	Debug         bool     // Debug mode -- show ulog messages on screen
	DebugToScreen bool     // Send logging info to screen too
	IntFuncTest   bool     // internal functional test mode
}

// Tgo is the instance of TGOApp for this application
var Tgo TGOApp

// OK, this is a major cop-out, but not sure what else to do...
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func processCommandLine() {
	dbugPtr := flag.Bool("d", false, "debug mode - includes debug info in logfile")
	dtscPtr := flag.Bool("D", false, "LogToScreen mode - prints log messages to stdout")
	itstPtr := flag.Bool("F", false, "Internal Functional Test mode")
	flag.Parse()
	Tgo.Debug = *dbugPtr
	Tgo.DebugToScreen = *dtscPtr
	Tgo.IntFuncTest = *itstPtr
}

func readEnvDescr(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		ulog("no such file or directory: %s\n", filename)
		envMap.UhuraURL = "http://localhost:8100/"
		ulog("assuming test mode: UhuraURL = %s\n", envMap.UhuraURL)
		return
	}

	content, e := ioutil.ReadFile(filename)
	if e != nil {
		ulog("File error on %s: %#v\n", filename, e)
		os.Exit(1) // no recovery from this
	}
	ulog("%s\n", string(content))

	// OK, now we have the json describing the environment in content (a string)
	// Parse it into an internal data structure...
	err := json.Unmarshal(content, &envMap)
	if err != nil {
		ulog("Error unmarshaling Environment Descriptor json: %s\n", err)
		check(err)
	}
}

func whoAmI() {
	filename := "uhura_map.json"
	readEnvDescr(filename)
	ulog("ParseEnvDescriptor - Loading %s\n", filename)
	// DPrintEnvDescr("envMap after initial parse:")
	ulog("uhura url: %s\n", envMap.UhuraURL)

	// Uhura tells us which instance we are, but it does not look up the app
	// and tell us which app instance. So we look it up here...
	var found bool
	for i := 0; i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
		if envMap.Instances[envMap.ThisInst].Apps[i].Name == "tgo" {
			envMap.ThisApp = i
			found = true
		}
	}
	if !found {
		ulog("*** NOTICE ***  did not find tgo in uhura_map.json instance %d\n", envMap.ThisInst)
	}
}

func initTgo() {
	ulog("**********   T G O   **********\n")
	whoAmI()
	Tgo.UhuraComm = make(chan int)
}

// This is uhura's standard loger
func ulog(format string, a ...interface{}) {
	p := fmt.Sprintf(format, a...)
	log.Print(p)
	if Tgo.DebugToScreen {
		fmt.Print(p)
	}
}

func main() {
	// Let's get a log file going first.  If I put this file create in any other call
	// it seems to stop working after the call returns. Must be some sort of a scoping thing
	// that I don't understand. But for now, creating the logfile in the main() routine
	// seems to be the way to make it work.
	var err error
	Tgo.LogFile, err = os.OpenFile("tgo.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer Tgo.LogFile.Close()
	log.SetOutput(Tgo.LogFile)

	// OK, now on with the show...
	processCommandLine()
	initTgo()

	fmt.Printf("Tgo.IntFuncTest = %v\n", Tgo.IntFuncTest)

	switch {
	case Tgo.IntFuncTest:
		errcount := IntFuncTest0()
		ulog("IntFuncTest0 error count: %d\n", errcount)
	default:
		c := make(chan int)                        // a channel to signal us when it's all done
		InitiateStateMachine(c)                    // initiate and pass in the channel
		<-c                                        // wait til it's done
		time.Sleep(time.Duration(1 * time.Second)) // grace period, let everything finish
	}
}
