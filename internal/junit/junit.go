// Package junit writes a JUnit XML report from an engine.Report.
package junit

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/emirmrkaljevic/behave/internal/diff"
	"github.com/emirmrkaljevic/behave/internal/engine"
)

type testSuites struct {
	XMLName  xml.Name    `xml:"testsuites"`
	Name     string      `xml:"name,attr"`
	Tests    int         `xml:"tests,attr"`
	Failures int         `xml:"failures,attr"`
	Errors   int         `xml:"errors,attr"`
	Skipped  int         `xml:"skipped,attr"`
	Time     string      `xml:"time,attr"`
	Suites   []testSuite `xml:"testsuite"`
}

type testSuite struct {
	Name     string     `xml:"name,attr"`
	Tests    int        `xml:"tests,attr"`
	Failures int        `xml:"failures,attr"`
	Errors   int        `xml:"errors,attr"`
	Skipped  int        `xml:"skipped,attr"`
	Time     string     `xml:"time,attr"`
	Cases    []testCase `xml:"testcase"`
}

type testCase struct {
	Name      string       `xml:"name,attr"`
	Classname string       `xml:"classname,attr"`
	Time      string       `xml:"time,attr"`
	Failure   *failureElem `xml:"failure,omitempty"`
	Error     *errorElem   `xml:"error,omitempty"`
	Skipped   *skippedElem `xml:"skipped,omitempty"`
}

type failureElem struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Body    string `xml:",chardata"`
}

type errorElem struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Body    string `xml:",chardata"`
}

type skippedElem struct {
	Message string `xml:"message,attr,omitempty"`
}

// Write encodes rep as JUnit XML to w.
func Write(w io.Writer, rep *engine.Report) error {
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}
	suite := testSuite{
		Name:     rep.Command,
		Tests:    len(rep.Results),
		Failures: rep.Fail,
		Errors:   rep.Error,
		Skipped:  rep.Skip,
		Time:     formatSecs(rep.Elapsed),
	}
	for _, r := range rep.Results {
		tc := testCase{
			Name:      r.Case.Name,
			Classname: r.Case.File,
			Time:      formatSecs(r.Elapsed),
		}
		switch {
		case r.Skipped != "":
			tc.Skipped = &skippedElem{Message: r.Skipped}
		case r.Err != nil:
			tc.Error = &errorElem{
				Message: r.Err.Error(),
				Type:    "error",
				Body:    r.Err.Error(),
			}
		case len(r.Mismatches) > 0:
			tc.Failure = &failureElem{
				Message: mismatchSummary(r.Mismatches),
				Type:    "mismatch",
				Body:    mismatchBody(r.Mismatches),
			}
		}
		suite.Cases = append(suite.Cases, tc)
	}
	root := testSuites{
		Name:     rep.Command,
		Tests:    suite.Tests,
		Failures: suite.Failures,
		Errors:   suite.Errors,
		Skipped:  suite.Skipped,
		Time:     suite.Time,
		Suites:   []testSuite{suite},
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(root); err != nil {
		return err
	}
	_, err := io.WriteString(w, "\n")
	return err
}

func formatSecs(d time.Duration) string {
	return fmt.Sprintf("%.3f", d.Seconds())
}

func mismatchSummary(ms []diff.Mismatch) string {
	fields := make([]string, 0, len(ms))
	seen := map[string]bool{}
	for _, m := range ms {
		if !seen[m.Field] {
			seen[m.Field] = true
			fields = append(fields, m.Field)
		}
	}
	return "differs in: " + strings.Join(fields, ", ")
}

func mismatchBody(ms []diff.Mismatch) string {
	var b strings.Builder
	for _, m := range ms {
		fmt.Fprintf(&b, "%s:\n%s\n", m.Field, m.Text)
	}
	return b.String()
}
