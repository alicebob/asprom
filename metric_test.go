package main

import (
	"reflect"
	"testing"
)

func TestParseInfo(t *testing.T) {
	for k, v := range map[string]map[string]string{
		"type=memory": map[string]string{
			"type": "memory",
		},
		"type=memory;objects=103;sub-objects=0;master-objects=103;master-sub-objects=0;prole-objects=0;prole-sub-objects=0": map[string]string{
			"type":               "memory",
			"objects":            "103",
			"sub-objects":        "0",
			"master-objects":     "103",
			"master-sub-objects": "0",
			"prole-objects":      "0",
			"prole-sub-objects":  "0",
		},
		"version=my=version": map[string]string{
			"version": "my=version",
		},
	} {
		if have, want := parseInfo(k), v; !reflect.DeepEqual(have, want) {
			t.Errorf("have %+v, want %+v", have, want)
		}
	}
}
