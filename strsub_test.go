// Test the string substitution code
package main

import "testing"

func TestStrSub(t *testing.T) {
	readEnvDescr("./test/utdata/strsubdata.json")
	sraw := "echosrv_test -h {esrv.HostName} -p {esrv.echosrv.UPort}"
	expect := "echosrv_test -h esrvmain.ec2.amazon.com -p 8200"
	got := envDescrSub(sraw)
	if got != expect {
		t.Errorf("envDescrSub error:\n\texpected: %s\n\tgot: %s\n", expect, got)
	}
}
