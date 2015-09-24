// Test reading the environment descriptors that uhura sends us
package main

import "testing"

func TestNewInstParse(t *testing.T) {
	ReadEnvDescr("./test/utdata/uhura_map.json")
	ulog("Number of instances: %d\n", len(EnvMap.Instances))
}
