package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/emirmrkaljevic/behave/internal/normalize"
	"github.com/emirmrkaljevic/behave/internal/spec"
)

// Create writes a new spec directory with the given cases and records them.
func Create(dir, command string, timeoutMS int, recordedBy string, cases []*spec.Case) (*spec.Spec, RecordStats, error) {
	if _, err := os.Stat(filepath.Join(dir, "behave.yaml")); err == nil {
		return nil, RecordStats{}, fmt.Errorf("%s already contains a spec; use record or another directory", dir)
	}
	s := &spec.Spec{Dir: dir, Meta: spec.Meta{Command: command, TimeoutMS: timeoutMS,
		Normalize: normalize.Default, RecordedBy: recordedBy}}
	if err := s.SaveMeta(); err != nil {
		return nil, RecordStats{}, err
	}
	for _, c := range cases {
		if err := s.SaveCase(c); err != nil {
			return nil, RecordStats{}, err
		}
	}
	st, err := Record(s, false)
	return s, st, err
}
