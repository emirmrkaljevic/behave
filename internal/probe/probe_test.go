package probe

import "testing"

const help = `Usage: wc [OPTION]... [FILE]...
  -c, --bytes            print the byte counts
  -m, --chars            print the character counts
      --files0-from=F    read input from the files specified by
  -L, --max-line-length  print the maximum display width
      --total=WHEN       when to print a line with total counts;
  -n NUM                 lines
      --help        display this help and exit
`

func TestParseHelp(t *testing.T) {
	fs := ParseHelp(help)
	got := map[string]Flag{}
	for _, f := range fs {
		got[f.Name()] = f
	}
	for name, val := range map[string]string{"--bytes": "", "--chars": "", "--files0-from": "F",
		"--max-line-length": "", "--total": "WHEN", "-n": "NUM", "--help": ""} {
		f, ok := got[name]
		if !ok {
			t.Errorf("missing %s (got %+v)", name, fs)
			continue
		}
		if f.Value != val {
			t.Errorf("%s value %q want %q", name, f.Value, val)
		}
	}
	if got["--bytes"].Short != "-c" {
		t.Errorf("short of --bytes: %q", got["--bytes"].Short)
	}
}

func TestGenerate(t *testing.T) {
	cs := Generate(ParseHelp(help), Options{FileInput: true, Max: 1000})
	if len(cs) < 50 {
		t.Fatalf("only %d cases", len(cs))
	}
	seen := map[string]bool{}
	for _, c := range cs {
		if seen[c.Name] {
			t.Errorf("duplicate name %q", c.Name)
		}
		seen[c.Name] = true
	}
	if n := len(Generate(ParseHelp(help), Options{Max: 10})); n != 10 {
		t.Fatalf("max not applied: %d", n)
	}
}

func TestLowercaseValue(t *testing.T) {
	fs := ParseHelp("  --indent n         Use n spaces\n  --tab              Use tabs\n  --arg name value   Set $name\n  -s, --slurp        Read all\n")
	want := map[string]string{"--indent": "n", "--tab": "", "--arg": "name", "--slurp": ""}
	for _, f := range fs {
		if w, ok := want[f.Name()]; ok && f.Value != w {
			t.Errorf("%s value %q want %q", f.Name(), f.Value, w)
		}
	}
	if len(fs) != 4 {
		t.Fatalf("%+v", fs)
	}
}

func TestOptionalValue(t *testing.T) {
	fs := ParseHelp("  -D, --all-repeated[=METHOD]  like -d\n      --color[=WHEN]   colorize\n")
	if len(fs) != 2 || !fs[0].Optional || fs[0].Value != "METHOD" || !fs[1].Optional {
		t.Fatalf("%+v", fs)
	}
	if a := flagArgs(fs[0], "x"); len(a) != 1 || a[0] != "--all-repeated=x" {
		t.Fatal(a)
	}
}

func TestRangeListNotFlags(t *testing.T) {
	h := `  -f, --fields=LIST       select only these fields
  -s, --only-delimited     do not print lines not containing delimiters

Each range is one of:

  N     N'th byte, character or field, counted from 1
  N-    from N'th byte, character or field, to end of line
  N-M   from N'th to M'th (included) byte, character or field
  -M    from first to M'th (included) byte, character or field

  -z    after the range list, a real short flag
`
	got := map[string]bool{}
	for _, f := range ParseHelp(h) {
		got[f.Name()] = true
	}
	if got["-M"] {
		t.Errorf("-M from a range list parsed as a flag")
	}
	for _, n := range []string{"--fields", "--only-delimited", "-z"} {
		if !got[n] {
			t.Errorf("missing %s", n)
		}
	}
}
