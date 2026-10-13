// Package normalize removes run-specific noise (temp paths, timestamps, colors,
// program name) so outputs from different runs or implementations compare equal.
package normalize

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/emirmrkaljevic/behave/internal/spec"
)

// Context carries per-run values that some normalizers need.
type Context struct {
	WorkDir string // temp dir the case ran in
	Command string // program that produced the output
}

type builtin struct {
	desc string
	fn   func([]byte, Context) []byte
}

var (
	ansiRe = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]|\x1b\][^\x07]*\x07`)
	isoRe  = regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:?\d{2})?`)
	timeRe = regexp.MustCompile(`\b\d{2}:\d{2}:\d{2}(\.\d+)?\b`)
	hexRe  = regexp.MustCompile(`\b0x[0-9a-fA-F]{6,16}\b`)
	durRe  = regexp.MustCompile(`\b\d+(\.\d+)?\s?(ns|µs|us|ms|s)\b`)
	pidRe  = regexp.MustCompile(`\bpid[ =:]?\d+\b`)
)

var builtins = map[string]builtin{
	"ansi":      {"strip ANSI color/escape sequences", func(b []byte, _ Context) []byte { return ansiRe.ReplaceAll(b, nil) }},
	"workdir":   {"replace the case's temp work dir with $WORKDIR", workdir},
	"cmd":       {"replace the program name (e.g. 'jq:' vs 'gojq:') with $CMD", cmdName},
	"timestamp": {"replace ISO dates and HH:MM:SS times with $TIME", func(b []byte, _ Context) []byte { return timeRe.ReplaceAll(isoRe.ReplaceAll(b, []byte("$$TIME")), []byte("$$TIME")) }},
	"hexaddr":   {"replace 0x… addresses with $ADDR", func(b []byte, _ Context) []byte { return hexRe.ReplaceAll(b, []byte("$$ADDR")) }},
	"duration":  {"replace durations like 12ms with $DURATION", func(b []byte, _ Context) []byte { return durRe.ReplaceAll(b, []byte("$$DURATION")) }},
	"pid":       {"replace 'pid 1234' with pid $PID", func(b []byte, _ Context) []byte { return pidRe.ReplaceAll(b, []byte("pid $$PID")) }},
}

// Default is the normalizer set used for new specs.
var Default = []string{"ansi", "workdir", "cmd", "timestamp"}

// Names lists built-in normalizers with descriptions.
func Names() []string {
	var out []string
	for k, v := range builtins {
		out = append(out, fmt.Sprintf("%-10s %s", k, v.desc))
	}
	sort.Strings(out)
	return out
}

// Normalizer applies a configured chain.
type Normalizer struct {
	fns   []func([]byte, Context) []byte
	rules []*regexp.Regexp
	with  []string
}

func New(names []string, rules []spec.Rule) (*Normalizer, error) {
	n := &Normalizer{}
	for _, name := range names {
		b, ok := builtins[name]
		if !ok {
			return nil, fmt.Errorf("unknown normalizer %q", name)
		}
		n.fns = append(n.fns, b.fn)
	}
	for _, r := range rules {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("replace %q: %w", r.Pattern, err)
		}
		n.rules = append(n.rules, re)
		n.with = append(n.with, r.With)
	}
	return n, nil
}

func (n *Normalizer) Apply(b []byte, ctx Context) []byte {
	for _, f := range n.fns {
		b = f(b, ctx)
	}
	for i, re := range n.rules {
		b = re.ReplaceAll(b, []byte(n.with[i]))
	}
	return b
}

// Result normalizes stdout and stderr in place.
func (n *Normalizer) Result(r *spec.Result, ctx Context) {
	r.Stdout = spec.NewBlob(n.Apply(r.Stdout.Bytes(), ctx))
	r.Stderr = spec.NewBlob(n.Apply(r.Stderr.Bytes(), ctx))
}

func workdir(b []byte, ctx Context) []byte {
	if ctx.WorkDir == "" {
		return b
	}
	s := string(b)
	s = strings.ReplaceAll(s, ctx.WorkDir+"/", "")
	s = strings.ReplaceAll(s, ctx.WorkDir, "$WORKDIR")
	return []byte(s)
}

func cmdName(b []byte, ctx Context) []byte {
	if ctx.Command == "" {
		return b
	}
	// Replace full path first, then the base name as a whole word.
	s := strings.ReplaceAll(string(b), ctx.Command, "$CMD")
	base := filepath.Base(ctx.Command)
	re := regexp.MustCompile(`(^|[\s'"(\[/])` + regexp.QuoteMeta(base) + `\b`)
	return re.ReplaceAll([]byte(s), []byte("${1}$$CMD"))
}
