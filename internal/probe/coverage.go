package probe

import (
	"sort"
	"strings"
)

// FlagUse counts how many cases exercise one documented flag.
type FlagUse struct {
	Flag  Flag
	Cases int
}

// Coverage maps documented flags to the cases that use them. Unknown lists
// option-looking arguments that matched no documented flag (often error-path
// cases, or flags the help text does not mention).
type Coverage struct {
	Flags   []FlagUse
	Unknown []string
}

// Covered returns how many documented flags appear in at least one case.
func (c Coverage) Covered() int {
	n := 0
	for _, f := range c.Flags {
		if f.Cases > 0 {
			n++
		}
	}
	return n
}

// MeasureCoverage reports which flags are used by the argument lists.
// It understands --long, --long=value, unambiguous --long prefixes,
// -s, -svalue, clustered short flags (-abc) and stops at "--".
func MeasureCoverage(flags []Flag, argLists [][]string) Coverage {
	cov := Coverage{Flags: make([]FlagUse, len(flags))}
	short := map[string]int{}
	for i, f := range flags {
		cov.Flags[i].Flag = f
		if f.Short != "" {
			short[f.Short] = i
		}
	}
	unknown := map[string]bool{}
	for _, args := range argLists {
		hit := map[int]bool{}
		for j := 0; j < len(args); j++ {
			a := args[j]
			if a == "--" {
				break
			}
			if strings.HasPrefix(a, "--") {
				name, _, hasVal := strings.Cut(a, "=")
				i := matchLong(flags, name)
				if i < 0 {
					unknown[name] = true
					continue
				}
				hit[i] = true
				if f := flags[i]; f.Value != "" && !f.Optional && !hasVal {
					j++ // value is the next argument
				}
				continue
			}
			if len(a) < 2 || a[0] != '-' || (a[1] >= '0' && a[1] <= '9') {
				continue // operand, "-" (stdin) or a negative number
			}
			for k := 1; k < len(a); k++ {
				i, ok := short["-"+a[k:k+1]]
				if !ok {
					unknown["-"+a[k:k+1]] = true
					break
				}
				hit[i] = true
				if f := flags[i]; f.Value != "" {
					if k == len(a)-1 && !f.Optional {
						j++
					}
					break // rest of the cluster is the value
				}
			}
		}
		for i := range hit {
			cov.Flags[i].Cases++
		}
	}
	for u := range unknown {
		cov.Unknown = append(cov.Unknown, u)
	}
	sort.Strings(cov.Unknown)
	return cov
}

// matchLong finds a long flag by exact name or unique prefix (GNU getopt style).
func matchLong(flags []Flag, name string) int {
	found := -1
	for i, f := range flags {
		if f.Long == "" {
			continue
		}
		if f.Long == name {
			return i
		}
		if strings.HasPrefix(f.Long, name) && len(name) > 3 {
			if found >= 0 {
				return -1 // ambiguous
			}
			found = i
		}
	}
	return found
}
