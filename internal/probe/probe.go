// Package probe discovers a command's flags from its help text and generates
// black-box test cases: flags alone and in pairs, varied stdin, file inputs,
// and error paths (unknown flags, missing files, bad values).
package probe

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/emirmrkaljevic/behave/internal/spec"
)

// Flag is an option found in help output.
type Flag struct {
	Short, Long string   // "-n", "--number" (either may be empty)
	Value       string   // value placeholder; empty for boolean flags
	Optional    bool     // value is optional (--flag[=VALUE]); must be attached with '='
	Values      []string // values the help text enumerates for this flag, in order
}

// Name is the preferred spelling (long if available).
func (f Flag) Name() string {
	if f.Long != "" {
		return f.Long
	}
	return f.Short
}

var (
	// "  -n, --number=NUM   text" | "  --raw-output   text" | "  -f file" | "  -c, --compact-output"
	optLine = regexp.MustCompile(`^\s{1,12}(-[A-Za-z0-9?](?:,\s*|\s*[/|]\s*)?)?(--[A-Za-z0-9][\w-]*)?(\[?[ =][<\[]?[A-Za-z][\w.|-]*[>\]]?\]?)?`)
)

// valueTail matches what may follow a lowercase value placeholder: more
// placeholders, then a column gap or end of line ("--arg name value   Set ...").
var valueTail = regexp.MustCompile(`^(\s[<\[]?[a-z][\w-]*[>\]]?)*(\s{2,}|$)`)

// termLine is an indented definition-list entry that is not an option, such
// as cut's range syntax ("  N-M   from N'th to M'th ...").
var termLine = regexp.MustCompile(`^\s{1,12}[A-Z0-9][\w'-]*\s{2,}\S`)

// shortDigit is a negative number, not a flag ("-1").
var shortDigit = regexp.MustCompile(`^-\d$`)

// ParseHelp extracts flags from help text. It is heuristic but conservative.
// Values a flag documents as a closed list ("WHEN can be: auto, always,
// never", clap's "[possible values: ...]", argparse's "{a,b,c}") are kept in
// Flag.Values in the order the help text lists them.
func ParseHelp(help string) []Flag {
	seen := map[string]int{}
	var out []Flag
	descs := map[int][]string{} // flag index -> description lines ("" = paragraph break)
	choices := map[int][]string{}
	cur, curIndent, gap := -1, 0, false
	var prose [][]string // paragraphs of lines outside option descriptions
	var para []string
	endPara := func() {
		if len(para) > 0 {
			prose = append(prose, para)
			para = nil
		}
	}
	// continues reports whether an indented, non-option line belongs to the
	// current option's description.
	continues := func(indent int) bool { return cur >= 0 && indent > curIndent }
	inTerms := false // inside a list of non-option terms
	for _, line := range strings.Split(help, "\n") {
		line = strings.ReplaceAll(line, "\t", "    ")
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if strings.TrimSpace(line) == "" {
			gap = cur >= 0
			endPara()
			inTerms = false
			continue
		}
		desc := func() {
			if gap {
				descs[cur] = append(descs[cur], "")
				gap = false
			}
			descs[cur] = append(descs[cur], strings.TrimSpace(line))
		}
		other := func() {
			if continues(indent) {
				desc()
				return
			}
			cur, gap = -1, false
			para = append(para, line)
		}
		if termLine.MatchString(line) {
			inTerms = true
			other()
			continue
		}
		norm, ch := normalizeOptLine(line)
		m := optLine.FindStringSubmatch(norm)
		if m == nil || (m[1] == "" && m[2] == "") {
			if !strings.HasPrefix(line, " ") {
				inTerms = false
			}
			other()
			continue
		}
		// "-M" right after "N-M" in a range list is syntax, not an option.
		if inTerms && m[2] == "" {
			other()
			continue
		}
		inTerms = false
		f := Flag{Short: strings.Trim(m[1], " ,/|"), Long: m[2]}
		val := strings.TrimSpace(strings.Trim(m[3], "[]= <>"))
		val = strings.Trim(val, "[]<> =")
		// a value placeholder is UPPER, <angled>, or after '='
		raw := m[3]
		rest := norm[len(m[0]):]
		sep := valueTail.MatchString(rest)
		if val != "" && (strings.ContainsAny(raw, "=<") || val == strings.ToUpper(val) || sep) {
			f.Value = val
		}
		// Short flag followed directly by description text: "-n  number lines" -> no value
		if f.Value != "" && strings.HasPrefix(raw, "[=") && f.Long != "" {
			f.Optional = true
		}
		if f.Name() == "--" {
			other()
			continue
		}
		if i, dup := seen[f.Name()]; dup {
			// a flag mentioned again inside a description, or the man page
			// repeating it: keep reading its text for values
			if continues(indent) {
				desc()
				continue
			}
			// GNU du lists "--time" and "--time=WORD": the value is optional
			if out[i].Value == "" && f.Value != "" && f.Value == strings.ToUpper(f.Value) && f.Long != "" && strings.HasPrefix(raw, "=") {
				out[i].Value, out[i].Optional = f.Value, true
			}
			cur, curIndent, gap = i, indent, false
			descs[i] = append(descs[i], "")
			descs[i] = append(descs[i], strings.TrimSpace(rest))
			continue
		}
		// avoid parsing prose like "  -- end of options" or negative numbers
		if f.Short != "" && shortDigit.MatchString(f.Short) && f.Long == "" {
			other()
			continue
		}
		seen[f.Name()] = len(out)
		cur, curIndent, gap = len(out), indent, false
		descs[cur] = []string{strings.TrimSpace(rest)}
		choices[cur] = append(ch, pipeChoices(f.Value)...)
		out = append(out, f)
	}
	endPara()

	for i := range out {
		addValues(&out[i], choices[i])
		addValues(&out[i], linesValues(descs[i]))
		addValues(&out[i], descValues(out[i].Value, joinDesc(descs[i])))
	}
	byPlaceholder := map[string][]string{}
	for _, p := range prose {
		for k, v := range sectionValues(p) {
			byPlaceholder[k] = append(byPlaceholder[k], v...)
		}
		for k, v := range proseValues(joinDesc(p)) {
			byPlaceholder[k] = append(byPlaceholder[k], v...)
		}
	}
	for i := range out {
		addValues(&out[i], byPlaceholder[out[i].Value])
	}
	return out
}

// Options control generation.
type Options struct {
	Args      []string // positional argument samples (e.g. jq filters); "" = none
	Inputs    []Input  // stdin/file samples
	Skip      []string // flags never to use
	Max       int      // cap on number of cases
	FileInput bool     // also pass samples as a file argument
}

// Input is a named stdin sample.
type Input struct{ Name, Data string }

// TextInputs are generic text samples covering common edge cases.
var TextInputs = []Input{
	{"empty", ""},
	{"one line", "hello\n"},
	{"lines", "banana\napple\ncherry\napple\n\n10\n9\n-1\n"},
	{"no trailing newline", "alpha\nbeta"},
	{"tabs and spaces", "a\tb  c\n  leading\ttab\ntrailing  \n"},
	{"unicode", "čćžšđ\n日本語\nemoji 🙂\n"},
	{"crlf", "one\r\ntwo\r\n"},
	{"long line", strings.Repeat("x", 5000) + "\n"},
}

// JSONInputs are JSON samples.
var JSONInputs = []Input{
	{"json empty", ""},
	{"json object", `{"a":1,"b":[1,2,3],"c":{"d":"x"},"e":null,"f":true}` + "\n"},
	{"json array", `[3,1,2,{"k":"v"},"s",null,1.5]` + "\n"},
	{"json stream", "1\n\"two\"\n[3]\n{\"four\":4}\n"},
	{"json unicode", `{"name":"Emir Čolić","emoji":"🙂","esc":"a\nb\u0000c"}` + "\n"},
	{"json numbers", `[0,-0,1e1000,-1e-400,9007199254740993,1.0,100000000000000000000]` + "\n"},
	{"json deep", strings.Repeat("[", 200) + strings.Repeat("]", 200) + "\n"},
	{"json invalid", `{"a":1,}` + "\n"},
	{"json truncated", `{"a":[1,2` + "\n"},
}

var badValues = []string{"0", "1", "-1", "abc", ""}

// Generate builds cases (without expectations).
func Generate(flags []Flag, o Options) []*spec.Case {
	if o.Max <= 0 {
		o.Max = 500
	}
	if len(o.Args) == 0 {
		o.Args = []string{""}
	}
	if len(o.Inputs) == 0 {
		o.Inputs = TextInputs
	}
	skip := map[string]bool{}
	for _, s := range o.Skip {
		skip[s] = true
	}
	var usable, boolFlags []Flag
	for _, f := range flags {
		if skip[f.Long] || skip[f.Short] || isMeta(f) {
			continue
		}
		usable = append(usable, f)
		if f.Value == "" || f.Optional {
			boolFlags = append(boolFlags, f)
		}
	}

	g := &gen{max: o.Max, seen: map[string]bool{}}
	pos := func(a string) []string {
		if a == "" {
			return nil
		}
		return []string{a}
	}
	// 1. meta and error paths
	g.add("help", []string{"--help"}, "", nil, "meta")
	g.add("short help", []string{"-h"}, "", nil, "meta")
	g.add("version", []string{"--version"}, "", nil, "meta")
	g.add("no args no input", nil, "", nil, "meta")
	g.add("unknown long flag", []string{"--behave-unknown-flag"}, "", nil, "error")
	g.add("unknown short flag", []string{"-\u00a7"}, "", nil, "error")
	g.add("missing file", append(pos(o.Args[0]), "does-not-exist.txt"), "", nil, "error")
	g.add("directory as file", append(pos(o.Args[0]), "."), "", nil, "error")

	// 2. every positional × every input
	for _, a := range o.Args {
		for _, in := range o.Inputs {
			g.add(join("arg", a, "stdin", in.Name), pos(a), in.Data, nil, "basic")
		}
	}
	// 3. every input as a file argument (first positional)
	if o.FileInput {
		for _, in := range o.Inputs {
			g.add(join("arg", o.Args[0], "file", in.Name), append(pos(o.Args[0]), "input.txt"), "", map[string]string{"input.txt": in.Data}, "file")
		}
		g.add("two files", append(pos(o.Args[0]), "a.txt", "b.txt"), "", map[string]string{"a.txt": o.Inputs[1%len(o.Inputs)].Data, "b.txt": o.Inputs[2%len(o.Inputs)].Data}, "file")
	}
	// 4. each flag alone, against a few inputs and positionals
	richInputs := pick(o.Inputs, 3)
	for _, f := range usable {
		for ai, a := range o.Args {
			if ai >= 3 {
				break
			}
			if f.Value == "" {
				for _, in := range richInputs {
					args := append(flagArgs(f, ""), pos(a)...)
					g.add(join("flag", f.Name(), "arg", a, "stdin", in.Name), args, in.Data, nil, "flag", "flag:"+f.Name())
				}
			} else {
				in := richInputs[0]
				// every value the help text lists (on a multi-line sample, so
				// e.g. sort --sort=month has something to order), then the
				// invalid ones
				enumIn := richInputs[min(1, len(richInputs)-1)]
				for _, v := range f.Values {
					args := append(flagArgs(f, v), pos(a)...)
					g.add(join("flag", f.Name(), "value", v, "arg", a), args, enumIn.Data, nil, "flag-enum", "flag:"+f.Name())
				}
				for _, v := range badValues {
					args := append(flagArgs(f, v), pos(a)...)
					g.add(join("flag", f.Name(), "value", v, "arg", a), args, in.Data, nil, "flag-value", "flag:"+f.Name())
				}
				if f.Optional {
					g.add(join("flag", f.Name(), "arg", a, "stdin", in.Name), append(flagArgs(f, ""), pos(a)...), in.Data, nil, "flag", "flag:"+f.Name())
				} else {
					g.add(join("flag", f.Name(), "missing value"), flagArgs(f, "")[:1], in.Data, nil, "error", "flag:"+f.Name())
				}
			}
		}
	}
	// 5. pairs of boolean flags
	in := richInputs[len(richInputs)-1]
	for i := 0; i < len(boolFlags); i++ {
		for j := i + 1; j < len(boolFlags); j++ {
			args := append(append(flagArgs(boolFlags[i], ""), flagArgs(boolFlags[j], "")...), pos(o.Args[0])...)
			g.add(join("pair", boolFlags[i].Name(), boolFlags[j].Name()), args, in.Data, nil, "pair", "flag:"+boolFlags[i].Name(), "flag:"+boolFlags[j].Name())
		}
	}
	// 6. short flag spelling equivalence and "--" handling
	for _, f := range boolFlags {
		if f.Short != "" && f.Long != "" {
			g.add(join("short", f.Short), append([]string{f.Short}, pos(o.Args[0])...), in.Data, nil, "alias", "flag:"+f.Name())
		}
	}
	g.add("double dash", append([]string{"--"}, pos(o.Args[0])...), in.Data, nil, "basic")
	return g.cases
}

type gen struct {
	max   int
	seen  map[string]bool
	cases []*spec.Case
}

func (g *gen) add(name string, args []string, stdin string, files map[string]string, tags ...string) {
	if len(g.cases) >= g.max {
		return
	}
	key := fmt.Sprintf("%q|%q|%v", args, stdin, sortedKeys(files))
	if g.seen[key] {
		return
	}
	g.seen[key] = true
	c := &spec.Case{Name: name, Args: args, Files: files, Tags: tags}
	if args == nil {
		c.Args = []string{}
	}
	if stdin != "" {
		c.Stdin = spec.NewBlob([]byte(stdin))
	}
	g.cases = append(g.cases, c)
}

func flagArgs(f Flag, v string) []string {
	name := f.Name()
	if f.Value == "" || (f.Optional && v == "") {
		return []string{name}
	}
	if f.Optional {
		return []string{name + "=" + v}
	}
	return []string{name, v}
}

func isMeta(f Flag) bool {
	switch f.Long {
	case "--help", "--version", "--usage":
		return true
	}
	return f.Short == "-h" && f.Long == "" || f.Short == "-?"
}

func pick(in []Input, n int) []Input {
	// prefer non-empty samples
	var out []Input
	for _, i := range in {
		if i.Data != "" && len(out) < n {
			out = append(out, i)
		}
	}
	if len(out) == 0 {
		out = in[:1]
	}
	return out
}

func join(parts ...string) string {
	var keep []string
	for i := 0; i+1 < len(parts); i += 2 {
		if parts[i+1] != "" {
			keep = append(keep, parts[i]+" "+parts[i+1])
		}
	}
	if len(parts)%2 == 1 {
		keep = append(keep, parts[len(parts)-1])
	}
	return strings.Join(keep, ", ")
}

func sortedKeys(m map[string]string) []string {
	var ks []string
	for k, v := range m {
		ks = append(ks, k+"="+v)
	}
	sort.Strings(ks)
	return ks
}
