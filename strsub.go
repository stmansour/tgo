package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// strDoEnvSub replaces string references to the internal environment with their
// values. Input format:
//		{<InstName>.<PropName>}   or
//      {<InstName>.<AppUID>.<PropName>}
// Examples
//		{eserv.HostName}		// example return val: "esrvmain.ec2.amazon.com"
//		{eserv.echosrv.UPort}	// example return val: "8200"
func strDoEnvSub(p string) string {
	s := strings.Trim(p, "{}")
	t := strings.Split(s, ".")
	lt := len(t)
	if lt < 2 || lt > 3 {
		return ""
	}

	// search for instance
	iid := -1
	for i := 0; i < len(envMap.Instances); i++ {
		if t[0] == envMap.Instances[i].InstName {
			iid = i
			break
		}
	}
	if iid < 0 {
		return ""
	}

	// look in the Instances structure for the field identified by t[1]
	if lt == 2 {
		r := reflect.ValueOf(&envMap.Instances[iid]).Elem()
		typeOfR := r.Type()
		for i := 0; i < r.NumField(); i++ {
			if typeOfR.Field(i).Name == t[1] {
				return fmt.Sprintf("%v", r.Field(i).Interface())
			}
		}
	}

	// look in the Instances[iid] for the App identified by the UID in t[1]
	// and the field identified by t[2]
	if lt == 3 {
		aid := -1
		for i := 0; i < len(envMap.Instances[iid].Apps); i++ {
			if t[1] == envMap.Instances[iid].Apps[i].UID {
				aid = i
				break
			}
		}
		if aid > -1 {
			r := reflect.ValueOf(&envMap.Instances[iid].Apps[aid]).Elem()
			typeOfR := r.Type()
			for i := 0; i < r.NumField(); i++ {
				if typeOfR.Field(i).Name == t[2] {
					return fmt.Sprintf("%v", r.Field(i).Interface())
				}
			}
		}
	}

	return ""
}

func envDescrSub(s string) string {
	rsub := regexp.MustCompile("({[^}]+})")
	t := strings.Split(s, " ")
	for i := 0; i < len(t); i++ {
		p := rsub.FindString(t[i])
		if "" != p {
			t[i] = strDoEnvSub(p)
		}
	}
	return strings.Join(t, " ")
}
