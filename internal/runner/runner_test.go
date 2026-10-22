package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/emirmrkaljevic/behave/internal/spec"
)

func TestRunBasics(t *testing.T) {
	c := &spec.Case{Args: []string{"-c", "cat in.txt; cat; echo err >&2; exit 3"},
		Stdin: spec.NewBlob([]byte("B")), Files: map[string]string{"in.txt": "A"}}
	o, err := Run("sh", c, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	r := o.Result
	if string(r.Stdout.Bytes()) != "AB" || string(r.Stderr.Bytes()) != "err\n" || r.Exit != 3 {
		t.Fatalf("%q %q %d", r.Stdout.Bytes(), r.Stderr.Bytes(), r.Exit)
	}
}

func TestTimeout(t *testing.T) {
	c := &spec.Case{Args: []string{"-c", "sleep 5 & sleep 5"}, TimeoutMS: 200}
	start := time.Now()
	o, err := Run("sh", c, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !o.Result.TimedOut || time.Since(start) > 2*time.Second {
		t.Fatalf("timed_out=%v after %v", o.Result.TimedOut, time.Since(start))
	}
}

func TestMissingBinary(t *testing.T) {
	if _, err := Run("behave-no-such-binary", &spec.Case{}, time.Second); err == nil {
		t.Fatal("want error")
	}
}

func TestFilesGetFixedTimes(t *testing.T) {
	c := &spec.Case{Args: []string{"-c", `for f in b.txt a.txt d/c.txt d; do stat -c '%n %Y' "$f"; done`},
		Files: map[string]string{"b.txt": "b", "a.txt": "a", "d/c.txt": "c"}}
	out, err := Run("sh", c, 0)
	if err != nil {
		t.Fatal(err)
	}
	base := FileTime.Unix()
	want := fmt.Sprintf("b.txt %d\na.txt %d\nd/c.txt %d\nd %d\n", base+3600, base, base+7200, base)
	if got := string(out.Result.Stdout.Bytes()); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRelativeCommandPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tool"), []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	o, err := Run("./tool", &spec.Case{}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(o.Result.Stdout.Bytes()); got != "ok\n" {
		t.Fatalf("stdout %q", got)
	}
}
