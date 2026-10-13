// Package diff compares an expected result with an actual one and renders
// a compact, line-based explanation of the differences.
package diff

import (
	"fmt"
	"strings"

	"github.com/emirmrkaljevic/behave/internal/spec"
)

// Options selects which parts of a result are compared.
type Options struct {
	IgnoreStderr bool
	IgnoreExit   bool
}

// Mismatch describes one differing field.
type Mismatch struct {
	Field string // exit, stdout, stderr, timeout
	Text  string // human readable diff
}

// Compare returns nil when the results match.
func Compare(want, got *spec.Result, o Options) []Mismatch {
	var ms []Mismatch
	if want.TimedOut != got.TimedOut {
		ms = append(ms, Mismatch{"timeout", fmt.Sprintf("timed_out: want %v, got %v", want.TimedOut, got.TimedOut)})
	}
	if !o.IgnoreExit && want.Exit != got.Exit {
		ms = append(ms, Mismatch{"exit", fmt.Sprintf("exit: want %d, got %d", want.Exit, got.Exit)})
	}
	if w, g := string(want.Stdout.Bytes()), string(got.Stdout.Bytes()); w != g {
		ms = append(ms, Mismatch{"stdout", Lines(w, g)})
	}
	if !o.IgnoreStderr {
		if w, g := string(want.Stderr.Bytes()), string(got.Stderr.Bytes()); w != g {
			ms = append(ms, Mismatch{"stderr", Lines(w, g)})
		}
	}
	return ms
}

// MaxLines bounds how many diff lines are rendered per field.
var MaxLines = 20

// Lines renders a minimal line diff (LCS) with -/+ prefixes and a little context.
func Lines(want, got string) string {
	a, b := split(want), split(got)
	if len(a)*len(b) > 4_000_000 { // too big for LCS; show heads
		return fmt.Sprintf("- (%d lines)\n+ (%d lines)\n", len(a), len(b))
	}
	n, m := len(a), len(b)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	type op struct {
		k byte
		s string
	}
	var ops []op
	i, j := 0, 0
	for i < n || j < m {
		switch {
		case i < n && j < m && a[i] == b[j]:
			ops = append(ops, op{' ', a[i]})
			i, j = i+1, j+1
		case j < m && (i == n || lcs[i][j+1] > lcs[i+1][j]):
			ops = append(ops, op{'+', b[j]})
			j++
		default:
			ops = append(ops, op{'-', a[i]})
			i++
		}
	}
	var sb strings.Builder
	shown := 0
	for k, o := range ops {
		if o.k == ' ' && !near(ops, k, func(x op) bool { return x.k != ' ' }) {
			continue
		}
		if shown == MaxLines {
			sb.WriteString("  …\n")
			break
		}
		fmt.Fprintf(&sb, "%c %s\n", o.k, o.s)
		shown++
	}
	return sb.String()
}

func near[T any](xs []T, k int, f func(T) bool) bool {
	for d := -1; d <= 1; d++ {
		if k+d >= 0 && k+d < len(xs) && f(xs[k+d]) {
			return true
		}
	}
	return false
}

func split(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.SplitAfter(s, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for i, l := range lines {
		if strings.HasSuffix(l, "\n") {
			lines[i] = strings.TrimSuffix(l, "\n")
		} else {
			lines[i] = l + "⏎̸" // marks a missing trailing newline
		}
	}
	return lines
}
