package main

import (
	"reflect"
	"testing"
)

func TestParseLatency(t *testing.T) {
	lat := "reads:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,12469.3,0.40,0.00,0.00;writes_master:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,0.0,0.00,0.00,0.00;proxy:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,0.0,0.00,0.00,0.00;udf:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,14730.7,0.03,0.00,0.00;query:19:21:58-GMT,ops/sec,>1ms,>8ms,>64ms;19:22:08,1.2,3.45,5.67,7.89;"
	want := map[string]map[string]float64{
		"reads": {
			"ops/sec": 12469.3,
			">1ms":    0.4,
			">8ms":    0.0,
			">64ms":   0.0,
		},
		"writes_master": {
			"ops/sec": 0.0,
			">1ms":    0.0,
			">8ms":    0.0,
			">64ms":   0.0,
		},
		"proxy": {
			"ops/sec": 0.0,
			">1ms":    0.0,
			">8ms":    0.0,
			">64ms":   0.0,
		},
		"udf": {
			"ops/sec": 14730.7,
			">1ms":    0.03,
			">8ms":    0.0,
			">64ms":   0.0,
		},
		"query": {
			"ops/sec": 1.2,
			">1ms":    3.45,
			">8ms":    5.67,
			">64ms":   7.89,
		},
	}
	if have := parseLatency(lat); !reflect.DeepEqual(have, want) {
		t.Errorf("have %+v, want %+v", have, want)
	}
}
