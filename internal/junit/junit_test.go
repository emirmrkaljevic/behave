package junit

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/emirmrkaljevic/behave/internal/diff"
	"github.com/emirmrkaljevic/behave/internal/engine"
	"github.com/emirmrkaljevic/behave/internal/spec"
)

func TestWrite(t *testing.T) {
	rep := &engine.Report{
		Command: "cut",
		Pass:    1,
		Fail:    1,
		Skip:    1,
		Error:   1,
		Elapsed: 1500 * time.Millisecond,
		Results: []engine.CaseResult{
			{Case: &spec.Case{Name: "ok", File: "0001.yaml"}, Elapsed: 10 * time.Millisecond},
			{
				Case:       &spec.Case{Name: "bad", File: "0002.yaml"},
				Mismatches: []diff.Mismatch{{Field: "exit", Text: "exit: want 0, got 1"}},
				Elapsed:    20 * time.Millisecond,
			},
			{Case: &spec.Case{Name: "flaky", File: "0003.yaml"}, Skipped: "flaky"},
			{Case: &spec.Case{Name: "boom", File: "0004.yaml"}, Err: errBoom{}, Elapsed: 5 * time.Millisecond},
		},
	}
	var buf bytes.Buffer
	if err := Write(&buf, rep); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, xml.Header) {
		t.Fatalf("missing xml header: %q", out[:min(40, len(out))])
	}
	for _, want := range []string{
		`<testsuites`,
		`name="cut"`,
		`tests="4"`,
		`failures="1"`,
		`errors="1"`,
		`skipped="1"`,
		`<testcase name="ok"`,
		`<failure message="differs in: exit"`,
		`<skipped message="flaky"`,
		`<error message="boom"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

type errBoom struct{}

func (errBoom) Error() string { return "boom" }
