package main

import (
	"reflect"
	"testing"
)

func TestParseLatency(t *testing.T) {
	lat := "error-no-data-yet-or-back-too-small;error-no-data-yet-or-back-too-small;{sys}-read:15:26:23-GMT,ops/sec,>1ms,>8ms,>64ms;15:26:33,54.4,1.10,0.55,0.00;{test}-write:15:26:23-GMT,ops/sec,>1ms,>8ms,>64ms;15:26:33,4.0,0.00,0.00,0.00;error-no-data-yet-or-back-too-small;{sys}-query:15:26:23-GMT,ops/sec,>1ms,>8ms,>64ms;15:26:33,0.2,100.00,0.00,0.00"
	want := map[string]map[string]float64{
		"sys": {
			"read>1ms":   1.10,
			"read>8ms":   0.55,
			"read>64ms":  0.0,
			"query>1ms":  100.00,
			"query>8ms":  0.0,
			"query>64ms": 0.0,
		},
		"test": {
			"write>1ms":  0.0,
			"write>8ms":  0.0,
			"write>64ms": 0.0,
		},
	}
	if have := parseLatency(lat); !reflect.DeepEqual(have, want) {
		t.Errorf("have %+v, want %+v", have, want)
	}
}
