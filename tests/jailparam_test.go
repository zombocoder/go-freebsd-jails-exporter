package tests

import (
	"testing"

	"github.com/zombocoder/go-freebsd-jails-exporter/internal/freebsd"
)

func TestParseJailBool(t *testing.T) {
	cases := map[string]bool{
		"true":  true, // what jailparam_export(3) returns for boolean params
		"1":     true,
		"false": false,
		"0":     false,
		"":      false,
	}
	for in, want := range cases {
		if got := freebsd.ParseJailBool(in); got != want {
			t.Errorf("ParseJailBool(%q) = %v, want %v", in, got, want)
		}
	}
}
