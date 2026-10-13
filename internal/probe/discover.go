package probe

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// HelpText runs the first help argument that yields parseable flags and
// appends the man page when one exists.
func HelpText(command string, helpArgs []string) string {
	if len(helpArgs) == 0 {
		helpArgs = []string{"--help", "-h"}
	}
	var help string
	for _, h := range helpArgs {
		b, _ := exec.Command(command, h).CombinedOutput()
		if len(ParseHelp(string(b))) > 0 {
			help = string(b)
			break
		}
	}
	if man, err := exec.Command("man", "-P", "cat", filepath.Base(command)).Output(); err == nil {
		help += "\n" + string(man)
	}
	return help
}

// InputsFor returns stdin samples for mode text, json, both or auto
// (json when the help text mentions JSON) and the resolved mode.
func InputsFor(mode, help string) ([]Input, string, error) {
	if mode == "" || mode == "auto" {
		mode = "text"
		if strings.Contains(strings.ToLower(help), "json") {
			mode = "json"
		}
	}
	switch mode {
	case "text":
		return TextInputs, mode, nil
	case "json":
		return JSONInputs, mode, nil
	case "both":
		return append(append([]Input{}, TextInputs...), JSONInputs...), mode, nil
	}
	return nil, "", fmt.Errorf("unknown inputs mode %q (want text, json, both, auto)", mode)
}
