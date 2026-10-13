// Package mcp exposes behave as an MCP server so a coding assistant can
// explore a command-line tool, grow a spec, and check a port against it.
package mcp

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/emirmrkaljevic/behave/internal/diff"
	"github.com/emirmrkaljevic/behave/internal/engine"
	"github.com/emirmrkaljevic/behave/internal/probe"
	"github.com/emirmrkaljevic/behave/internal/runner"
	"github.com/emirmrkaljevic/behave/internal/spec"
)

// NewServer builds the behave MCP server.
func NewServer(version string) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "behave", Version: version}, &mcp.ServerOptions{
		Instructions: "behave records how a command-line tool behaves and checks whether another " +
			"implementation behaves the same. Typical flow: behave_exec to explore, behave_probe to " +
			"generate and record a spec, behave_add_case for interesting cases you find, then " +
			"behave_diff with impl set to the new implementation. Fix differences and re-run behave_diff.",
	})
	mcp.AddTool(s, &mcp.Tool{Name: "behave_exec", Description: "Run a command once in an isolated temp " +
		"directory and return its exit code, stdout and stderr. Nothing is saved. Use it to explore a tool."}, execTool)
	mcp.AddTool(s, &mcp.Tool{Name: "behave_probe", Description: "Read a command's --help/man page, generate " +
		"test cases (flags, flag pairs, edge-case stdin, files, error paths) and record the command's behavior " +
		"into a new spec directory."}, probeTool(version))
	mcp.AddTool(s, &mcp.Tool{Name: "behave_add_case", Description: "Add one case to a spec and record the " +
		"original command's behavior for it."}, addTool)
	mcp.AddTool(s, &mcp.Tool{Name: "behave_list_cases", Description: "List the cases in a spec with their " +
		"arguments and recorded exit codes."}, listTool)
	mcp.AddTool(s, &mcp.Tool{Name: "behave_diff", Description: "Replay a spec against an implementation " +
		"(the recorded command if impl is empty) and report which cases and flags behave differently."}, diffTool)
	return s
}

// Run serves over stdio until the client disconnects.
func Run(ctx context.Context, version string) error {
	return NewServer(version).Run(ctx, &mcp.StdioTransport{})
}

type ExecIn struct {
	Command   string            `json:"command" jsonschema:"program name or path"`
	Args      []string          `json:"args,omitempty" jsonschema:"arguments"`
	Stdin     string            `json:"stdin,omitempty" jsonschema:"standard input"`
	Files     map[string]string `json:"files,omitempty" jsonschema:"files to create in the work dir, name to content"`
	TimeoutMS int               `json:"timeout_ms,omitempty" jsonschema:"timeout in milliseconds (default 5000)"`
}

type ExecOut struct {
	Exit      int    `json:"exit"`
	TimedOut  bool   `json:"timed_out,omitempty"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ElapsedMS int64  `json:"elapsed_ms"`
}

func execTool(_ context.Context, _ *mcp.CallToolRequest, in ExecIn) (*mcp.CallToolResult, ExecOut, error) {
	if in.Command == "" {
		return nil, ExecOut{}, errors.New("command is required")
	}
	c := &spec.Case{Args: in.Args, Files: in.Files, TimeoutMS: in.TimeoutMS}
	if in.Stdin != "" {
		c.Stdin = spec.NewBlob([]byte(in.Stdin))
	}
	o, err := runner.Run(in.Command, c, runner.DefaultTimeout)
	if err != nil {
		return nil, ExecOut{}, err
	}
	return nil, ExecOut{Exit: o.Result.Exit, TimedOut: o.Result.TimedOut, Stdout: clip(o.Result.Stdout.Bytes()),
		Stderr: clip(o.Result.Stderr.Bytes()), ElapsedMS: o.Elapsed.Milliseconds()}, nil
}

type ProbeIn struct {
	Command   string   `json:"command" jsonschema:"program to probe"`
	Dir       string   `json:"dir" jsonschema:"spec directory to create"`
	Args      []string `json:"args,omitempty" jsonschema:"positional argument samples, e.g. jq filters"`
	Inputs    string   `json:"inputs,omitempty" jsonschema:"stdin samples: text, json, both or auto (default)"`
	SkipFlags []string `json:"skip_flags,omitempty" jsonschema:"flags never to use"`
	Max       int      `json:"max,omitempty" jsonschema:"maximum number of cases (default 500)"`
}

type ProbeOut struct {
	Dir      string   `json:"dir"`
	Flags    []string `json:"flags"`
	Cases    int      `json:"cases"`
	Flaky    int      `json:"flaky"`
	TimedOut int      `json:"timed_out"`
}

func probeTool(version string) mcp.ToolHandlerFor[ProbeIn, ProbeOut] {
	return func(_ context.Context, _ *mcp.CallToolRequest, in ProbeIn) (*mcp.CallToolResult, ProbeOut, error) {
		if in.Command == "" || in.Dir == "" {
			return nil, ProbeOut{}, errors.New("command and dir are required")
		}
		if _, err := engine.Resolve(in.Command); err != nil {
			return nil, ProbeOut{}, err
		}
		help := probe.HelpText(in.Command, nil)
		flags := probe.ParseHelp(help)
		ins, _, err := probe.InputsFor(in.Inputs, help)
		if err != nil {
			return nil, ProbeOut{}, err
		}
		cases := probe.Generate(flags, probe.Options{Args: in.Args, Inputs: ins, Skip: in.SkipFlags, Max: in.Max, FileInput: true})
		sp, st, err := engine.Create(in.Dir, in.Command, 5000, "behave "+version, cases)
		if err != nil {
			return nil, ProbeOut{}, err
		}
		out := ProbeOut{Dir: sp.Dir, Cases: len(sp.Cases), Flaky: st.Flaky, TimedOut: st.TimedOut, Flags: []string{}}
		for _, f := range flags {
			out.Flags = append(out.Flags, f.Name())
		}
		return nil, out, nil
	}
}

type AddIn struct {
	Dir   string            `json:"dir" jsonschema:"spec directory"`
	Name  string            `json:"name,omitempty" jsonschema:"short description of the case"`
	Args  []string          `json:"args" jsonschema:"arguments"`
	Stdin string            `json:"stdin,omitempty" jsonschema:"standard input"`
	Files map[string]string `json:"files,omitempty" jsonschema:"input files, name to content"`
}

type CaseOut struct {
	File   string   `json:"file"`
	Name   string   `json:"name"`
	Args   []string `json:"args"`
	Exit   int      `json:"exit"`
	Flaky  bool     `json:"flaky,omitempty"`
	Stdout string   `json:"stdout,omitempty"`
	Stderr string   `json:"stderr,omitempty"`
}

func addTool(_ context.Context, _ *mcp.CallToolRequest, in AddIn) (*mcp.CallToolResult, CaseOut, error) {
	s, err := spec.Load(in.Dir)
	if err != nil {
		return nil, CaseOut{}, err
	}
	c := &spec.Case{Name: in.Name, Args: append([]string{}, in.Args...), Files: in.Files, Tags: []string{"manual"}}
	if c.Name == "" {
		c.Name = "manual " + strings.Join(in.Args, " ")
	}
	if in.Stdin != "" {
		c.Stdin = spec.NewBlob([]byte(in.Stdin))
	}
	if err := s.SaveCase(c); err != nil {
		return nil, CaseOut{}, err
	}
	if _, err := engine.Record(s, true); err != nil {
		return nil, CaseOut{}, err
	}
	return nil, CaseOut{File: filepath.Base(c.File), Name: c.Name, Args: c.Args, Exit: c.Expect.Exit, Flaky: c.Flaky,
		Stdout: clip(c.Expect.Stdout.Bytes()), Stderr: clip(c.Expect.Stderr.Bytes())}, nil
}

type ListIn struct {
	Dir string `json:"dir" jsonschema:"spec directory"`
	Tag string `json:"tag,omitempty" jsonschema:"only cases with this tag, e.g. flag:--sort or manual"`
}

type ListOut struct {
	Command string    `json:"command"`
	Total   int       `json:"total"`
	Cases   []CaseOut `json:"cases"`
}

func listTool(_ context.Context, _ *mcp.CallToolRequest, in ListIn) (*mcp.CallToolResult, ListOut, error) {
	s, err := spec.Load(in.Dir)
	if err != nil {
		return nil, ListOut{}, err
	}
	out := ListOut{Command: s.Meta.Command, Cases: []CaseOut{}}
	for _, c := range s.Cases {
		if in.Tag != "" && !has(c.Tags, in.Tag) {
			continue
		}
		co := CaseOut{File: filepath.Base(c.File), Name: c.Name, Args: c.Args, Flaky: c.Flaky}
		if c.Expect != nil {
			co.Exit = c.Expect.Exit
		}
		out.Cases = append(out.Cases, co)
	}
	out.Total = len(out.Cases)
	return nil, out, nil
}

type DiffIn struct {
	Dir          string   `json:"dir" jsonschema:"spec directory"`
	Impl         string   `json:"impl,omitempty" jsonschema:"implementation to test; empty replays the recorded command"`
	IgnoreStderr bool     `json:"ignore_stderr,omitempty" jsonschema:"do not compare stderr"`
	IgnoreExit   bool     `json:"ignore_exit,omitempty" jsonschema:"do not compare exit codes"`
	SkipTags     []string `json:"skip_tags,omitempty" jsonschema:"skip cases with these tags, e.g. meta"`
	MaxFailures  int      `json:"max_failures,omitempty" jsonschema:"failing cases to include in detail (default 10)"`
}

type Failure struct {
	File string   `json:"file"`
	Name string   `json:"name"`
	Args []string `json:"args"`
	Diff string   `json:"diff"`
}

type DiffOut struct {
	Pass     int            `json:"pass"`
	Fail     int            `json:"fail"`
	Skip     int            `json:"skip"`
	Errors   int            `json:"errors"`
	ByField  map[string]int `json:"by_field"`
	Failures []Failure      `json:"failures"`
	Summary  string         `json:"summary"`
}

func diffTool(_ context.Context, _ *mcp.CallToolRequest, in DiffIn) (*mcp.CallToolResult, DiffOut, error) {
	s, err := spec.Load(in.Dir)
	if err != nil {
		return nil, DiffOut{}, err
	}
	if len(in.SkipTags) > 0 {
		var keep []*spec.Case
		for _, c := range s.Cases {
			skip := false
			for _, t := range in.SkipTags {
				skip = skip || has(c.Tags, t)
			}
			if !skip {
				keep = append(keep, c)
			}
		}
		s.Cases = keep
	}
	rep, err := engine.Replay(s, in.Impl, diff.Options{IgnoreStderr: in.IgnoreStderr, IgnoreExit: in.IgnoreExit})
	if err != nil {
		return nil, DiffOut{}, err
	}
	max := in.MaxFailures
	if max <= 0 {
		max = 10
	}
	out := DiffOut{Pass: rep.Pass, Fail: rep.Fail, Skip: rep.Skip, Errors: rep.Error, ByField: rep.ByField, Failures: []Failure{}}
	for _, r := range rep.Results {
		if len(out.Failures) == max {
			break
		}
		if r.Err == nil && len(r.Mismatches) == 0 {
			continue
		}
		var sb strings.Builder
		if r.Err != nil {
			fmt.Fprintf(&sb, "error: %v\n", r.Err)
		}
		for _, m := range r.Mismatches {
			fmt.Fprintf(&sb, "%s:\n%s", m.Field, m.Text)
		}
		out.Failures = append(out.Failures, Failure{File: filepath.Base(r.Case.File), Name: r.Case.Name, Args: r.Case.Args, Diff: sb.String()})
	}
	out.Summary = fmt.Sprintf("%d/%d passed, %d failed, %d skipped in %s", rep.Pass, len(rep.Results), rep.Fail, rep.Skip, rep.Elapsed.Round(time.Millisecond))
	return nil, out, nil
}

func has(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

func clip(b []byte) string {
	const max = 8000
	if len(b) > max {
		return string(b[:max]) + fmt.Sprintf("\n… (%d more bytes)", len(b)-max)
	}
	return string(b)
}
