package spec

import (
	"bytes"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := &Spec{Dir: dir, Meta: Meta{Command: "cat"}}
	if err := s.SaveMeta(); err != nil {
		t.Fatal(err)
	}
	bin := []byte{0xff, 0x00, 'a'}
	c := &Case{Name: "Binary & text!", Args: []string{"-n"}, Stdin: NewBlob([]byte("a\nb\n")),
		Expect: &Result{Exit: 1, Stdout: NewBlob(bin), Stderr: NewBlob([]byte("oops "))}}
	if err := s.SaveCase(c); err != nil {
		t.Fatal(err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Cases) != 1 {
		t.Fatalf("cases=%d", len(got.Cases))
	}
	g := got.Cases[0]
	if !bytes.Equal(g.Expect.Stdout.Bytes(), bin) || string(g.Stdin.Bytes()) != "a\nb\n" ||
		string(g.Expect.Stderr.Bytes()) != "oops " || g.Expect.Exit != 1 {
		t.Fatalf("mismatch: %+v", g)
	}
	if Slug("Binary & text!") != "binary-text" {
		t.Fatal(Slug("Binary & text!"))
	}
}
