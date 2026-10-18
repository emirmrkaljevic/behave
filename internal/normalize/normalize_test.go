package normalize

import (
	"testing"

	"github.com/emirmrkaljevic/behave/internal/spec"
)

func TestBuiltins(t *testing.T) {
	n, err := New([]string{"ansi", "workdir", "cmd", "timestamp", "hexaddr"},
		[]spec.Rule{{Pattern: `v\d+\.\d+`, With: "vX"}})
	if err != nil {
		t.Fatal(err)
	}
	ctx := Context{WorkDir: "/tmp/behave-123", Command: "/home/u/bin/gojq"}
	cases := map[string]string{
		"\x1b[31mred\x1b[0m":                         "red",
		"gojq: error: cannot open /tmp/behave-123/x": "$CMD: error: cannot open x",
		"usage: gojq [OPTIONS]":                      "usage: $CMD [OPTIONS]",
		"at 2026-10-14T01:02:03Z and 10:11:12":       "at $TIME and $TIME",
		"4\t2026-10-17 08:18\t.":                     "4\t$TIME\t.",
		"-rw-r--r-- 1 u u 6 Oct 17 08:18 a.txt":      "-rw-r--r-- 1 u u 6 $TIME a.txt",
		"-rw-r--r-- 1 u u 6 Oct  7  2025 a.txt":      "-rw-r--r-- 1 u u 6 $TIME a.txt",
		"ptr 0xdeadbeef00 v1.7":                      "ptr $ADDR vX",
		"gojqx stays":                                "gojqx stays",
		"in $TMP: /tmp/behave-123":                   "in $TMP: $WORKDIR",
	}
	for in, want := range cases {
		if got := string(n.Apply([]byte(in), ctx)); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
	if _, err := New([]string{"nope"}, nil); err == nil {
		t.Fatal("want error")
	}
}

func TestCmdBareName(t *testing.T) {
	n, _ := New([]string{"cmd"}, nil)
	got := string(n.Apply([]byte("sort: option '--sort' requires an argument; resort"), Context{Command: "sort"}))
	if got != "$CMD: option '--sort' requires an argument; resort" {
		t.Fatal(got)
	}
}

func TestANSIStripsOSC8BothTerminators(t *testing.T) {
	n, err := New([]string{"ansi"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	bel := "\x1b]8;;file://h/x/a.txt\x07a.txt\x1b]8;;\x07\n"
	st := "\x1b]8;;file://h/x/a.txt\x1b\\a.txt\x1b]8;;\x1b\\\n"
	for _, in := range []string{bel, st} {
		if got := string(n.Apply([]byte(in), Context{})); got != "a.txt\n" {
			t.Errorf("%q -> %q", in, got)
		}
	}
}
