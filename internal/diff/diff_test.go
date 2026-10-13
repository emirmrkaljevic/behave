package diff

import (
	"strings"
	"testing"

	"github.com/emirmrkaljevic/behave/internal/spec"
)

func r(exit int, out, errs string) *spec.Result {
	return &spec.Result{Exit: exit, Stdout: spec.NewBlob([]byte(out)), Stderr: spec.NewBlob([]byte(errs))}
}

func TestCompare(t *testing.T) {
	if ms := Compare(r(0, "a\n", ""), r(0, "a\n", ""), Options{}); ms != nil {
		t.Fatal(ms)
	}
	ms := Compare(r(0, "a\nb\nc\n", "x"), r(1, "a\nB\nc\n", "y"), Options{IgnoreStderr: true})
	if len(ms) != 2 || ms[0].Field != "exit" || ms[1].Field != "stdout" {
		t.Fatalf("%+v", ms)
	}
	if !strings.Contains(ms[1].Text, "- b\n+ B\n") {
		t.Fatal(ms[1].Text)
	}
	if !strings.Contains(Lines("a\n", "a"), "⏎̸") {
		t.Fatal("trailing newline not marked")
	}
}

func TestBinaryRendering(t *testing.T) {
	got := Lines("ok\n", "ok\n\xff\x00\r\n")
	if !strings.Contains(got, `+ \xff\x00\r`) {
		t.Fatal(got)
	}
}
