package main

import (
	"reflect"
	"testing"
)

func TestParseLatency(t *testing.T) {
	type cas struct {
		lat   string
		error string
		want  map[string]map[string]float64
	}
	for _, c := range []cas{
		{
			lat: "error-no-data-yet-or-back-too-small;error-no-data-yet-or-back-too-small;{sys}-read:15:26:23-GMT,ops/sec,>1ms,>8ms,>64ms;15:26:33,54.4,1.10,0.55,0.00;{test}-write:15:26:23-GMT,ops/sec,>1ms,>8ms,>64ms;15:26:33,4.0,0.00,0.00,0.00;error-no-data-yet-or-back-too-small;{sys}-query:15:26:23-GMT,ops/sec,>1ms,>8ms,>64ms;15:26:33,0.2,100.00,0.00,0.00",
			want: map[string]map[string]float64{
				"{sys}-read": map[string]float64{
					"ops/sec": 54.4,
					">1ms":    1.10,
					">8ms":    0.55,
					">64ms":   0.0,
				},
				"{test}-write": map[string]float64{
					"ops/sec": 4.0,
					">1ms":    0.0,
					">8ms":    0.0,
					">64ms":   0.0,
				},
				"{sys}-query": map[string]float64{
					"ops/sec": 0.2,
					">1ms":    100.00,
					">8ms":    0.0,
					">64ms":   0.0,
				},
			},
		},
		{
			lat: "error-no-data-yet-or-back-too-small;error-no-data-yet-or-back-too-small;error-no-data-yet-or-back-too-small;error-no-data-yet-or-back-too-small;{rz}-read:12:33:49-GMT,ops/sec,>1ms,>4ms,>8ms,>16ms,>32ms,>64ms,>128ms,>256ms,>512ms,>1024ms,>2048ms,>4096ms,>8192ms;12:33:59,0.4,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00;error-no-data-yet-or-back-too-small;{rz}-udf:12:33:49-GMT,ops/sec,>1ms,>4ms,>8ms,>16ms,>32ms,>64ms,>128ms,>256ms,>512ms,>1024ms,>2048ms,>4096ms,>8192ms;12:33:59,0.5,20.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00;error-no-data-yet-or-back-too-small",
			want: map[string]map[string]float64{
				"{rz}-read": map[string]float64{
					"ops/sec": 0.4,
					">1ms":    0.0,
					">4ms":    0.0,
					">8ms":    0.0,
					">16ms":   0.0,
					">32ms":   0.0,
					">64ms":   0.0,
					">128ms":  0.0,
					">256ms":  0.0,
					">512ms":  0.0,
					">1024ms": 0.0,
					">2048ms": 0.0,
					">4096ms": 0.0,
					">8192ms": 0.0,
				},
				"{rz}-udf": map[string]float64{
					"ops/sec": 0.5,
					">1ms":    20.0,
					">4ms":    0.0,
					">8ms":    0.0,
					">16ms":   0.0,
					">32ms":   0.0,
					">64ms":   0.0,
					">128ms":  0.0,
					">256ms":  0.0,
					">512ms":  0.0,
					">1024ms": 0.0,
					">2048ms": 0.0,
					">4096ms": 0.0,
					">8192ms": 0.0,
				},
			},
		},
		{
			lat:   "error-no-data-yet-or-back-too-small;error-no-data-yet-or-back-too-small;{sys}-read:15:26:23-GMT,ops/sec,>1ms,>8ms,>64ms",
			error: "latency: missing measurements line",
		},
	} {
		res, err := parseLatency(c.lat)
		haveerr := ""
		if err != nil {
			haveerr = err.Error()
		}
		if have, want := haveerr, c.error; have != want {
			t.Errorf("have %q, want %q", have, want)
			continue
		}
		if have, want := res, c.want; !reflect.DeepEqual(have, want) {
			t.Errorf("have %+v, want %+v", have, want)
		}
	}
}
