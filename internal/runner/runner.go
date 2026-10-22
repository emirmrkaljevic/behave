// Package runner executes one case in an isolated temporary work directory.
package runner

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/emirmrkaljevic/behave/internal/spec"
)

// DefaultTimeout is used when neither the case nor the spec sets one.
const DefaultTimeout = 5 * time.Second

// MaxOutput caps captured stdout/stderr per stream.
const MaxOutput = 1 << 20

// Output is a raw run result plus the work dir used (for normalization).
type Output struct {
	Result  spec.Result
	WorkDir string
	Elapsed time.Duration
}

// Run executes command with the case's args/stdin/env/files.
func Run(command string, c *spec.Case, defTimeout time.Duration) (*Output, error) {
	// The case runs in its own work dir, so a relative path like ./mytool
	// has to be pinned to the caller's directory first.
	if strings.ContainsRune(command, filepath.Separator) && !filepath.IsAbs(command) {
		if abs, err := filepath.Abs(command); err == nil {
			command = abs
		}
	}
	// The work dir sits inside a private parent so "ls -a" sees a ".." that
	// does not change with whatever else lives in the system temp dir.
	root, err := os.MkdirTemp("", "behave-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(root)
	wd := filepath.Join(root, "work")
	if err := os.Mkdir(wd, 0o755); err != nil {
		return nil, err
	}
	if err := writeFiles(wd, c.Files); err != nil {
		return nil, err
	}
	for _, d := range []string{wd, root} {
		if err := os.Chtimes(d, FileTime, FileTime); err != nil {
			return nil, err
		}
	}
	timeout := defTimeout
	if c.TimeoutMS > 0 {
		timeout = time.Duration(c.TimeoutMS) * time.Millisecond
	}
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, c.Args...)
	cmd.Dir = wd
	cmd.Env = baseEnv(wd)
	for k, v := range c.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	cmd.Stdin = bytes.NewReader(c.Stdin.Bytes())
	var so, se capBuf
	cmd.Stdout, cmd.Stderr = &so, &se
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = time.Second

	start := time.Now()
	err = cmd.Run()
	out := &Output{WorkDir: wd, Elapsed: time.Since(start)}
	out.Result.Stdout = spec.NewBlob(so.Bytes())
	out.Result.Stderr = spec.NewBlob(se.Bytes())
	var ee *exec.ExitError
	switch {
	case ctx.Err() == context.DeadlineExceeded:
		out.Result.TimedOut = true
		out.Result.Exit = -1
	case err == nil:
	case errors.As(err, &ee):
		out.Result.Exit = ee.ExitCode()
		if ws, ok := ee.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
			out.Result.Exit = 128 + int(ws.Signal())
		}
	default:
		return nil, err // could not start: missing binary etc.
	}
	return out, nil
}

// baseEnv is a small, deterministic environment.
// FileTime is the mtime of the first input file (in name order); each next
// file is one hour newer and directories get FileTime. Fixed times keep
// ls -l, ls -t, du --time and find -newer output identical between runs
// (including the work dir itself and its parent).
var FileTime = time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)

// writeFiles creates files in reverse name order and touches each directory
// as soon as its last entry is written. ctime and birth time cannot be set,
// so this makes "newest first" equal name order: ls -c and ls --time=birth
// then print the same listing whether or not the files share a clock tick.
func writeFiles(wd string, files map[string]string) error {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	parents := func(name string) []string {
		var ds []string
		for d := filepath.Dir(filepath.Join(wd, filepath.FromSlash(name))); d != wd && strings.HasPrefix(d, wd+string(filepath.Separator)); d = filepath.Dir(d) {
			ds = append(ds, d)
		}
		return ds
	}
	pending := map[string]bool{}
	for i := len(names) - 1; i >= 0; i-- {
		name := names[i]
		p := filepath.Join(wd, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, []byte(files[name]), 0o644); err != nil {
			return err
		}
		t := FileTime.Add(time.Duration(i) * time.Hour)
		if err := os.Chtimes(p, t, t); err != nil {
			return err
		}
		for _, d := range parents(name) {
			pending[d] = true
		}
		next := map[string]bool{}
		if i > 0 {
			for _, d := range parents(names[i-1]) {
				next[d] = true
			}
		}
		for d := range pending {
			if !next[d] {
				if err := os.Chtimes(d, FileTime, FileTime); err != nil {
					return err
				}
				delete(pending, d)
			}
		}
	}
	return nil
}

func baseEnv(wd string) []string {
	env := []string{"HOME=" + wd, "TMPDIR=" + wd, "LANG=C.UTF-8", "LC_ALL=C.UTF-8", "TZ=UTC",
		"NO_COLOR=1", "TERM=dumb", "COLUMNS=80"}
	if p := os.Getenv("PATH"); p != "" {
		env = append(env, "PATH="+p)
	}
	return env
}

type capBuf struct{ bytes.Buffer }

func (c *capBuf) Write(p []byte) (int, error) {
	if room := MaxOutput - c.Len(); room > 0 {
		if len(p) > room {
			c.Buffer.Write(p[:room])
		} else {
			c.Buffer.Write(p)
		}
	}
	return len(p), nil
}
