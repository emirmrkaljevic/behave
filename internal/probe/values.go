package probe

import (
	"regexp"
	"strings"
)

// maxValues caps how many enumerated values one flag contributes.
const maxValues = 16

var (
	// clap: "[possible values: always, auto, never]"
	possibleRe = regexp.MustCompile(`\[possible values:\s*([^\]]*)\]`)
	defaultRe  = regexp.MustCompile(`\[default:[^\]]*\]`)
	// clap long help: "Possible values:" followed by "- auto: ..." lines
	possibleHdr  = regexp.MustCompile(`^\s*Possible values:\s*$`)
	possibleItem = regexp.MustCompile(`^\s*- ([a-z0-9][a-z0-9_.-]*)(?::\s|$)`)
	// argparse: "--format {json,yaml}"; docopt-style "<auto|always|never>"
	braceChoice = regexp.MustCompile(`\{([A-Za-z0-9_.+-]+(?:,[A-Za-z0-9_.+-]+)+)\}`)
	// "-o OUT, --output OUT" and "-I[FMT], --iso-8601[=FMT]" -> "-o, --output OUT"
	shortValue = regexp.MustCompile(`^(\s{1,12}-[A-Za-z0-9?])(?:[ =](?:[A-Z][A-Z0-9_]*|<[^>]+>)|\[[A-Z][A-Z0-9_]*\])(,\s*--)`)
	// list introducers: "WHEN can be: a, b", "is one of a, b"
	canBe      = regexp.MustCompile(`(?i)\b(?:can(?: also)? be(?: one of)?|may be one of|must be one of|is one of)\b:?`)
	explicitRe = regexp.MustCompile(`(?i)\b(?:accepted|possible|valid|allowed|supported) values(?: for this (?:flag|option))? (?:are|include)(?: one of)?\b:?`)
	// sentence or clause boundary
	segSplit = regexp.MustCompile(`;|\.\s|\.$`)
	parens   = regexp.MustCompile(`\([^)]*\)`)
	orAnd    = regexp.MustCompile(`(?:^|\s+)(?:or|and)\s+`)
	alias    = regexp.MustCompile(`\s+-[A-Za-z0-9]$`)
	quoted   = regexp.MustCompile("['\u2018`]([a-z0-9][a-z0-9_.-]*)['\u2019`]")
	valueTok = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9_.-]*[a-z0-9])?$`)
	digits   = regexp.MustCompile(`^[0-9.]+$`)
	upperTok = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	// a prose sentence about a placeholder: "The WHEN argument can be ..."
	aboutRe = regexp.MustCompile(`^(?:The\s+)?([A-Z][A-Z0-9_]+)\s+(?:argument\s+|value\s+)?(?:defaults|can|may|must|is)\b`)
	// a section listing a placeholder's values: "UNIT options:"
	sectionHdr = regexp.MustCompile(`^([A-Z][A-Z0-9_]+) (?:options|values):\s*$`)
	sectionRow = regexp.MustCompile(`^(\s{1,6})([a-z0-9][a-z0-9_.-]*)\s{2,}\S`)
)

// normalizeOptLine rewrites choice syntax on an option line into a plain
// placeholder the flag regexp understands and returns the listed choices.
func normalizeOptLine(line string) (string, []string) {
	if !strings.HasPrefix(strings.TrimLeft(line, " "), "-") {
		return line, nil
	}
	head, tail := line, ""
	if i := strings.Index(strings.TrimLeft(line, " "), "  "); i >= 0 {
		i += len(line) - len(strings.TrimLeft(line, " "))
		head, tail = line[:i], line[i:]
	}
	var vals []string
	if m := braceChoice.FindStringSubmatch(head); m != nil {
		vals = validList(strings.Split(m[1], ","), 1)
		head = braceChoice.ReplaceAllString(head, "CHOICE")
	}
	head = shortValue.ReplaceAllString(head, "$1$2")
	return head + tail, vals
}

// pipeChoices splits a "<auto|always|never>" placeholder into values.
func pipeChoices(val string) []string {
	if !strings.Contains(val, "|") {
		return nil
	}
	return validList(strings.Split(val, "|"), 2)
}

// descValues finds values enumerated in one flag's description.
func descValues(placeholder, desc string) []string {
	var out []string
	for _, m := range possibleRe.FindAllStringSubmatch(desc, -1) {
		out = append(out, validList(strings.Split(m[1], ","), 1)...)
	}
	desc = possibleRe.ReplaceAllString(desc, " ")
	desc = defaultRe.ReplaceAllString(desc, " ")
	for i, seg := range segments(desc) {
		if i == 0 {
			// "--format=WORD  across,horizontal (-x), commas (-m), ..."
			out = append(out, parseList(seg, 3)...)
		}
		out = append(out, quotedAbout(placeholder, seg)...)
		out = append(out, introduced(placeholder, seg)...)
	}
	return out
}

// proseValues finds "The WHEN argument defaults to 'always' and can also be
// 'auto' or 'never'." sentences and returns values per placeholder.
func proseValues(paragraph string) map[string][]string {
	out := map[string][]string{}
	for _, seg := range sentences(paragraph) {
		m := aboutRe.FindStringSubmatch(seg)
		if m == nil {
			continue
		}
		out[m[1]] = append(out[m[1]], quotedAbout(m[1], seg)...)
		out[m[1]] = append(out[m[1]], introduced(m[1], seg)...)
	}
	return out
}

// introduced returns the lists one clause introduces for placeholder:
// "WHEN can be: a, b", "to WORD: a, b", "access time (-u): atime, use" or
// "Accepted values are a, b, or c". Lists about something else in the same
// description ("labels are limited to one of: red, blue") are ignored.
func introduced(placeholder, seg string) []string {
	var out []string
	if loc := explicitRe.FindStringIndex(seg); loc != nil {
		out = append(out, parseList(seg[loc[1]:], 2)...)
	}
	if loc := canBe.FindStringIndex(seg); loc != nil && containsWord(seg[:loc[0]], placeholder) {
		out = append(out, parseList(seg[loc[1]:], 2)...)
	}
	for i, c := range seg {
		if c != ':' || (i+1 < len(seg) && seg[i+1] != ' ') {
			continue // "src:include:cpp" is syntax, not a list
		}
		head := seg[:i]
		if containsWord(head, placeholder) || len(strings.Fields(parens.ReplaceAllString(head, " "))) <= 3 {
			out = append(out, parseList(seg[i+1:], 2)...)
		}
	}
	return out
}

// quotedAbout returns 'quoted' values from a clause that names the
// placeholder ("FMT='date' ..., 'hours', 'minutes', or 'ns'").
func quotedAbout(placeholder, seg string) []string {
	if placeholder == "" || !containsWord(seg, placeholder) {
		return nil
	}
	var vals []string
	for _, m := range quoted.FindAllStringSubmatch(seg, -1) {
		vals = append(vals, m[1])
	}
	if len(dedupe(vals)) < 2 {
		return nil
	}
	return vals
}

// parseList reads "a, b (default), c -x, or +FORMAT" as a value list. It
// returns nil unless every item is a plain lowercase token (placeholders like
// +FORMAT are skipped) and there are at least min of them.
func parseList(s string, min int) []string {
	s = parens.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "`", "")
	var items []string
	for _, p := range strings.Split(s, ",") {
		for _, q := range orAnd.Split(p, -1) {
			items = append(items, q)
		}
	}
	return validList(items, min)
}

func validList(items []string, min int) []string {
	var out []string
	allDigits := true
	for _, it := range items {
		it = strings.TrimSpace(it)
		it = alias.ReplaceAllString(it, "")
		it = strings.Trim(it, "'\"\u2018\u2019 .")
		if it == "" {
			continue
		}
		if strings.HasPrefix(it, "+") || upperTok.MatchString(it) {
			continue // a placeholder such as +FORMAT
		}
		if !valueTok.MatchString(it) {
			return nil
		}
		if !digits.MatchString(it) {
			allDigits = false
		}
		out = append(out, it)
	}
	out = dedupe(out)
	if len(out) < min || allDigits {
		return nil
	}
	return out
}

func segments(s string) []string {
	return segSplit.Split(s, -1)
}

func sentences(s string) []string {
	return regexp.MustCompile(`\.\s|\.$`).Split(s, -1)
}

func containsWord(s, w string) bool {
	return regexp.MustCompile(`(^|[^\w-])` + regexp.QuoteMeta(w) + `([^\w-]|$)`).MatchString(s)
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func addValues(f *Flag, vals []string) {
	if f.Value == "" || len(vals) == 0 {
		return
	}
	f.Values = dedupe(append(f.Values, vals...))
	if len(f.Values) > maxValues {
		f.Values = f.Values[:maxValues]
	}
}

// joinDesc joins description lines; "" marks a paragraph break, which also
// ends any list.
func joinDesc(lines []string) string {
	var b strings.Builder
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			if b.Len() > 0 {
				b.WriteString(". ")
			}
			continue
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(l)
	}
	return b.String()
}

// linesValues reads clap's long-help block:
//
//	Possible values:
//	- always: ...
//	- never
func linesValues(lines []string) []string {
	var out []string
	in := false
	for _, l := range lines {
		switch {
		case possibleHdr.MatchString(l):
			in = true
		case in:
			m := possibleItem.FindStringSubmatch(l)
			if m == nil {
				in = false
				continue
			}
			out = append(out, m[1])
		}
	}
	return out
}

// sectionValues reads a section such as numfmt's
//
//	UNIT options:
//	  none       no auto-scaling is done
//	  auto       accept optional suffix
//	               (deeper lines continue a row)
func sectionValues(para []string) map[string][]string {
	out := map[string][]string{}
	for i := 0; i < len(para); i++ {
		h := sectionHdr.FindStringSubmatch(strings.TrimRight(para[i], " "))
		if h == nil {
			continue
		}
		rowIndent := -1
		var vals []string
		for _, l := range para[i+1:] {
			m := sectionRow.FindStringSubmatch(l)
			indent := len(l) - len(strings.TrimLeft(l, " "))
			if m != nil && (rowIndent < 0 || indent == rowIndent) {
				rowIndent = indent
				vals = append(vals, m[2])
				continue
			}
			if rowIndent >= 0 && indent > rowIndent {
				continue
			}
			break
		}
		if len(vals) >= 2 {
			out[h[1]] = append(out[h[1]], vals...)
		}
	}
	return out
}
