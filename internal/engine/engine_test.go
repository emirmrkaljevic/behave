package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/emirmrkaljevic/behave/internal/diff"
	"github.com/emirmrkaljevic/behave/internal/normalize"
	"github.com/emirmrkaljevic/behave/internal/spec"
)

// script writes an executable shell script and returns its path.
func script(t *testing.T, dir, name, body string) string {
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRecordAndDiff(t *testing.T) {
	dir := t.TempDir()
	orig := script(t, dir, "upper", `if [ "$1" = "--bad" ]; then echo "$0: bad flag" >&2; exit 2; fi; tr a-z A-Z; date +%s%N`+"\n")
	clone := script(t, dir, "upper2", `if [ "$1" = "--bad" ]; then echo "$0: bad flag" >&2; exit 1; fi; tr a-z A-Z; date +%s%N`+"\n")
	_ = clone
	stable := script(t, dir, "stable", `if [ "$1" = "--bad" ]; then echo "$0: bad flag" >&2; exit 2; fi; tr a-z A-Z`+"\n")
	clone2 := script(t, dir, "stable2", `if [ "$1" = "--bad" ]; then echo "$0: bad flag" >&2; exit 1; fi; tr a-z A-Z`+"\n")

	s := &spec.Spec{Dir: filepath.Join(dir, "spec"), Meta: spec.Meta{Command: stable, Normalize: normalize.Default}}
	if err := s.SaveMeta(); err != nil {
		t.Fatal(err)
	}
	for _, c := range []*spec.Case{
		{Name: "upper", Args: []string{}, Stdin: spec.NewBlob([]byte("abc\n"))},
		{Name: "bad", Args: []string{"--bad"}},
	} {
		s.SaveCase(c)
	}
	if st, err := Record(s, false); err != nil || st.Recorded != 2 || st.Flaky != 0 {
		t.Fatalf("%+v %v", st, err)
	}
	s, _ = spec.Load(s.Dir)
	if got := string(s.Cases[1].Expect.Stderr.Bytes()); got != "$CMD: bad flag\n" {
		t.Fatalf("stderr not normalized: %q", got)
	}
	rep, err := Replay(s, "", diff.Options{})
	if err != nil || rep.Pass != 2 {
		t.Fatalf("self replay: %+v %v", rep, err)
	}
	rep, _ = Replay(s, clone2, diff.Options{})
	if rep.Pass != 1 || rep.Fail != 1 || rep.ByField["exit"] != 1 || rep.ByField["stderr"] != 0 {
		t.Fatalf("clone replay: %+v", rep)
	}

	// a non-deterministic command is marked flaky and skipped
	s2 := &spec.Spec{Dir: filepath.Join(dir, "spec2"), Meta: spec.Meta{Command: orig, Normalize: normalize.Default}}
	s2.SaveMeta()
	s2.SaveCase(&spec.Case{Name: "time", Args: []string{}})
	if st, _ := Record(s2, false); st.Flaky != 1 {
		t.Fatalf("flaky not detected: %+v", st)
	}
	if rep, _ := Replay(s2, "", diff.Options{}); rep.Skip != 1 {
		t.Fatalf("%+v", rep)
	}
}

func TestRecordStoresVersionLine(t *testing.T) {
	dir := t.TempDir()
	v1 := script(t, dir, "tool", `if [ "$1" = "--version" ]; then printf '\ntool (demo) 1.2\nmore text\n'; exit 0; fi; cat`+"\n")
	v2 := script(t, dir, "tool2", `if [ "$1" = "--version" ]; then echo "tool (rewrite) 0.1"; exit 0; fi; cat`+"\n")
	none := script(t, dir, "plain", `[ "$1" = "--version" ] && exit 1; cat`+"\n")

	if got := VersionLine(v1); got != "tool (demo) 1.2" {
		t.Fatalf("VersionLine = %q", got)
	}
	if got := VersionLine(none); got != "" {
		t.Fatalf("no --version should give empty, got %q", got)
	}

	s := &spec.Spec{Dir: filepath.Join(dir, "spec"), Meta: spec.Meta{Command: v1}}
	if err := s.SaveMeta(); err != nil {
		t.Fatal(err)
	}
	s.SaveCase(&spec.Case{Name: "echo", Args: []string{}, Stdin: spec.NewBlob([]byte("x\n"))})
	if _, err := Record(s, false); err != nil {
		t.Fatal(err)
	}
	s, _ = spec.Load(s.Dir)
	if s.Meta.RecordedFrom != "tool (demo) 1.2" {
		t.Fatalf("recorded_from = %q", s.Meta.RecordedFrom)
	}
	rep, err := Replay(s, v2, diff.Options{})
	if err != nil || rep.RecordedFrom != "tool (demo) 1.2" || rep.Version != "tool (rewrite) 0.1" || rep.Pass != 1 {
		t.Fatalf("%+v %v", rep, err)
	}
}
