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
		"\x1b[31mred\x1b[0m":                     "red",
		"gojq: error: cannot open /tmp/behave-123/x": "$CMD: error: cannot open x",
		"usage: gojq [OPTIONS]":                  "usage: $CMD [OPTIONS]",
		"at 2026-10-14T01:02:03Z and 10:11:12":   "at $TIME and $TIME",
		"ptr 0xdeadbeef00 v1.7":                  "ptr $ADDR vX",
		"gojqx stays":                            "gojqx stays",
		"in $TMP: /tmp/behave-123":               "in $TMP: $WORKDIR",
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
