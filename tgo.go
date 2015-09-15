package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
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
	UhuraURL      string
	LogFile       *os.File
	Port          int  // What port are we listening on
	Debug         bool // Debug mode -- show ulog messages on screen
	DebugToScreen bool // Send logging info to screen too
	IntFuncTest   bool // internal functional test mode
}

// Tgo is the instance of TGOApp for this application
var Tgo TGOApp

func processCommandLine() {
	dbugPtr := flag.Bool("d", false, "debug mode - includes debug info in logfile")
	dtscPtr := flag.Bool("D", false, "LogToScreen mode - prints log messages to stdout")
	itstPtr := flag.Bool("F", false, "Internal Functional Test mode")
	flag.Parse()
	Tgo.Debug = *dbugPtr
	Tgo.DebugToScreen = *dtscPtr
	Tgo.IntFuncTest = *itstPtr
}

func initTgo() {
	// Read phonehome to find out the address of uhura
	content, e := ioutil.ReadFile("phonehome")
	if e != nil {
		ulog("Cannot read phonehome file! Error: %v\n", e)
		os.Exit(1) // no recovery from this
	}
	s := string(content)

	Tgo.UhuraURL = strings.TrimRight(s, "\n\r")
	ulog("**********   T G O   **********\n")
	ulog("UhuraURL = %s\n", Tgo.UhuraURL)
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

	// OK, now on with the show...
	processCommandLine()
	initTgo()

	fmt.Printf("Tgo.IntFuncTest = %v\n", Tgo.IntFuncTest)

	switch {
	case Tgo.IntFuncTest:
		IntFuncTest0()
	default:
	}
}
