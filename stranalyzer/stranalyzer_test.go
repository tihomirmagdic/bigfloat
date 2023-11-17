package stranalyzer

import (
	"fmt"
	"testing"
)

func TestAnalyze(t *testing.T) {
	cases := []struct {
		in        string
		wantError bool
	}{
		{"0", false},
		{"01", false},
		{"-01", false},
		{"00.1", false},
		{"00.001", false},
		{".001", false},
		{"-1", false},
		{"-8000.003", false},
		{"+5.000002", false},
		{"1E+5", false},
		{"1E-4", false},
		{"-5E4", false},
		{"-123E3", false},
		{"-42e-5", false},
		{"123.", false},
		{".123", false},
		{"1.2e+5", false},
		{"1.2e-1", false},
		{".01", false},
		{"1e1", false},
		{"1e-1", false},
		{"   +  1   2 .  221", false},
		{"+.0", false},
		{"   .+", true},
		{"-123E3.1", true},
		{" 1 + ", true},
		{"-1+", true},
		{"1e", true},
		{"1e-", true},
		{"..", true},
		{"e", true},
		{"e1", true},
		{"-.", true},
		{".", true},
		{"+.", true},
		{"1e5e1", true},
		{"1e+5+", true},
		{"	1", false},
		{"1a", true},
		{"-0", false},
		{"1.5e-2", false},
		{"1.5e+3", false},
		{"1.52e+1", false},
	}
	for _, c := range cases {
		a, error := Analyze(c.in)
		if error != nil {
			fmt.Printf("%q: error: %s\n", c.in, error)
		} else {
			fmt.Printf("%q: norm: %q, sign: %d, decimals: %d, len: %d\n", c.in, a.Norm, a.Sign, a.Decimals, a.Len)
			if a.Len != len(a.Norm) {
				t.Errorf("Wrong Len")
			} else if a.Decimals > a.Len {
				t.Errorf("Wrong Decimals")
			} else if a.Sign < -1 || a.Sign > 1 {
				t.Errorf("Wrong Sign")
			}
		}
		if c.wantError && error == nil {
			t.Errorf("Analyze: should be error for %q", c.in)
		} else if !c.wantError && error != nil {
			t.Errorf("Analyze: %q", error)
		}
	}
}
