package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// StatusMsg is status message structure we use to
// communicate with uhura.
type StatusMsg struct {
	State    string
	InstName string
	UID      string
	Tstamp   string
}

type StatusReply struct {
	Status    string
	ReplyCode int
	Timestamp string
}

const (
	RespOK = iota
	RespNoSuchInstance
	InvalidState
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
	req, err := http.NewRequest("POST", Tgo.UhuraURL+"status/", bytes.NewBuffer(b))
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

	body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("response Body:", string(body))
	// fmt.Println("response Headers:", resp.Header)
	json.Unmarshal(body, r)
	// fmt.Printf("Status Reply = %+v\n", r)

	return rc, err
}
