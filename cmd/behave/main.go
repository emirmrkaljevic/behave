// Command behave records how a CLI behaves and checks that another
// implementation behaves the same.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/emirmrkaljevic/behave/internal/diff"
	"github.com/emirmrkaljevic/behave/internal/engine"
	"github.com/emirmrkaljevic/behave/internal/junit"
	behavemcp "github.com/emirmrkaljevic/behave/internal/mcp"
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
  behave coverage <spec-dir>            which documented flags the cases exercise
  behave mcp                            serve the MCP tools over stdio
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
	case "coverage":
		code, err = cmdCoverage(args)
	case "mcp":
		err = behavemcp.Run(context.Background(), version)
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

// parse lets flags appear after positional arguments. Everything from the
// first "--" in flag position on is returned unparsed, "--" included.
func parse(fs *flag.FlagSet, args []string) ([]string, error) {
	args, rest := splitDashDash(fs, args)
	var pos []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		args = fs.Args()
		if len(args) == 0 {
			return append(pos, rest...), nil
		}
		pos = append(pos, args[0])
		args = args[1:]
	}
}

// splitDashDash splits args at the first "--" that is not the value of a flag.
func splitDashDash(fs *flag.FlagSet, args []string) (before, rest []string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			return args[:i], args[i:]
		}
		if len(a) < 2 || a[0] != '-' || strings.Contains(a, "=") {
			continue
		}
		f := fs.Lookup(strings.TrimLeft(a, "-"))
		if f == nil {
			continue
		}
		if b, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && b.IsBoolFlag() {
			continue
		}
		i++ // skip the flag's value
	}
	return args, nil
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
	help := probe.HelpText(command, helpArgs)
	flags := probe.ParseHelp(help)
	ins, mode, err := probe.InputsFor(*inputs, help)
	if err != nil {
		return err
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
	sp, st, err := engine.Create(dir, command, *timeout, "behave "+version, cases)
	if err != nil {
		return err
	}
	fmt.Printf("recorded %d cases from %s into %s (%d flaky, %d timed out)\n", st.Recorded, command, sp.Dir, st.Flaky, st.TimedOut)
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
	rest, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(rest) < 1 || rest[0] == "--" {
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
	verbose := fs.Bool("v", false, "show every failing case")
	maxFail := fs.Int("max-failures", 10, "failing cases to show in detail")
	asJSON := fs.Bool("json", false, "print a JSON report")
	junitPath := fs.String("junit", "", "write a JUnit XML report to this file")
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
		printReport(s, rep, *verbose, *maxFail)
	}
	if *junitPath != "" {
		f, err := os.Create(*junitPath)
		if err != nil {
			return 2, err
		}
		err = junit.Write(f, rep)
		closeErr := f.Close()
		if err != nil {
			return 2, err
		}
		if closeErr != nil {
			return 2, closeErr
		}
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

func printReport(s *spec.Spec, rep *engine.Report, verbose bool, maxFail int) {
	shown := 0
	for _, r := range rep.Results {
		if r.Err == nil && len(r.Mismatches) == 0 {
			continue
		}
		if !verbose && shown == maxFail {
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
	printFlagTable(rep)
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

// printFlagTable lists the flags whose cases fail most, using "flag:" tags.
func printFlagTable(rep *engine.Report) {
	type st struct {
		name        string
		fail, total int
	}
	m := map[string]*st{}
	for _, r := range rep.Results {
		if r.Skipped != "" || contains(r.Case.Tags, "pair") {
			continue // pairs would blame a flag for its partner's difference
		}
		for _, t := range r.Case.Tags {
			name, ok := strings.CutPrefix(t, "flag:")
			if !ok {
				continue
			}
			if m[name] == nil {
				m[name] = &st{name: name}
			}
			m[name].total++
			if r.Err != nil || len(r.Mismatches) > 0 {
				m[name].fail++
			}
		}
	}
	var rows []*st
	for _, v := range m {
		if v.fail > 0 {
			rows = append(rows, v)
		}
	}
	if len(rows) == 0 {
		return
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].fail*rows[j].total != rows[j].fail*rows[i].total {
			return rows[i].fail*rows[j].total > rows[j].fail*rows[i].total
		}
		return rows[i].name < rows[j].name
	})
	fmt.Println("flags with differences:")
	for i, r := range rows {
		if i == 12 {
			fmt.Printf("  … %d more\n", len(rows)-i)
			break
		}
		fmt.Printf("  %-28s %3d/%-3d cases differ\n", r.name, r.fail, r.total)
	}
	fmt.Printf("flags fully compatible: %d of %d\n\n", len(m)-len(rows), len(m))
}

func cmdCoverage(args []string) (int, error) {
	fs := flag.NewFlagSet("coverage", flag.ContinueOnError)
	var helpArgs multi
	fs.Var(&helpArgs, "help-arg", "argument that prints help (default --help, then -h)")
	min := fs.Float64("min", 0, "exit 1 if fewer than this percent of flags are covered")
	asJSON := fs.Bool("json", false, "print the report as JSON")
	all := fs.Bool("all", false, "list covered flags too, not just the missing ones")
	pos, err := parse(fs, args)
	if err != nil {
		return 0, err
	}
	if len(pos) != 1 {
		return 0, errors.New("usage: behave coverage <spec-dir> [--min PCT] [--json]")
	}
	sp, err := spec.Load(pos[0])
	if err != nil {
		return 0, err
	}
	if _, err := engine.Resolve(sp.Meta.Command); err != nil {
		return 0, err
	}
	flags := probe.ParseHelp(probe.HelpText(sp.Meta.Command, helpArgs))
	if len(flags) == 0 {
		return 0, fmt.Errorf("no flags found in %s help output (try --help-arg)", sp.Meta.Command)
	}
	var lists [][]string
	for _, c := range sp.Cases {
		lists = append(lists, c.Args)
	}
	cov := probe.MeasureCoverage(flags, lists)
	pct := 100 * float64(cov.Covered()) / float64(len(flags))
	if *asJSON {
		type row struct {
			Flag  string `json:"flag"`
			Short string `json:"short,omitempty"`
			Cases int    `json:"cases"`
		}
		out := struct {
			Command      string   `json:"command"`
			Cases        int      `json:"cases"`
			Flags        int      `json:"flags"`
			Covered      int      `json:"covered"`
			Percent      float64  `json:"percent"`
			PerFlag      []row    `json:"per_flag"`
			Undocumented []string `json:"undocumented,omitempty"`
		}{sp.Meta.Command, len(sp.Cases), len(flags), cov.Covered(), pct, nil, cov.Unknown}
		for _, u := range cov.Flags {
			out.PerFlag = append(out.PerFlag, row{u.Flag.Name(), u.Flag.Short, u.Cases})
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Printf("%s: %d cases exercise %d of %d documented flags (%.0f%%)\n", sp.Meta.Command, len(sp.Cases), cov.Covered(), len(flags), pct)
		var missing []string
		for _, u := range cov.Flags {
			if u.Cases == 0 {
				missing = append(missing, u.Flag.Name())
			} else if *all {
				fmt.Printf("  %-28s %d cases\n", u.Flag.Name(), u.Cases)
			}
		}
		if len(missing) > 0 {
			fmt.Printf("not exercised (%d): %s\n", len(missing), strings.Join(missing, " "))
		}
		if len(cov.Unknown) > 0 {
			fmt.Printf("used but not in help: %s\n", strings.Join(cov.Unknown, " "))
		}
	}
	if pct < *min {
		if !*asJSON {
			fmt.Printf("coverage %.0f%% is below --min %.0f%%\n", pct, *min)
		}
		return 1, nil
	}
	return 0, nil
}
