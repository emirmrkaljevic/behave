# Examples

## `cut/`: 27 recorded cases for GNU `cut`

A small hand-written spec: fields, ranges, delimiters, `--complement`,
`--only-delimited`, `--output-delimiter`, file arguments, and the usual error paths.
Recorded with GNU coreutils 9.7. Each case is one YAML file in `cut/cases/`, so you can
read exactly what was run and what came back.

```sh
git clone https://github.com/emirmrkaljevic/behave && cd behave
go install ./cmd/behave        # or grab a binary from Releases

# replay against the cut on your PATH (passes on any recent GNU coreutils)
behave run examples/cut
# cut vs cut: 27/27 passed, 0 failed, 0 skipped
```

Now point it at another implementation, for example the Rust rewrite from
[uutils/coreutils](https://github.com/uutils/coreutils) (0.12.0 here):

```sh
behave diff examples/cut --impl ./coreutils-0.12.0/cut
```

```
✗ fields: --output-delimiter value starting with -  0012-fields-output-delimiter-value-starting-w.yaml
  $ cut -d, -f1,2 --output-delimiter -1
  exit:
    exit: want 0, got 1
  stdout:
    - a-1b
    - 1-12
  stderr:
    + error: unexpected argument '-1' found
    ...

✗ error: unknown option  0027-error-unknown-option.yaml
  $ cut --no-such-flag
  stderr:
    - $CMD: unrecognized option '--no-such-flag'
    + error: unexpected argument '--no-such-flag' found
    ...

cut vs ./coreutils-0.12.0/cut: 25/27 passed, 2 failed, 0 skipped  (differs in: exit 1, stderr 2, stdout 1)
```

Add `--ignore-stderr` if you only care about output and exit codes (26/27 here).

On macOS the system `cut` is BSD, so `behave run` will show real differences there;
with Homebrew coreutils use `behave diff examples/cut --impl gcut`.


## `paste/`: 18 recorded cases for GNU `paste`

Parallel and serial paste, custom delimiter lists (including empty and cycling),
stdin with files, empty inputs, `-z`, and error paths. Recorded with GNU coreutils 9.7.

```sh
behave run examples/paste
# paste vs paste: 18/18 passed, 0 failed, 0 skipped
```

Against uutils 0.12.0 (stderr wording differs on errors; output and exits match):

```sh
behave diff examples/paste --impl ./coreutils-0.12.0/paste
# 16/18: both failures are stderr only
behave diff examples/paste --impl ./coreutils-0.12.0/paste --ignore-stderr
# 18/18
```



## `join/`: 26 recorded cases for GNU `join`

Join keys (`-1`/`-2`/`-j`), `-a`/`-v` unpairable lines, `-e` empty fills with `-o`,
`-t` delimiters, `-i`, `--header`, stdin as a file, empty inputs, `-z`, and error paths.
Recorded with GNU coreutils 9.7.

```sh
behave run examples/join
# join vs join: 26/26 passed, 0 failed, 0 skipped
```

Against uutils 0.12.0:

```sh
behave diff examples/join --impl ./coreutils-0.12.0/join
# 23/26: two stderr-only usage errors, plus -o fields as separate operands (exit/stdout)
behave diff examples/join --impl ./coreutils-0.12.0/join --ignore-stderr
# 25/26: remaining miss is GNU accepting `join -o 1.1 1.2 2.2 …` (uutils rejects)
```



## `nl/`: 26 recorded cases for GNU `nl`

Body numbering (`-b a`/`t`/`n`/`pPATTERN`), formats (`-n ln`/`rn`/`rz`), `-w`, `-s`,
`-v`, `-i`, `-l`, section delimiters (`-d`, `-p`), stdin/files, empty input, and errors.
Recorded with GNU coreutils 9.7.

```sh
behave run examples/nl
# nl vs nl: 26/26 passed, 0 failed, 0 skipped
```

Against uutils 0.12.0 (stderr wording differs on usage errors; output and exits match):

```sh
behave diff examples/nl --impl ./coreutils-0.12.0/nl
# 24/26: both failures are stderr only
behave diff examples/nl --impl ./coreutils-0.12.0/nl --ignore-stderr
# 26/26
```

## `fold/`: 28 recorded cases for GNU `fold`

`-w` widths (default 80, obsolete `-5`, width 1), `-s` breaking at blanks, `-b` bytes,
tabs, backspace and carriage return, multibyte input, stdin/files, empty input, and errors.
Recorded with GNU coreutils 9.7.

```sh
behave run examples/fold
# fold vs fold: 28/28 passed, 0 failed, 0 skipped
```

Against uutils 0.12.0:

```sh
behave diff examples/fold --impl ./coreutils-0.12.0/fold
# 24/28: three stderr-only errors, plus multibyte input (stdout)
behave diff examples/fold --impl ./coreutils-0.12.0/fold --ignore-stderr
# 27/28: GNU 9.7 wraps UTF-8 by bytes; 9.8+ (and uutils) wrap by column
```


## `jq/`: 33 recorded cases for GNU/`jq`

Filters, builtins, common flags (`-r`/`-c`/`-n`/`-s`/`-e`/`--arg`), stdin/file input,
and error paths. Recorded with jq 1.7.

```sh
behave run examples/jq
# jq vs jq: 33/33 passed, 0 failed, 0 skipped
```

Against gojq 0.12.19 (stderr wording differs on errors; output and exits match):

```sh
behave diff examples/jq --impl gojq
# 29/33: stderr only
behave diff examples/jq --impl gojq --ignore-stderr
# 33/33
```

Against jaq 3.1.1 (`1/0` → `Infinity` is intentional; see jaq#230):

```sh
behave diff examples/jq --impl jaq --ignore-stderr
# 32/33
```

## `head/`: 24 recorded cases for GNU `head`

`-n` and `-c` counts (including negative "all but the last N", `0`, and the `k` suffix),
obsolete `-5`, `-z`, headers with several files, `-q`/`-v`, `-` for stdin, empty files,
and error paths. Recorded with GNU coreutils 9.12.

```sh
behave run examples/head
# head vs head: 24/24 passed, 0 failed, 0 skipped
```

Against uutils coreutils built from main (0.13.0), output and exit codes match; the four
failures are error wording only (quote style in invalid-count messages, and the clap
usage errors for an unknown option and a missing `-n` value):

```sh
behave diff examples/head --impl ./coreutils-main/head
# 20/24, all stderr
behave diff examples/head --impl ./coreutils-main/head --ignore-stderr
# 24/24
```

## `expand/`: 21 recorded cases for GNU `expand`

Default and `-t` tab stops, obsolete `-4`, tab lists with commas or spaces, the `/N` and
`+N` extensions, `-i`, backspace, several files with `-` for stdin, and error paths.
Recorded with GNU coreutils 9.12. There is no multibyte case on purpose: GNU 9.12 counts
`é` as one column, while 9.4 (still the Ubuntu 24.04 default) counts its two bytes, so
the spec would fail on older systems for a reason unrelated to the tool under test.

```sh
behave run examples/expand
# expand vs expand: 21/21 passed, 0 failed, 0 skipped
```

Against uutils coreutils built from main (0.13.0), one output difference: a bare `-`
after a file name is dropped, so `expand f - f` never reads stdin (already fixed in
[uutils/coreutils#15716](https://github.com/uutils/coreutils/pull/15716), not merged yet).
The other two are error wording (quote style, and the clap message for an unknown option):

```sh
behave diff examples/expand --impl ./coreutils-main/expand
# 18/21 (stdout 1, stderr 2)
```

## `unexpand/`: 26 recorded cases for GNU `unexpand`

The reverse of `expand`: leading blanks only by default, `-a` for all blanks, `-t` (which
implies `-a`) with `--first-only`, obsolete `-4`, tab lists, the `/N` and `+N` extensions,
blanks past the last stop, a single space before a stop (kept), backspace, input without
a trailing newline, several files with `-` for stdin, and error paths. Recorded with GNU
coreutils 9.12, and it also passes on 9.7.

```sh
behave run examples/unexpand
# unexpand vs unexpand: 26/26 passed, 0 failed, 0 skipped
```

Against uutils coreutils built from main (0.13.0) the result matches `expand`: `unexpand
f - f` skips stdin (also covered by
[uutils/coreutils#15716](https://github.com/uutils/coreutils/pull/15716)), plus the same
two error wording differences:

```sh
behave diff examples/unexpand --impl ./coreutils-main/unexpand
# 23/26 (stdout 1, stderr 2)
```

## `tr/`: 29 recorded cases for GNU `tr`

Translate with ranges and character classes, set2 padding versus `-t`, the `[c*]` and
`[c*n]` repeat constructs, octal and `\n` escapes, `-d`, `-s`, `-c` and their
combinations, `[=a=]`, and the error paths (reverse range, missing or extra operand,
classes in set2, `[c*]` in set1, unknown class or option). Recorded with GNU coreutils
9.12, and it also passes on 9.7.

```sh
behave run examples/tr
# tr vs tr: 29/29 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0 every stdout and exit code matches; the 6 failures are
error wording only (quote style, the set2 class message split over two lines in GNU,
tracked in [uutils/coreutils#15948](https://github.com/uutils/coreutils/issues/15948),
and the clap message for an unknown option):

```sh
behave diff examples/tr --impl ./uutils/tr
# 23/29 (stderr 6)
```

## `comm/`: 30 recorded cases for GNU `comm`

The three-column layout and every `-1`/`-2`/`-3` combination, `--output-delimiter`
(including the empty delimiter, which writes NUL), `--total`, `-` for stdin on either
side, duplicate lines, a missing final newline, `-z` records, sort-order checking
(default warning, `--check-order`, `--nocheck-order`) and the error paths (missing,
single or extra operand, missing file, unknown option). Recorded with GNU coreutils
9.12, and it also passes on 9.7.

```sh
behave run examples/comm
# comm vs comm: 30/30 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0 all output, ordering checks and exit codes match; the 4
failures are the clap usage messages for operand and option errors:

```sh
behave diff examples/comm --impl ./uutils/comm
# 26/30 (stderr 4)
```

## `uniq/`: 31 recorded cases for GNU `uniq`

Adjacent-only collapsing, `-c`/`-d`/`-u`/`-D` and their combinations, `--all-repeated`
and `--group` with every method, `-i`, field and character skipping (`-f`, `-s`, `-w`,
alone and combined), `-z` records, file and `-` input, a missing final newline, empty
input, and the error paths (meaningless option combos, bad method, bad number, missing
file, extra operand, unknown option). Recorded with GNU coreutils 9.12, and it also
passes on 9.7.

```sh
behave run examples/uniq
# uniq vs uniq: 31/31 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0 all output and exit codes match. Four failures are clap
usage messages; the fifth was a real wording difference for a bad `-f`/`-s`/`-w` value
(`x: invalid number of fields to skip`), reported upstream as
[uutils/coreutils#15970](https://github.com/uutils/coreutils/issues/15970):

```sh
behave diff examples/uniq --impl ./uutils/uniq
# 26/31 (stderr 5)
```

## `sort/`: 35 recorded cases for GNU `sort`

Plain, reverse and unique sorts, numeric (`-n`, `-g`, `-h`), version and month order,
keys with `-k` (whole fields, character offsets, several keys breaking ties) and `-t`,
`-s`, `-f`, `-b`, `-c`/`-C` checks, `-z`, `-m` merges, `-o`, several files and `-`,
a missing final newline, empty input, and the error paths (bad key, incompatible modes,
multi-character tab, `-c` with two files, missing file, unknown option). Inputs stay in
plain ASCII so the order does not depend on the locale. Recorded with GNU coreutils
9.12, and it also passes on 9.7.

```sh
behave run examples/sort
# sort vs sort: 35/35 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0 all output and exit codes match. The four failures are
stderr only: a clap usage message, straight instead of curly quotes for `-k0`, and
`-t ab` / `-c a b` errors that use different wording or add a usage hint:

```sh
behave diff examples/sort --impl ./uutils/sort
# 31/35 (stderr 4)
```

## Grow it

```sh
behave add examples/cut --name 'fields: unicode' --stdin $'é,ü\n' -- -d, -f2
behave run examples/cut
```

`behave add` runs the original right away and stores what it did. To record a much
bigger spec automatically from `--help`, use `behave probe cut` (about 175 cases).


## `behave/`: behave's own CLI, recorded with behave

behave tests itself the same way it tests other tools: usage and unknown-command errors,
`behave normalizers`, missing spec dirs, and two end-to-end cases that write a tiny spec
for `echo` into the work dir and replay it (once with `run`, once with `diff --impl printf`,
which fails because `printf` drops the trailing newline). CI replays these on every push,
so a change to messages or exit codes shows up as a diff.

```sh
behave run examples/behave
# behave vs behave: 7/7 passed, 0 failed, 0 skipped
```
