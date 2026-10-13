// Package engine records expectations and replays specs against an implementation.
package engine

import (
	"fmt"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/emirmrkaljevic/behave/internal/diff"
	"github.com/emirmrkaljevic/behave/internal/normalize"
	"github.com/emirmrkaljevic/behave/internal/runner"
	"github.com/emirmrkaljevic/behave/internal/spec"
)

// Jobs is the parallelism used for running cases.
var Jobs = runtime.NumCPU()

// Exec runs a case and returns the normalized result.
func Exec(s *spec.Spec, n *normalize.Normalizer, command string, c *spec.Case) (*spec.Result, time.Duration, error) {
	o, err := runner.Run(command, c, time.Duration(s.Meta.TimeoutMS)*time.Millisecond)
	if err != nil {
		return nil, 0, err
	}
	n.Result(&o.Result, normalize.Context{WorkDir: o.WorkDir, Command: command})
	return &o.Result, o.Elapsed, nil
}

// Resolve checks that a command can be executed.
func Resolve(command string) (string, error) {
	p, err := exec.LookPath(command)
	if err != nil {
		return "", fmt.Errorf("cannot find %q: %w", command, err)
	}
	return p, nil
}

func each(cases []*spec.Case, f func(int, *spec.Case)) {
	var wg sync.WaitGroup
	ch := make(chan int)
	jobs := max(Jobs, 1)
	for w := 0; w < jobs; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range ch {
				f(i, cases[i])
			}
		}()
	}
	for i := range cases {
		ch <- i
	}
	close(ch)
	wg.Wait()
}

// RecordRuns is how many times each case runs while recording; any
// difference between runs marks the case flaky (excluded from diffs).
var RecordRuns = 3

// RecordStats summarizes a record pass.
type RecordStats struct{ Recorded, Flaky, TimedOut int }

// Record runs every case RecordRuns times against the spec's command and stores the
// normalized result; cases whose two runs differ are marked flaky.
func Record(s *spec.Spec, onlyMissing bool) (RecordStats, error) {
	var st RecordStats
	n, err := normalize.New(s.Meta.Normalize, s.Meta.Replace)
	if err != nil {
		return st, err
	}
	if _, err := Resolve(s.Meta.Command); err != nil {
		return st, err
	}
	var mu sync.Mutex
	var firstErr error
	each(s.Cases, func(_ int, c *spec.Case) {
		if onlyMissing && c.Expect != nil {
			return
		}
		r1, _, err := Exec(s, n, s.Meta.Command, c)
		flaky := false
		for k := 1; k < RecordRuns && err == nil && !flaky; k++ {
			var r2 *spec.Result
			if r2, _, err = Exec(s, n, s.Meta.Command, c); err == nil {
				flaky = diff.Compare(r1, r2, diff.Options{}) != nil
			}
		}
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return
		}
		c.Expect = r1
		c.Flaky = flaky
		if c.Flaky {
			st.Flaky++
		}
		if r1.TimedOut {
			st.TimedOut++
		}
		st.Recorded++
	})
	if firstErr != nil {
		return st, firstErr
	}
	for _, c := range s.Cases {
		if err := s.SaveCase(c); err != nil {
			return st, err
		}
	}
	return st, nil
}

// CaseResult is the outcome of replaying one case.
type CaseResult struct {
	Case       *spec.Case
	Got        *spec.Result
	Mismatches []diff.Mismatch
	Skipped    string
	Err        error
	Elapsed    time.Duration
}

// Report is the outcome of a replay.
type Report struct {
	Command                 string
	Results                 []CaseResult
	Pass, Fail, Skip, Error int
	ByField                 map[string]int
	Elapsed                 time.Duration
}

// Replay runs all cases against command (the spec's own command if empty).
func Replay(s *spec.Spec, command string, o diff.Options) (*Report, error) {
	if command == "" {
		command = s.Meta.Command
	}
	path, err := Resolve(command)
	if err != nil {
		return nil, err
	}
	_ = path
	n, err := normalize.New(s.Meta.Normalize, s.Meta.Replace)
	if err != nil {
		return nil, err
	}
	rep := &Report{Command: command, Results: make([]CaseResult, len(s.Cases)), ByField: map[string]int{}}
	start := time.Now()
	each(s.Cases, func(i int, c *spec.Case) {
		cr := CaseResult{Case: c}
		switch {
		case c.Expect == nil:
			cr.Skipped = "not recorded"
		case c.Flaky:
			cr.Skipped = "flaky"
		default:
			cr.Got, cr.Elapsed, cr.Err = Exec(s, n, command, c)
			if cr.Err == nil {
				cr.Mismatches = diff.Compare(c.Expect, cr.Got, o)
			}
		}
		rep.Results[i] = cr
	})
	rep.Elapsed = time.Since(start)
	for _, r := range rep.Results {
		switch {
		case r.Skipped != "":
			rep.Skip++
		case r.Err != nil:
			rep.Error++
		case len(r.Mismatches) > 0:
			rep.Fail++
			for _, m := range r.Mismatches {
				rep.ByField[m.Field]++
			}
		default:
			rep.Pass++
		}
	}
	return rep, nil
}
