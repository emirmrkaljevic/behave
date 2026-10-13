// Package runner executes one case in an isolated temporary work directory.
package runner

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
	wd, err := os.MkdirTemp("", "behave-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(wd)
	for name, content := range c.Files {
		p := filepath.Join(wd, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
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
