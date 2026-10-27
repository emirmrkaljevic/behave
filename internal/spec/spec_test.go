package spec

import (
	"bytes"
	"os"
	"path/filepath"
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

func TestBlobEdgeRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := &Spec{Dir: dir, Meta: Meta{Command: "x"}}
	s.SaveMeta()
	vals := []string{"\n", "\n\n", "\na\n", "a\n\n\n", "a\n", "a", "", "  a\n", "a \nb\n", "a\r\n", "x\ty\n", "a\nb", "a\n\n"}
	for i, v := range vals {
		c := &Case{Name: "v", Expect: &Result{Stdout: NewBlob([]byte(v)), Stderr: NewBlob(nil)}}
		c.File = dir + "/cases/" + string(rune('a'+i)) + ".yaml"
		if err := s.SaveCase(c); err != nil {
			t.Fatal(err)
		}
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	for i, c := range got.Cases {
		if string(c.Expect.Stdout.Bytes()) != vals[i] {
			t.Errorf("%q round-tripped to %q", vals[i], c.Expect.Stdout.Bytes())
		}
	}
}

func TestSaveCaseCreatesCasesDir(t *testing.T) {
	dir := t.TempDir()
	// behave.yaml present, but no cases/ yet — mirrors hand-written bootstrap for `behave add`
	if err := os.WriteFile(filepath.Join(dir, "behave.yaml"), []byte("command: fold\ntimeout_ms: 5000\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	c := &Case{Name: "basic", Args: []string{"-w", "10"}, Tags: []string{"manual"}, Expect: &Result{Exit: 0, Stdout: NewBlob([]byte("hi\n")), Stderr: NewBlob(nil)}}
	if err := s.SaveCase(c); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "cases")); err != nil {
		t.Fatalf("cases/ not created: %v", err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Cases) != 1 || got.Cases[0].Name != "basic" {
		t.Fatalf("reload: %+v", got.Cases)
	}
}

func TestFileContentsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := &Spec{Dir: dir, Meta: Meta{Command: "x"}}
	s.SaveMeta()
	files := FileMap{"tab": "\tc\n", "mixed": "a\tb\n", "lead": "  x\ny\n", "crlf": "a\r\n", "plain": "a\nb\n", "empty": ""}
	c := &Case{Name: "files", Files: files, Expect: &Result{Stdout: NewBlob(nil), Stderr: NewBlob(nil)}}
	if err := s.SaveCase(c); err != nil {
		t.Fatal(err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range files {
		if got.Cases[0].Files[k] != v {
			t.Errorf("%s: %q round-tripped to %q", k, v, got.Cases[0].Files[k])
		}
	}
}
