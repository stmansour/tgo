package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// StatusMsg is status message structure we use to
// communicate with uhura.
type StatusMsg struct {
	State    string
	InstName string
	UID      string
	Tstamp   string
}

// StatusReply represents the structure of information
// that Uhura returns in response to TGO's status messages.
type StatusReply struct {
	Status    string
	ReplyCode int
	Timestamp string
}

// UCommand is the structure of commands that Uhura sends TGO
type UCommand struct {
	Command   string
	CmdCode   int
	Timestamp string
}

// RespOK and the rest are meaningful names associated with the
// StatusReply.ReplyCode that uhura sends TGO
const (
	RespOK             = iota // 0
	RespNoSuchInstance        // 1
	InvalidState              // 2
	RespBadCmd                // 3
	RespInvalidState          // 4
)

// PostStatus is used to send a status message to uhura
// returns the HTTP statuscode of the response and the error
// from the http.POST
func PostStatus(sm *StatusMsg, r *StatusReply) (int, error) {
	b, err := json.Marshal(sm)
	if err != nil {
		ulog("Cannot marshal status struct! Error: %v\n", err)
		os.Exit(2) // no recovery from this
	}
	req, err := http.NewRequest("POST", envMap.UhuraURL+"status/", bytes.NewBuffer(b))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ulog("Cannot Post status message! Error: %v\n", err)
		return 0, err // ?? maybe there's some retry we can do??
	}
	defer resp.Body.Close()

	// pull out the HTTP response code
	var rc int
	var more string
	fmt.Sscanf(resp.Status, "%d %s", &rc, &more)

	// body, _ := ioutil.ReadAll(resp.Body)
	// ulog("raw reply data: %s\n", string(body))
	// json.Unmarshal(body, r)
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(r); err != nil {
		panic(err)
	}
	return rc, err
}

// SendReply sends a response back to uhura.
func SendReply(w http.ResponseWriter, rc int, s string) {
	w.Header().Set("Content-Type", "application/json")
	m := StatusReply{Status: s, ReplyCode: rc, Timestamp: time.Now().Format(time.RFC822)}
	str, err := json.Marshal(m)
	if nil != err {
		fmt.Fprintf(w, "{\n\"Status\": \"%s\"\n\"Timestamp:\": \"%s\"\n}\n",
			"encoding error", time.Now().Format(time.RFC822))
	} else {
		fmt.Fprintf(w, string(str))
	}
}

// CommsHandler handles any incoming network communication. We
// only expect to receive commands from Uhura.
func CommsHandler(w http.ResponseWriter, r *http.Request) {
	ulog("Comms Handler\n")
	var s UCommand
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&s); err != nil {
		ulog("CommsHandler could not decode message:\n%s\nWill ignore this message\n", s)
		SendReply(w, 0, "Undecodable Message")
		return
	}

	ulog("Received comms from Uhura:  %+v\n", s)
	switch {
	case s.Command == "TESTNOW":
		SendReply(w, RespOK, "OK")
		Tgo.UhuraComm <- s.CmdCode // tell the state machine to proceed
		<-Tgo.UhuraComm            // wait til the handler says it's ok to proceed
	default:
		ulog("Received unknown cmd from Uhura: %+v", s)
		SendReply(w, RespBadCmd, "BADCMD")
	}
}

// UhuraComms sets up the handlers for any commands that Uhura sends this
// TGO instance. The main thing Uhura contacts us about is to
// notify us when testing can begin.
func UhuraComms() {
	// Set up an http service that listens on our assigned
	// port for any messages
	http.HandleFunc("/", CommsHandler)
	s := fmt.Sprintf(":%d",
		envMap.Instances[envMap.ThisInst].Apps[envMap.ThisApp].UPort)
	ulog("UhuraComms http service listening on port: %d\n",
		envMap.Instances[envMap.ThisInst].Apps[envMap.ThisApp].UPort)
	go http.ListenAndServe(s, nil)
}
