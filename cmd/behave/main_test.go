package main

import (
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/emirmrkaljevic/behave/internal/spec"
)

func TestAddFlagsAfterSpecDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "spec")
	s := &spec.Spec{Dir: dir, Meta: spec.Meta{Command: "cat", TimeoutMS: 5000}}
	if err := s.SaveMeta(); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "cases"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := cmdAdd([]string{dir, "--name", "cat stdin", "--stdin", "hi\n", "--", "-"}); err != nil {
		t.Fatal(err)
	}
	s, err := spec.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Cases) != 1 {
		t.Fatalf("got %d cases", len(s.Cases))
	}
	c := s.Cases[0]
	if c.Name != "cat stdin" || !reflect.DeepEqual(c.Args, []string{"-"}) {
		t.Fatalf("name %q args %q", c.Name, c.Args)
	}
	if got := string(c.Expect.Stdout.Bytes()); got != "hi\n" {
		t.Fatalf("stdout %q", got)
	}
}

func TestParseDashDash(t *testing.T) {
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	stdin := fs.String("stdin", "", "")
	v := fs.Bool("v", false, "")
	pos, err := parse(fs, []string{"spec", "--stdin", "--", "-v", "--", "-d,", "-f2"})
	if err != nil {
		t.Fatal(err)
	}
	if *stdin != "--" || !*v {
		t.Fatalf("stdin %q v %v", *stdin, *v)
	}
	if want := []string{"spec", "--", "-d,", "-f2"}; !reflect.DeepEqual(pos, want) {
		t.Fatalf("pos %q, want %q", pos, want)
	}
}
