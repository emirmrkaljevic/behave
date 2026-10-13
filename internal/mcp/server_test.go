package mcp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func call(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any, out any) {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if res.IsError {
		t.Fatalf("%s returned error: %+v", name, res.Content[0])
	}
	b, _ := json.Marshal(res.StructuredContent)
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatal(err)
	}
}

func TestServer(t *testing.T) {
	ctx := context.Background()
	st, ct := mcp.NewInMemoryTransports()
	if _, err := NewServer("test").Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "t", Version: "0"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	tools, err := cs.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 5 {
		t.Fatalf("tools: %v %v", tools, err)
	}

	var ex ExecOut
	call(t, cs, "behave_exec", map[string]any{"command": "wc", "args": []string{"-l"}, "stdin": "a\nb\n"}, &ex)
	if ex.Exit != 0 || ex.Stdout != "2\n" {
		t.Fatalf("exec: %+v", ex)
	}

	dir := filepath.Join(t.TempDir(), "spec")
	var pr ProbeOut
	call(t, cs, "behave_probe", map[string]any{"command": "wc", "dir": dir, "max": 40}, &pr)
	if pr.Cases != 40 || len(pr.Flags) == 0 {
		t.Fatalf("probe: %+v", pr)
	}

	var add CaseOut
	call(t, cs, "behave_add_case", map[string]any{"dir": dir, "args": []string{"-w"}, "stdin": "one two three\n"}, &add)
	if add.Exit != 0 || add.Stdout != "3\n" {
		t.Fatalf("add: %+v", add)
	}

	var ls ListOut
	call(t, cs, "behave_list_cases", map[string]any{"dir": dir, "tag": "manual"}, &ls)
	if ls.Total != 1 {
		t.Fatalf("list: %+v", ls)
	}

	var d DiffOut
	call(t, cs, "behave_diff", map[string]any{"dir": dir}, &d)
	if d.Fail != 0 || d.Pass != 41 {
		t.Fatalf("diff self: %+v", d)
	}
	call(t, cs, "behave_diff", map[string]any{"dir": dir, "impl": "cat"}, &d)
	if d.Fail == 0 || len(d.Failures) == 0 || d.Failures[0].Diff == "" {
		t.Fatalf("diff cat: %+v", d)
	}
}
