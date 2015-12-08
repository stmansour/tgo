package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func processAppResourceNeeds() {
	ulog("Entering processAppResourceNeeds\n")
	for i := 0; i < len(envMap.Instances[envMap.ThisInst].Apps); i++ {
		a := &envMap.Instances[envMap.ThisInst].Apps[i]

		if len(a.AppRes.cmd) > 0 {
			ulog("App[%d] requests cmd: %s\n", i, a.AppRes.cmd)
			//-----------------------------------------------------------------
			// switch to the directory containing the app
			//-----------------------------------------------------------------
			dirname := fmt.Sprintf("../%s", a.Name)
			if err := os.Chdir(dirname); err != nil {
				ulog("could not cd to %s:  %v\n", dirname, err)
			}
			ulog("cd to %s\n", dirname)

			m := strings.Split(a.AppRes.cmd, " ")
			c := m[0]
			args := make([]string, 0)
			for i := 1; i <= len(m[1:]); i++ {
				args = append(args, m[i])
			}

			ulog("c = %s,  args = %+v\n", c, args)

			cmd := exec.Command(c, args...)
			err := cmd.Start()
			if err != nil {
				panic(err)
			}
			cmd.Wait()

			ulog("%s ran. no errors reported\n", c)

			if err := os.Chdir("../tgo"); err != nil {
				ulog("could not cd to ../tgo:  %v\n", err)
			}
			ulog("cd back to ../tgo\n")

		} else if len(a.AppRes.DBname) > 0 { // this restores a test db.  It is deprecated. Legacy code.
			ulog("App[%d] requests restoredb:  db=%s, file=%s\n", i, a.AppRes.DBname, a.AppRes.RestoreMySQLdb)
			//-----------------------------------------------------------------
			// switch to the directory containing the sql commands we need...
			//-----------------------------------------------------------------
			dirname := fmt.Sprintf("../%s", a.Name)
			if err := os.Chdir(dirname); err != nil {
				ulog("could not cd to %s:  %v\n", dirname, err)
			}
			ulog("cd to %s\n", dirname)

			args := []string{
				a.AppRes.DBname,         // name of the database
				a.AppRes.RestoreMySQLdb, // name of file with sql commands
			}

			script := "/usr/local/accord/testtools/restoreMySQLdb.sh"
			_, err := os.Stat(script)
			if nil != err {
				script := "/c/Accord/testtools/restoreMySQLdb.sh"
				_, err := os.Stat(script)
				if nil != err {
					ulog("neither /c/Accord/testtools nor /usr/local/accord/testtools exist!!\nPlease check installation\n")
					check(err)
				}
			}

			ulog("script = %s, args = %+v\n", script, args)

			cmd := exec.Command(script, args...)
			err = cmd.Start()
			if err != nil {
				panic(err)
			}
			cmd.Wait()

			ulog("script ran. no errors reported\n")

			if err := os.Chdir("../tgo"); err != nil {
				ulog("could not cd to ../tgo:  %v\n", err)
			}
			ulog("cd back to ../tgo\n")
		}
	}
}
