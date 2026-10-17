package probe

import (
	"reflect"
	"strings"
	"testing"
)

func values(t *testing.T, help string) map[string][]string {
	t.Helper()
	got := map[string][]string{}
	for _, f := range ParseHelp(help) {
		if len(f.Values) > 0 {
			got[f.Name()] = f.Values
		}
	}
	return got
}

func expect(t *testing.T, got map[string][]string, want map[string][]string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("values:\n got  %v\n want %v", got, want)
	}
}

// GNU coreutils 9.7 ls --help (excerpt)
const gnuLs = `Usage: ls [OPTION]... [FILE]...
      --color[=WHEN]         color the output WHEN; more info below
  -F, --classify[=WHEN]      append indicator (one of */=>@|) to entries WHEN
      --format=WORD          across,horizontal (-x), commas (-m), long (-l),
                             single-column (-1), verbose (-l), vertical (-C)

      --full-time            like -l --time-style=full-iso
      --hide=PATTERN         do not list implied entries matching shell PATTERN
                             (overridden by -a or -A)

      --indicator-style=WORD
                             append indicator with style WORD to entry names:
                             none (default), slash (-p),
                             file-type (--file-type), classify (-F)

  -p, --indicator-style=slash
                             append / indicator to directories
      --quoting-style=WORD   use quoting style WORD for entry names:
                             literal, locale, shell, shell-always,
                             shell-escape, shell-escape-always, c, escape
                             (overrides QUOTING_STYLE environment variable)

      --sort=WORD            change default 'name' sort to WORD:
                               none (-U), size (-S), time (-t),
                               version (-v), extension (-X), name, width

      --time=WORD            select which timestamp used to display or sort;
                               access time (-u): atime, access, use;
                               metadata change time (-c): ctime, status;
                               modified time (default): mtime, modification;
                               birth time: birth, creation;
                             with -l, WORD determines which time to show;
                             with --sort=time, sort by WORD (newest first)

      --time-style=TIME_STYLE
                             time/date format with -l; see TIME_STYLE below
  -T, --tabsize=COLS         assume tab stops at each COLS instead of 8
      --version     output version information and exit

The SIZE argument is an integer and optional unit (example: 10K is 10*1024).
Units are K,M,G,T,P,E,Z,Y,R,Q (powers of 1024) or KB,MB,... (powers of 1000).

The TIME_STYLE argument can be full-iso, long-iso, iso, locale, or +FORMAT.
FORMAT is interpreted like in date(1).  If FORMAT is FORMAT1<newline>FORMAT2,
then FORMAT1 applies to non-recent files and FORMAT2 to recent files.

The WHEN argument defaults to 'always' and can also be 'auto' or 'never'.
`

func TestValuesGNULs(t *testing.T) {
	when := []string{"always", "auto", "never"}
	expect(t, values(t, gnuLs), map[string][]string{
		"--color":           when,
		"--classify":        when,
		"--format":          {"across", "horizontal", "commas", "long", "single-column", "verbose", "vertical"},
		"--indicator-style": {"none", "slash", "file-type", "classify"},
		"--quoting-style":   {"literal", "locale", "shell", "shell-always", "shell-escape", "shell-escape-always", "c", "escape"},
		"--sort":            {"none", "size", "time", "version", "extension", "name", "width"},
		"--time":            {"atime", "access", "use", "ctime", "status", "mtime", "modification", "birth", "creation"},
		"--time-style":      {"full-iso", "long-iso", "iso", "locale"},
	})
}

func TestValuesGNUMisc(t *testing.T) {
	// numfmt, du, sort, date and wc from GNU coreutils 9.7
	help := `      --from=UNIT      auto-scale input numbers to UNITs; default is 'none';
                         see UNIT below
      --invalid=MODE   failure mode for invalid numbers: MODE can be:
                         abort (default), fail, warn, ignore
      --round=METHOD   use METHOD for rounding when scaling; METHOD can be:
                         up, down, from-zero (default), towards-zero, nearest
      --to=UNIT        auto-scale output numbers to UNITs; see UNIT below
      --time            show time of the last modification of any file in the
                          directory, or any of its subdirectories
      --time=WORD       show time as WORD instead of modification time:
                          atime, access, use, ctime or status
      --time-style=STYLE  show times using STYLE, which can be:
                            full-iso, long-iso, iso, or +FORMAT;
                            FORMAT is interpreted like in 'date'
      --sort=WORD             sort according to WORD:
                                general-numeric -g, human-numeric -h, month -M,
                                numeric -n, random -R, version -V
  -I[FMT], --iso-8601[=FMT]  output date/time in ISO 8601 format.
                               FMT='date' for date only (the default),
                               'hours', 'minutes', 'seconds', or 'ns'
                               for date and time to the indicated precision.
                               Example: 2006-08-14T02:34:56-06:00
      --total=WHEN       when to print a line with total counts;
                           WHEN can be: auto, always, only, never

UNIT options:
  none       no auto-scaling is done; suffixes will trigger an error
  auto       accept optional single/two letter suffix:
               1K = 1000, 1k = 1000,
               1Ki = 1024,
  si         accept optional single letter suffix:
  iec        accept optional single letter suffix:
  iec-i      accept optional two-letter suffix:
`
	units := []string{"none", "auto", "si", "iec", "iec-i"}
	expect(t, values(t, help), map[string][]string{
		"--from":       units,
		"--to":         units,
		"--invalid":    {"abort", "fail", "warn", "ignore"},
		"--round":      {"up", "down", "from-zero", "towards-zero", "nearest"},
		"--time":       {"atime", "access", "use", "ctime", "status"},
		"--time-style": {"full-iso", "long-iso", "iso"},
		"--sort":       {"general-numeric", "human-numeric", "month", "numeric", "random", "version"},
		"--iso-8601":   {"date", "hours", "minutes", "seconds", "ns"},
		"--total":      {"auto", "always", "only", "never"},
	})
	for _, f := range ParseHelp(help) {
		if f.Name() == "--time" && (!f.Optional || f.Value != "WORD") {
			t.Errorf("--time listed as both bare and =WORD should take an optional value: %+v", f)
		}
		if f.Name() == "--iso-8601" && f.Short != "-I" {
			t.Errorf("-I[FMT] short form lost: %+v", f)
		}
	}
}

func TestValuesClap(t *testing.T) {
	// uutils coreutils 0.12.0 (clap)
	help := `      --hyperlink[=<WHEN>]
          hyperlink file names WHEN [default: never] [possible values: always, auto, never]
      --quoting-style <quoting-style>
          Set quoting style. [possible values: literal, locale, shell, shell-escape, shell-always,
          shell-escape-always, clocale, c, escape]
      --sort <field>
          Sort by ` + "`<field>`" + `: name, none (-U), time (-t), size (-S), extension (-X) or width
          [possible values: name, none, time, size, version, extension, width]
  -F, --classify[=<when>]
          Append a character to each file name indicating the file type. when may be omitted, or one of:
              none - Do not classify. This is the default.
              auto - Only classify if standard output is a terminal.
          Specifying --classify and no when is equivalent to --classify=always. [possible values: always,
          auto, never]
  -c, --check[=<check>...]
          check for sorted input; do not sort [possible values: silent, quiet, diagnose-first]
      --round <METHOD>           use METHOD for rounding when scaling [default: from-zero] [possible
                                 values: up, down, from-zero, towards-zero, nearest]
      --color <WHEN>
          Coloring

          Possible values:
          - always: Always colorize
          - never:  Never colorize
          - auto
  -w, --width <COLS>
          Assume that the terminal is COLS columns wide.
`
	expect(t, values(t, help), map[string][]string{
		"--hyperlink":     {"always", "auto", "never"},
		"--quoting-style": {"literal", "locale", "shell", "shell-escape", "shell-always", "shell-escape-always", "clocale", "c", "escape"},
		"--sort":          {"name", "none", "time", "size", "version", "extension", "width"},
		"--classify":      {"always", "auto", "never"},
		"--check":         {"silent", "quiet", "diagnose-first"},
		"--round":         {"up", "down", "from-zero", "towards-zero", "nearest"},
		"--color":         {"always", "never", "auto"},
	})
}

func TestValuesArgparseAndPipes(t *testing.T) {
	// Python 3.12 argparse, plus docopt-style <a|b|c>
	help := `usage: tool [-h] [-f {json,yaml,text}] [--level {debug,info}]

options:
  -h, --help            show this help message and exit
  -f, --format {json,yaml,text}
                        output format
  --level {debug,info}
  -o OUT, --output OUT  write to OUT
  --mode <fast|safe>    pick a mode
  --path <FILE|DIR>     a path
`
	expect(t, values(t, help), map[string][]string{
		"--format": {"json", "yaml", "text"},
		"--level":  {"debug", "info"},
		"--mode":   {"fast", "safe"},
	})
	got := map[string]Flag{}
	for _, f := range ParseHelp(help) {
		got[f.Name()] = f
	}
	if f := got["--output"]; f.Short != "-o" || f.Value != "OUT" {
		t.Errorf("-o OUT, --output OUT parsed as %+v", f)
	}
	if f := got["--format"]; f.Short != "-f" || f.Value == "" {
		t.Errorf("-f, --format {...} parsed as %+v", f)
	}
}

func TestValuesNotFromProse(t *testing.T) {
	// ripgrep 14 --help: lists about other things must not become values
	help := `    --colors=COLOR_SPEC
        This flag specifies color settings for use in the output. This flag may
        be provided multiple times. Settings are applied iteratively.
        Pre-existing color labels are limited to one of eight choices: red,
        blue, green, cyan, magenta, yellow, white and black. Styles are limited
        to nobold, bold, nointense, intense, nounderline or underline.

        The format of the flag is {type}:{attribute}:{value}. type should be
        one of path, line, column or match. attribute can be fg, bg or style.

    --type-add=TYPESPEC
        For example, to create a type called src that matches C++, Python and
        Markdown files, one can use:

            --type-add 'src:include:cpp,py,md'

    --engine=ENGINE
        Specify which regular expression engine to use.

        Accepted values are default, pcre2, or auto.

    --color=WHEN
        This flag controls when to use colors. The default setting is auto,
        which means ripgrep will try to guess when to use colors.

        ripgrep will suppress color output by default in some other
        circumstances as well. These include, but are not limited to:

        • When the TERM environment variable is not set or set to dumb.
  -F, --classify[=WHEN]      append indicator (one of */=>@|) to entries WHEN
  -B, --block-size=SIZE      scale sizes by SIZE; e.g., '--block-size=M'
      --field=FIELDS   replace the numbers in these input fields (default=1);
      --grouping       use locale-defined grouping of digits, e.g. 1,000,000
`
	expect(t, values(t, help), map[string][]string{
		"--engine": {"default", "pcre2", "auto"},
	})
}

func TestGenerateValues(t *testing.T) {
	cs := Generate(ParseHelp(gnuLs), Options{Max: 2000})
	var color, sortV []string
	for _, c := range cs {
		a := strings.Join(c.Args, " ")
		if strings.HasPrefix(a, "--color=") {
			color = append(color, a)
		}
		if strings.HasPrefix(a, "--sort ") {
			sortV = append(sortV, a)
		}
	}
	// listed values first, in help order, then the invalid-value probes
	// ("--color=" with an empty value is the bare --color case)
	want := []string{"--color=always", "--color=auto", "--color=never", "--color=0", "--color=1", "--color=-1", "--color=abc"}
	if !reflect.DeepEqual(color, want) {
		t.Errorf("--color cases %q\nwant %q", color, want)
	}
	if len(sortV) != 7+len(badValues) || sortV[0] != "--sort none" || sortV[6] != "--sort width" {
		t.Errorf("--sort cases %q", sortV)
	}
	// deterministic
	again := Generate(ParseHelp(gnuLs), Options{Max: 2000})
	for i := range cs {
		if strings.Join(cs[i].Args, "\x00") != strings.Join(again[i].Args, "\x00") {
			t.Fatalf("case %d differs between runs", i)
		}
	}
}

func TestValuesManPageMerged(t *testing.T) {
	// HelpText appends the man page: same flags again, man(1) layout
	man := `
OPTIONS
       --color[=WHEN]
              color the output WHEN; more info below

       --format=WORD
              across,horizontal (-x), commas (-m), long (-l), single-column (-1), verbose (-l), vertical (-C)

       The WHEN argument defaults to 'always' and can also be 'auto' or 'never'.
`
	got := values(t, gnuLs+man)
	if w := []string{"always", "auto", "never"}; !reflect.DeepEqual(got["--color"], w) {
		t.Errorf("--color %v", got["--color"])
	}
	if len(got["--format"]) != 7 {
		t.Errorf("--format %v", got["--format"])
	}
	only := values(t, man)
	if len(only["--format"]) != 7 || len(only["--color"]) != 3 {
		t.Errorf("man page alone: %v", only)
	}
}
