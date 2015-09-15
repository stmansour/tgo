package main

import (
	"fmt"
	"os"
	"time"
)

type cft struct {
	tcase    int
	httpResp int
	sm       StatusMsg
	ur       StatusReply
}

//  Expected responses when the is started with the environment
//  descriptor env1.  These Tests can be used for unit testing
//  as well as functional testing
var tests = []cft{
	// test#  http	StatusMsg								          Expected StatusReply
	cft{1, 200, StatusMsg{"INIT", "MainTestInstance", "wprog2", "x"}, StatusReply{"x", RespNoSuchInstance, "x"}},
	cft{1, 200, StatusMsg{"YACK", "MainTestInstance", "prog2", "x"}, StatusReply{"x", InvalidState, "x"}},
	cft{1, 200, StatusMsg{"INIT", "MainTestInstance", "prog2", "x"}, StatusReply{"x", RespOK, "x"}},
	cft{1, 200, StatusMsg{"INIT", "MainWinInstance", "wprog2", "x"}, StatusReply{"x", RespOK, "x"}},
	cft{1, 200, StatusMsg{"READY", "MainTestInstance", "prog2", "x"}, StatusReply{"x", RespOK, "x"}},
	cft{1, 200, StatusMsg{"READY", "MainWinInstance", "wprog2", "x"}, StatusReply{"x", RespOK, "x"}},
	cft{1, 200, StatusMsg{"TEST", "MainTestInstance", "prog2", "x"}, StatusReply{"x", RespOK, "x"}},
	cft{1, 200, StatusMsg{"TEST", "MainWinInstance", "wprog2", "x"}, StatusReply{"x", RespOK, "x"}},
	cft{1, 200, StatusMsg{"DONE", "MainTestInstance", "prog2", "x"}, StatusReply{"x", RespOK, "x"}},
	cft{1, 200, StatusMsg{"DONE", "MainWinInstance", "wprog2", "x"}, StatusReply{"x", RespOK, "x"}},
}

// Runs a bunch of tests against a local uhura
// returns the number of failed tests.
func IntFuncTest0() int {
	var testFailCount int
	for i := 0; i < len(tests); i++ {
		tests[i].sm.Tstamp = time.Now().Format(time.RFC822)
		var r StatusReply
		rc, e := PostStatus(&tests[i].sm, &r)
		if nil != e {
			ulog("PostStatus returned error:  %v\n", e)
			os.Exit(5)
		}
		//fmt.Printf("http response code: %d,   Uhura Response: %#v\n", rc, r)
		// Verify response
		if rc != tests[i].httpResp {
			ulog("Bad HTTP response code.  Expected %d,  got %d\n", tests[i].httpResp, rc)
			testFailCount++
			break
		}
		if r.ReplyCode != tests[i].ur.ReplyCode {
			ulog("Bad ReplyCode.  Expected %d,  got %d\n", tests[i].ur.ReplyCode, r.ReplyCode)
			testFailCount++
			break
		}
		fmt.Printf("test %d PASSED\n", i)
	}
	return testFailCount
}
