// Package spec defines the on-disk format of a behave spec: a directory with
// a behave.yaml (metadata) and a cases/ directory holding one YAML file per case.
package spec

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// Meta is stored in <dir>/behave.yaml.
type Meta struct {
	Command    string   `yaml:"command"`              // program recorded from (name or path)
	TimeoutMS  int      `yaml:"timeout_ms,omitempty"` // default per-case timeout
	Normalize  []string `yaml:"normalize,omitempty"`  // built-in normalizers to apply
	Replace    []Rule   `yaml:"replace,omitempty"`    // extra regex replacements
	RecordedBy string   `yaml:"recorded_by,omitempty"`
}

// Rule is a user-defined regex replacement applied to stdout/stderr.
type Rule struct {
	Pattern string `yaml:"pattern"`
	With    string `yaml:"with"`
}

// Case is a single black-box observation of the command.
type Case struct {
	Name      string            `yaml:"name"`
	Args      []string          `yaml:"args"`
	Stdin     *Blob             `yaml:"stdin,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
	Files     map[string]string `yaml:"files,omitempty"` // input files created in the work dir
	TimeoutMS int               `yaml:"timeout_ms,omitempty"`
	Tags      []string          `yaml:"tags,omitempty"`
	Flaky     bool              `yaml:"flaky,omitempty"` // non-deterministic; skipped by diff
	Expect    *Result           `yaml:"expect,omitempty"`

	File string `yaml:"-"` // path on disk
}

// Result is the observed behavior of one run.
type Result struct {
	Exit     int   `yaml:"exit"`
	TimedOut bool  `yaml:"timed_out,omitempty"`
	Stdout   *Blob `yaml:"stdout"`
	Stderr   *Blob `yaml:"stderr"`
}

// Blob is text stored readably in YAML, or base64 when it is not valid UTF-8.
type Blob struct{ Data []byte }

func NewBlob(b []byte) *Blob { return &Blob{Data: b} }

func (b *Blob) Bytes() []byte {
	if b == nil {
		return nil
	}
	return b.Data
}

func (b *Blob) MarshalYAML() (any, error) {
	if b == nil {
		return "", nil
	}
	if utf8.Valid(b.Data) && !strings.ContainsRune(string(b.Data), 0) {
		n := &yaml.Node{Kind: yaml.ScalarNode, Value: string(b.Data)}
		if strings.Contains(n.Value, "\n") && !strings.HasSuffix(n.Value, " ") {
			n.Style = yaml.LiteralStyle
		} else {
			n.Style = yaml.DoubleQuotedStyle
		}
		return n, nil
	}
	return map[string]string{"base64": base64.StdEncoding.EncodeToString(b.Data)}, nil
}

func (b *Blob) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.MappingNode {
		var m map[string]string
		if err := n.Decode(&m); err != nil {
			return err
		}
		d, err := base64.StdEncoding.DecodeString(m["base64"])
		b.Data = d
		return err
	}
	b.Data = []byte(n.Value)
	return nil
}

// Spec is a loaded spec directory.
type Spec struct {
	Dir   string
	Meta  Meta
	Cases []*Case
}

func Load(dir string) (*Spec, error) {
	s := &Spec{Dir: dir}
	raw, err := os.ReadFile(filepath.Join(dir, "behave.yaml"))
	if err != nil {
		return nil, fmt.Errorf("not a behave spec (%w)", err)
	}
	if err := yaml.Unmarshal(raw, &s.Meta); err != nil {
		return nil, fmt.Errorf("behave.yaml: %w", err)
	}
	files, _ := filepath.Glob(filepath.Join(dir, "cases", "*.yaml"))
	sort.Strings(files)
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		c := &Case{}
		if err := yaml.Unmarshal(raw, c); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		c.File = f
		s.Cases = append(s.Cases, c)
	}
	return s, nil
}

func (s *Spec) SaveMeta() error {
	if err := os.MkdirAll(filepath.Join(s.Dir, "cases"), 0o755); err != nil {
		return err
	}
	return writeYAML(filepath.Join(s.Dir, "behave.yaml"), &s.Meta)
}

// SaveCase writes c to its file, assigning a file name if it has none.
func (s *Spec) SaveCase(c *Case) error {
	if c.File == "" {
		c.File = filepath.Join(s.Dir, "cases", fmt.Sprintf("%04d-%s.yaml", len(s.Cases)+1, Slug(c.Name)))
		s.Cases = append(s.Cases, c)
	}
	return writeYAML(c.File, c)
}

func writeYAML(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// Slug makes a short file-name-safe string.
func Slug(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
		if b.Len() >= 40 {
			break
		}
	}
	return strings.Trim(b.String(), "-")
}
