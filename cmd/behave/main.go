// Command behave records how a CLI behaves and checks that another
// implementation behaves the same.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/emirmrkaljevic/behave/internal/diff"
	"github.com/emirmrkaljevic/behave/internal/engine"
	"github.com/emirmrkaljevic/behave/internal/normalize"
	"github.com/emirmrkaljevic/behave/internal/probe"
	"github.com/emirmrkaljevic/behave/internal/spec"
)

var version = "0.1.0-dev"

const usage = `behave - record how a CLI behaves, prove a rewrite behaves the same

Usage:
  behave probe <cmd> [flags]            discover flags, generate and record cases
  behave add <spec-dir> [flags] -- ARGS  add (and record) one hand-written case
  behave record <spec-dir>              (re)record expectations from the original
  behave run <spec-dir>                 replay against the original (regression check)
  behave diff <spec-dir> --impl <bin>   replay against another implementation
  behave normalizers                    list built-in output normalizers
  behave version

Run 'behave <command> -h' for flags.
`

type multi []string

func (m *multi) String() string     { return strings.Join(*m, ",") }
func (m *multi) Set(v string) error { *m = append(*m, v); return nil }

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	var err error
	code := 0
	args := os.Args[2:]
	switch os.Args[1] {
	case "probe":
		err = cmdProbe(args)
	case "add":
		err = cmdAdd(args)
	case "record":
		err = cmdRecord(args)
	case "run":
		code, err = cmdRun(args, false)
	case "diff":
		code, err = cmdRun(args, true)
	case "normalizers":
		for _, n := range normalize.Names() {
			fmt.Println(n)
		}
	case "version", "--version":
		fmt.Println("behave", version)
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "behave: unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "behave:", err)
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(2)
		}
		os.Exit(1)
	}
	os.Exit(code)
}

// parse lets flags appear after positional arguments.
func parse(fs *flag.FlagSet, args []string) ([]string, error) {
	var pos []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		args = fs.Args()
		if len(args) == 0 {
			return pos, nil
		}
		if args[0] == "--" {
			return append(pos, args...), nil
		}
		pos = append(pos, args[0])
		args = args[1:]
	}
}

func cmdProbe(args []string) error {
	fs := flag.NewFlagSet("probe", flag.ContinueOnError)
	out := fs.String("out", "", "spec directory to write (default: behave-<cmd>)")
	var argSamples, skip, helpArgs multi
	fs.Var(&argSamples, "arg", "positional argument sample, repeatable (e.g. a jq filter)")
	fs.Var(&skip, "skip-flag", "never use this flag, repeatable")
	fs.Var(&helpArgs, "help-arg", "argument that prints help (default --help, then -h)")
	inputs := fs.String("inputs", "auto", "stdin samples: text, json, both, auto")
	max := fs.Int("max", 500, "maximum number of cases")
	timeout := fs.Int("timeout", 5000, "per-case timeout in ms")
	noFile := fs.Bool("no-file-input", false, "do not pass samples as file arguments")
	dry := fs.Bool("dry-run", false, "print discovered flags and case count, write nothing")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return errors.New("usage: behave probe <cmd> [flags]")
	}
	command := pos[0]
	if _, err := engine.Resolve(command); err != nil {
		return err
	}
	if len(helpArgs) == 0 {
		helpArgs = multi{"--help", "-h"}
	}
	var help string
	for _, h := range helpArgs {
		b, _ := exec.Command(command, h).CombinedOutput()
		if len(probe.ParseHelp(string(b))) > 0 {
			help = string(b)
			break
		}
	}
	if man, err := exec.Command("man", "-P", "cat", filepath.Base(command)).Output(); err == nil {
		help += "\n" + string(man)
	}
	flags := probe.ParseHelp(help)
	var ins []probe.Input
	mode := *inputs
	if mode == "auto" {
		mode = "text"
		if strings.Contains(strings.ToLower(help), "json") {
			mode = "json"
		}
	}
	switch mode {
	case "text":
		ins = probe.TextInputs
	case "json":
		ins = probe.JSONInputs
	case "both":
		ins = append(append([]probe.Input{}, probe.TextInputs...), probe.JSONInputs...)
	default:
		return fmt.Errorf("--inputs: unknown value %q", *inputs)
	}
	cases := probe.Generate(flags, probe.Options{Args: argSamples, Inputs: ins, Skip: skip, Max: *max, FileInput: !*noFile})
	fmt.Printf("discovered %d flags, generated %d cases (%s inputs)\n", len(flags), len(cases), mode)
	if *dry {
		for _, f := range flags {
			fmt.Printf("  %-4s %-28s %s\n", f.Short, f.Long, f.Value)
		}
		return nil
	}
	dir := *out
	if dir == "" {
		dir = "behave-" + spec.Slug(filepath.Base(command))
	}
	if _, err := os.Stat(filepath.Join(dir, "behave.yaml")); err == nil {
		return fmt.Errorf("%s already contains a spec; use 'behave record' or another --out", dir)
	}
	s := &spec.Spec{Dir: dir, Meta: spec.Meta{Command: command, TimeoutMS: *timeout, Normalize: normalize.Default,
		RecordedBy: "behave " + version}}
	if err := s.SaveMeta(); err != nil {
		return err
	}
	for _, c := range cases {
		if err := s.SaveCase(c); err != nil {
			return err
		}
	}
	return record(s)
}

func record(s *spec.Spec) error {
	st, err := engine.Record(s, false)
	if err != nil {
		return err
	}
	fmt.Printf("recorded %d cases from %s into %s (%d flaky, %d timed out)\n", st.Recorded, s.Meta.Command, s.Dir, st.Flaky, st.TimedOut)
	return nil
}

func cmdRecord(args []string) error {
	fs := flag.NewFlagSet("record", flag.ContinueOnError)
	missing := fs.Bool("missing", false, "only record cases without expectations")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return errors.New("usage: behave record <spec-dir>")
	}
	s, err := spec.Load(pos[0])
	if err != nil {
		return err
	}
	st, err := engine.Record(s, *missing)
	if err != nil {
		return err
	}
	fmt.Printf("recorded %d cases (%d flaky, %d timed out)\n", st.Recorded, st.Flaky, st.TimedOut)
	return nil
}

func cmdAdd(args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	name := fs.String("name", "", "case name")
	stdin := fs.String("stdin", "", "stdin content")
	var files multi
	fs.Var(&files, "file", "input file NAME=CONTENT, repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rest := fs.Args()
	if len(rest) < 1 {
		return errors.New("usage: behave add <spec-dir> [--name N] [--stdin S] [--file N=C] -- ARGS...")
	}
	s, err := spec.Load(rest[0])
	if err != nil {
		return err
	}
	cargs := rest[1:]
	if len(cargs) > 0 && cargs[0] == "--" {
		cargs = cargs[1:]
	}
	c := &spec.Case{Name: *name, Args: append([]string{}, cargs...), Tags: []string{"manual"}}
	if c.Name == "" {
		c.Name = "manual " + strings.Join(cargs, " ")
	}
	if *stdin != "" {
		c.Stdin = spec.NewBlob([]byte(*stdin))
	}
	for _, f := range files {
		k, v, ok := strings.Cut(f, "=")
		if !ok {
			return fmt.Errorf("--file %q: want NAME=CONTENT", f)
		}
		if c.Files == nil {
			c.Files = map[string]string{}
		}
		c.Files[k] = v
	}
	if err := s.SaveCase(c); err != nil {
		return err
	}
	if _, err := engine.Record(s, true); err != nil {
		return err
	}
	fmt.Printf("added %s (exit %d)\n", c.File, c.Expect.Exit)
	return nil
}

func cmdRun(args []string, needImpl bool) (int, error) {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	impl := fs.String("impl", "", "implementation to test (default: the recorded command)")
	ignStderr := fs.Bool("ignore-stderr", false, "do not compare stderr (error message wording)")
	ignExit := fs.Bool("ignore-exit", false, "do not compare exit codes")
	verbose := fs.Bool("v", false, "show every failing case (default: first 10)")
	asJSON := fs.Bool("json", false, "print a JSON report")
	var tags, skipTags multi
	fs.Var(&tags, "tag", "only run cases with this tag, repeatable")
	fs.Var(&skipTags, "skip-tag", "skip cases with this tag, repeatable (e.g. meta for --help/--version)")
	jobs := fs.Int("jobs", engine.Jobs, "parallel jobs")
	pos, err := parse(fs, args)
	if err != nil {
		return 2, err
	}
	if len(pos) != 1 || (needImpl && *impl == "") {
		return 2, errors.New("usage: behave diff <spec-dir> --impl <bin>  |  behave run <spec-dir>")
	}
	engine.Jobs = *jobs
	s, err := spec.Load(pos[0])
	if err != nil {
		return 2, err
	}
	if len(skipTags) > 0 {
		var keep []*spec.Case
		for _, c := range s.Cases {
			hit := false
			for _, t := range c.Tags {
				hit = hit || contains(skipTags, t)
			}
			if !hit {
				keep = append(keep, c)
			}
		}
		s.Cases = keep
	}
	if len(tags) > 0 {
		var keep []*spec.Case
		for _, c := range s.Cases {
			for _, t := range c.Tags {
				if contains(tags, t) {
					keep = append(keep, c)
					break
				}
			}
		}
		s.Cases = keep
	}
	rep, err := engine.Replay(s, *impl, diff.Options{IgnoreStderr: *ignStderr, IgnoreExit: *ignExit})
	if err != nil {
		return 2, err
	}
	if *asJSON {
		printJSON(rep)
	} else {
		printReport(s, rep, *verbose)
	}
	if rep.Fail+rep.Error > 0 {
		return 1, nil
	}
	return 0, nil
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if x == y {
			return true
		}
	}
	return false
}

func color(code, s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	if fi, err := os.Stdout.Stat(); err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func printReport(s *spec.Spec, rep *engine.Report, verbose bool) {
	shown := 0
	for _, r := range rep.Results {
		if r.Err == nil && len(r.Mismatches) == 0 {
			continue
		}
		if !verbose && shown == 10 {
			fmt.Printf("… %d more failing cases (use -v)\n\n", rep.Fail+rep.Error-shown)
			break
		}
		shown++
		fmt.Printf("%s %s  %s\n", color("31", "✗"), r.Case.Name, color("2", filepath.Base(r.Case.File)))
		fmt.Printf("  %s %s\n", color("2", "$ "+filepath.Base(rep.Command)), quoteArgs(r.Case.Args))
		if r.Err != nil {
			fmt.Printf("  error: %v\n", r.Err)
		}
		for _, m := range r.Mismatches {
			fmt.Printf("  %s\n", color("33", m.Field+":"))
			for _, l := range strings.Split(strings.TrimRight(m.Text, "\n"), "\n") {
				switch {
				case strings.HasPrefix(l, "-"):
					l = color("31", l)
				case strings.HasPrefix(l, "+"):
					l = color("32", l)
				}
				fmt.Printf("    %s\n", l)
			}
		}
		fmt.Println()
	}
	total := len(rep.Results)
	var fields []string
	for f, n := range rep.ByField {
		fields = append(fields, fmt.Sprintf("%s %d", f, n))
	}
	sort.Strings(fields)
	fmt.Printf("%s vs %s: %s, %s, %d skipped", s.Meta.Command, rep.Command,
		color("32", fmt.Sprintf("%d/%d passed", rep.Pass, total)), color("31", fmt.Sprintf("%d failed", rep.Fail)), rep.Skip)
	if rep.Error > 0 {
		fmt.Printf(", %d errors", rep.Error)
	}
	if len(fields) > 0 {
		fmt.Printf("  (differs in: %s)", strings.Join(fields, ", "))
	}
	fmt.Printf("  [%.1fs]\n", rep.Elapsed.Seconds())
}

func quoteArgs(args []string) string {
	var out []string
	for _, a := range args {
		if a == "" || strings.ContainsAny(a, " \t\n'\"$\\|&;<>()*?[]{}!#~") {
			a = "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
		}
		out = append(out, a)
	}
	return strings.Join(out, " ")
}

func printJSON(rep *engine.Report) {
	type jcase struct {
		Name       string          `json:"name"`
		File       string          `json:"file"`
		Status     string          `json:"status"`
		Mismatches []diff.Mismatch `json:"mismatches,omitempty"`
		Error      string          `json:"error,omitempty"`
	}
	out := struct {
		Command string         `json:"command"`
		Pass    int            `json:"pass"`
		Fail    int            `json:"fail"`
		Skip    int            `json:"skip"`
		Error   int            `json:"error"`
		ByField map[string]int `json:"by_field"`
		Cases   []jcase        `json:"cases"`
	}{Command: rep.Command, Pass: rep.Pass, Fail: rep.Fail, Skip: rep.Skip, Error: rep.Error, ByField: rep.ByField}
	for _, r := range rep.Results {
		j := jcase{Name: r.Case.Name, File: r.Case.File, Status: "pass", Mismatches: r.Mismatches}
		switch {
		case r.Skipped != "":
			j.Status = "skip"
		case r.Err != nil:
			j.Status, j.Error = "error", r.Err.Error()
		case len(r.Mismatches) > 0:
			j.Status = "fail"
		}
		out.Cases = append(out.Cases, j)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(out)
}
