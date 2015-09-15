package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

type ct struct {
	tcase    int
	httpResp int
	sm       StatusMsg
	ur       StatusReply
}

var tgont struct {
	curTest int // index of currently executing test
}

//  Expected responses when the is started with the environment
//  descriptor env1.  These Tests can be used for unit testing
//  as well as functional testing
var Tests = []ct{
	// test#  http	StatusMsg								          Expected StatusReply
	ct{1, 200, StatusMsg{"INIT", "MainTestInstance", "wprog2", "x"}, StatusReply{"x", RespNoSuchInstance, "x"}},
	ct{1, 200, StatusMsg{"INIT", "MainTestInstance", "prog2", "x"}, StatusReply{"x", InvalidState, "x"}},
	ct{1, 200, StatusMsg{"INIT", "MainTestInstance", "prog2", "x"}, StatusReply{"x", RespOK, "x"}},
	ct{1, 200, StatusMsg{"INIT", "MainWinInstance", "wprog2", "x"}, StatusReply{"x", RespOK, "x"}},
	ct{1, 200, StatusMsg{"READY", "MainTestInstance", "prog2", "x"}, StatusReply{"x", RespOK, "x"}},
	ct{1, 200, StatusMsg{"READY", "MainWinInstance", "wprog2", "x"}, StatusReply{"x", RespOK, "x"}},
	ct{1, 200, StatusMsg{"TEST", "MainTestInstance", "prog2", "x"}, StatusReply{"x", RespOK, "x"}},
	ct{1, 200, StatusMsg{"TEST", "MainWinInstance", "wprog2", "x"}, StatusReply{"x", RespOK, "x"}},
	ct{1, 200, StatusMsg{"DONE", "MainTestInstance", "prog2", "x"}, StatusReply{"x", RespOK, "x"}},
	ct{1, 200, StatusMsg{"DONE", "MainWinInstance", "wprog2", "x"}, StatusReply{"x", RespOK, "x"}},
}

func setup() {
	var err error
	Tgo.LogFile, err = os.OpenFile("tgo.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer Tgo.LogFile.Close()
}

func UhuraStatusHandler(w http.ResponseWriter, r *http.Request) {
	sr := StatusReply{"x", Tests[tgont.curTest].ur.ReplyCode, time.Now().Format(time.RFC822)}
	w.Header().Add("Content-Type", "application/json")
	b, _ := json.Marshal(sr)
	fmt.Fprint(w, string(b))
}

func TestSendStatus(t *testing.T) {
	initTgo() // do this just once
	for i := 0; i < len(Tests); i++ {
		tgont.curTest = i
		Tests[i].sm.Tstamp = time.Now().Format(time.RFC822)
		TgoNetTest(t, Tests[i].httpResp, &Tests[i].sm, &Tests[i].ur)
	}
}

// Provide a bad instance / uid combination
func TgoNetTest(t *testing.T, expectHttpResp int, sm *StatusMsg, urexpect *StatusReply) {
	ts := httptest.NewServer(http.HandlerFunc(UhuraStatusHandler))
	defer ts.Close()
	Tgo.UhuraURL = ts.URL + "/"

	// Call PostStatus and let's see what we get back
	var ur StatusReply
	httpResp, err := PostStatus(sm, &ur)
	if nil != err {
		t.Errorf("Error returned from PostStatus: %v", err)
	}

	// Verify response
	if httpResp != expectHttpResp {
		t.Errorf("Bad HTTP response code.  Expected %d,  got %d\n", expectHttpResp, httpResp)
		return
	}
	if ur.ReplyCode != urexpect.ReplyCode {
		t.Errorf("Bad ReplyCode.  Expected %d,  got %d\n", urexpect.ReplyCode, ur.ReplyCode)
		return
	}
}
