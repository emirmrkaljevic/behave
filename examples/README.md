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

## `cat/`: 39 recorded cases for GNU `cat`

Plain copies, empty input, a missing final newline, several files and `-` (also `- -`),
numbering with `-n` and `-b` (and `-b` winning over `-n`, numbering that carries on across
files, even when a file has no final newline), `-s` squeezing (also across files), `-E`,
`-T`, `-v` on control, DEL and high bytes, `-A`/`-e`/`-t`, long options, `-u`, NUL bytes,
a file called `-n` after `--`, and the error paths (missing file, directory operand,
unknown option). Recorded with GNU coreutils 9.12, and it also passes on 9.7.

```sh
behave run examples/cat
# cat vs cat: 39/39 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0 everything matches except the two bad-option cases,
where clap prints its own usage message:

```sh
behave diff examples/cat --impl ./uutils/cat
# 37/39 (stderr 2)
```

## `wc/`: 38 recorded cases for GNU `wc`

Default counts on stdin and files, empty input, a last line without a newline, each of
`-l`, `-w`, `-c`, `-m`, `-L` (tabs counted to the next multiple of 8), combined flags and
the fixed column order whatever order you pass them in, long options, several files with a
total line, `-` between files, every `--total` mode, `--files0-from`, form feed and vertical
tab as word separators, NUL bytes, a file name with spaces, and the error paths (missing
file, directory operand, unknown option, bad `--total` value, `--files0-from` misuse).
Recorded with GNU coreutils 9.12, and it also passes on 9.7.

```sh
behave run examples/wc
# wc vs wc: 38/38 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0 every count matches; only the two bad-option cases differ,
where clap prints its own usage message:

```sh
behave diff examples/wc --impl ./uutils/wc
# 36/38 (stderr 2)
```

## `tac/`: 33 recorded cases for GNU `tac`

Reversing stdin and files, empty input, a last line without a newline, blank lines, files
reversed one by one (and a file without a trailing newline running into the next), `-`
between files, `-b`/`--before`, `-s`/`--separator` with one and several characters, an
accepted empty separator, `-r`/`--regex` (including GNU's backward matching, where
`-r -s '[0-9]+'` on `a1b22c333` prints `33c32b2a1`), NUL bytes, UTF-8, a long line, and
the error paths (missing file, unknown option, `-s` without a value, invalid regex).
Recorded with GNU coreutils 9.12, and it also passes on 9.7.

```sh
behave run examples/tac
# tac vs tac: 33/33 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0 every output matches, even the backward regex matching;
the 3 differences are error wording on stderr (clap usage text and the regex parser's
own message):

```sh
behave diff examples/tac --impl ./uutils/tac
# 30/33 (stderr 3)
```

## `numfmt/`: 41 recorded cases for GNU `numfmt`

`--to` and `--from` with `si`, `iec`, `iec-i` and `auto`, the four `--round` modes,
`--suffix`, `--padding` both ways, `--format` width and precision, `--grouping`,
`--from-unit`/`--to-unit`, several arguments, stdin, `--field` (one field and a range),
`--header`, `-d`, `-z`, the three `--invalid` modes, negative and decimal input, and the
error paths (invalid number, invalid or rejected suffix, a value too large to print,
unknown option, bad `--to` value, bad `--format`). Recorded with GNU coreutils 9.12, and it
also passes on 9.7.

```sh
behave run examples/numfmt
# numfmt vs numfmt: 41/41 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0, 32/41 match. Most differences are quoting and clap
wording on stderr, but one is real: `numfmt --from=si 2000Y` prints
`2000000000000000026575110144` and exits 0, where GNU refuses any value it cannot print
exactly (`value too large to be printed: '2e+27'`, exit 2):

```sh
behave diff examples/numfmt --impl ./uutils/numfmt
# 32/41 (stderr 9, stdout 1, exit 1)
```

## `od/`: 36 recorded cases for GNU `od`

Octal, hex, decimal and named-character output (`-b`, `-c`, `-a`, `-x`, `-d`, `-t x1` up
to `x8`, signed and unsigned bytes, `f4` floats), several types at once, the `z` printable
column, the four address radixes, `-j` and `-N` (with a `b` suffix), `-w`, repeated-line
collapsing with and without `-v`, `--endian`, `-S` strings, the traditional `+OFFSET`
form, empty input, and the error paths (unknown option, bad type, unsupported size, bad
radix, skip past end, missing file). Binary input is stored as base64 in the case file.
Recorded with GNU coreutils 9.12, and it also passes on 9.7.

```sh
behave run examples/od
# od vs od: 36/36 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0, 32/36 match. Every dump matches byte for byte; the four
differences are error wording on stderr (`Radix must be one of [o, d, x, n], got: z`
where GNU says `invalid output address radix 'z'; it must be one character from [doxn]`,
and similar for bad `-t` types):

```sh
behave diff examples/od --impl ./uutils/od
# 32/36 (stderr 4)
```

## `basenc/`: 35 recorded cases for GNU `basenc`

Encoding and decoding with every alphabet that exists since coreutils 9.4: `--base64`,
`--base64url`, `--base32`, `--base32hex`, `--base16`, `--base2msbf`, `--base2lsbf` and
`--z85` (including its multiple-of-4 rule), `-w` wrapping (default 76, `-w 0`, `-w 8`),
`-i` to skip garbage, newlines inside encoded input, file and `-` operands, and the error
paths (no encoding given, bad and negative `-w`, invalid input, unknown option, extra
operand, missing file), plus three decodes that hit a bad byte after valid input: GNU
still writes the complete groups it decoded before the error. Recorded with GNU coreutils 9.12, and it also passes on 9.7.
`--base58` (new in 9.8) and lowercase `--base16 -d` (accepted since 9.5) are left out so
the spec also replays on the older GNU in CI.

```sh
behave run examples/basenc
# basenc vs basenc: 35/35 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0, 26/32 match. All encoded and decoded output matches; the
six differences are on stderr: ASCII quotes where GNU uses `‘’` in a UTF-8 locale, an
extra `error: ` before `invalid input`, and clap rejecting `-w -1` and `--bogus` with its
own message. The `error: ` prefix and `-w -1` are already fixed on uutils main.
On uutils main (0.13.0) the three partial-decode cases fail too: `--z85`, `--base16` and
`--base64` print nothing before `invalid input`, while GNU keeps the decoded prefix
(tracked upstream in [#6008](https://github.com/uutils/coreutils/issues/6008) and
[#12923](https://github.com/uutils/coreutils/issues/12923)).

```sh
behave diff examples/basenc --impl ./uutils/basenc
# 26/32 (stderr 6)
```

## `seq/`: 39 recorded cases for GNU `seq`

One, two and three operands, negative and fractional steps, an empty range, decimal
precision taken from the operands (`0.10 0.05 0.25`), exponent and hex float input, `-w`
equal width (also with negatives and fractions), `-s` separators (empty and multi-char),
`-f` formats (`%03g`, `%e`, `%f`, text around the directive, `%%`), and the error paths:
zero increment, non-numeric operands, `nan`, missing and extra operands, a format with no
or two directives, `-w` together with `-f`, and an unknown option. Recorded with GNU
coreutils 9.12, and it also passes on 9.7. Huge start values are left out, since 9.12
changed how they are rounded.

```sh
behave run examples/seq
# seq vs seq: 39/39 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0, 31/39 match. Every number it prints matches; the eight
differences are on stderr. Most are ASCII quotes where GNU uses `‘’` in a UTF-8 locale
and clap's own `extra operand` message, but one is a real parsing difference: uutils lets
`--no-such-flag` through as a number and reports `invalid floating point argument`, where
GNU says `unrecognized option`.

```sh
behave diff examples/seq --impl ./uutils/seq
# 31/39 (stderr 8)
```

## `expr/`: 43 recorded cases for GNU `expr`

Arithmetic with truncating division and a modulo that keeps the dividend's sign,
operator precedence and parentheses, a sum past 64 bits, numeric versus string
comparison (`2 < 10` is true, `2 < 10a` compares as text), `|` and `&` with empty and
zero values, `:` and `match` (anchored at the start, a `\(...\)` group returns the text),
`substr`, `index`, `length`, `+` to quote a keyword, and every exit status: 1 when the
result is empty or zero, 2 for syntax errors, missing operands and a bad regex, and
division by zero. Recorded with GNU coreutils 9.12, and it also passes on 9.7.

```sh
behave run examples/expr
# expr vs expr: 43/43 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0, 40/43 match. Every result and exit status is the same;
the three differences are ASCII quotes in syntax errors where GNU uses `‘’` in a UTF-8
locale.

```sh
behave diff examples/expr --impl ./uutils/expr
# 40/43 (stderr 3)
```

## `printf/`: 38 recorded cases for GNU `printf`

Escapes in the format (`\n`, octal, `\x41`, `\u00e9`, `\c` to stop output), the format
reused for extra arguments and missing arguments read as empty, `%d` with width, `-`, `0`
and `+` flags, `%*d`, hex, octal and character-constant input (`'A` is 65), `%x %X %o`
with and without `#`, `%u` of -1, `%f %e %g` with precision, `%.3s`, `%c`, `%b` and `%q`,
`%%`, and the errors: a non-numeric or half-numeric argument (printed as far as it parsed,
exit 1), a value out of range, a missing format and an invalid conversion. Recorded with
GNU coreutils 9.12, and it also passes on 9.7.

```sh
behave run examples/printf
# printf vs printf: 38/38 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0, 33/38 match. Four are ASCII quotes in error messages where
GNU uses `‘’` in a UTF-8 locale. The fifth is `printf '%.2f' 2.675`: GNU parses the
argument into a binary `long double` (2.67499999...) and prints `2.67`, while uutils keeps
the exact decimal and prints `2.68`. uutils already tracks this as a known limitation (an
ignored test that needs `long double` precision), so it was not filed.

```sh
behave diff examples/printf --impl ./uutils/printf
# 33/38 (stderr 4, stdout 1)
```

## `factor/`: 31 recorded cases for GNU `factor`

Primes, composites, repeated factors, several arguments, 0, 1 and 2, `2^31-1`, the 64-bit
maximum and numbers above it up to 128 bits, a product of two large primes, a leading `+`,
leading zeros and spaces, `-h`/`--exponents`, numbers from stdin split on any whitespace,
and the errors: negative numbers (with and without `--`), words, decimals, hex, an empty
argument, a bad token in the middle (the rest still gets factored, exit 1) and an unknown
option. Recorded with GNU coreutils 9.12, and it also passes on 9.7.

```sh
behave run examples/factor
# factor vs factor: 31/31 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0, 21/31 match. Seven are ASCII quotes in error messages and
two are clap's wording for an unknown option. The last one is a real difference:
`factor ' 30 '` prints `30: 2 3 5` in uutils, while GNU skips only leading spaces and rejects
the trailing one (exit 1). Reported in
[uutils/coreutils#16142](https://github.com/uutils/coreutils/issues/16142), fix in
[#16143](https://github.com/uutils/coreutils/pull/16143).

```sh
behave diff examples/factor --impl ./uutils/factor
# 21/31 (stderr 10, stdout 1, exit 1)
```

## `cksum/`: 35 recorded cases for GNU `cksum`

The default CRC, `bsd`, `sysv`, `crc32b`, `md5`, `sha1`, `sha256`, `sha384`, `sha512`,
`blake2b` (default and `-l 256`) and `sm3`, tagged and `--untagged` output, `--base64`, `-z`,
stdin with and without `-`, several files, and `--check` with good, mismatched, malformed and
missing entries (`--quiet`, `--status`, `--strict`, `--ignore-missing`), plus the errors: a
missing file, a directory, a bad algorithm, a bad `-l`, `-l` with sha256, `--check` with
`-a crc` and an unknown option. Recorded with GNU coreutils 9.12. On 9.7, 33/35 pass: the list of
valid algorithms and the `--length` error changed in between.

```sh
behave run examples/cksum
# cksum vs cksum: 35/35 passed, 0 failed, 0 skipped
```

Against uutils coreutils 0.12.0, 31/35 match. Three are error wording (clap's messages for a
bad algorithm and an unknown option, ASCII quotes in the `-l` error). The fourth is real:
`cksum -` prints `3015617425 6` in uutils, where GNU prints the name, `3015617425 6 -`. It was
already reported in [uutils/coreutils#15984](https://github.com/uutils/coreutils/issues/15984)
with a fix in review ([#16031](https://github.com/uutils/coreutils/pull/16031)).

```sh
behave diff examples/cksum --impl ./uutils/cksum
# 31/35 (stderr 3, stdout 1)
```

## `md5sum/`: 37 recorded cases for GNU `md5sum`

Digests of stdin (with and without `-`), one and two files, `--tag`, `-b`, `-t`, `-z` (also with
`--tag`), and the escaping of a name with a backslash. `--check` with good, mismatched, binary
(`*`), tagged and escaped lines, from stdin, with a missing or malformed entry, `--quiet`,
`--status`, `-w`, `--strict` and `--ignore-missing`, plus an empty or garbage checksum file. The
errors: a missing file, a directory, `--tag` with `--text`, `--quiet` or `--tag` where they make
no sense, and an unknown option. Recorded with GNU coreutils 9.12. On 9.7, 36/37 pass: 9.7
prints an escaped name as `\a\\b: OK` where 9.12 quotes it, `'a\b': OK`.

```sh
behave run examples/md5sum
# md5sum vs md5sum: 37/37 passed, 0 failed, 0 skipped
```

Against uutils coreutils (0.12.0 and main), 32/37 match. Four are error wording (clap's message
for an unknown option, no `Try --help` line after three usage errors). The fifth is real:
`md5sum -c` rejects the line md5sum writes for a name with a backslash or a newline
(`\b1946...  a\\b`), because only the tagged parser skips the leading `\`, and it unescapes
names on lines without it. Reported in
[uutils/coreutils#16166](https://github.com/uutils/coreutils/issues/16166), fix in
[#16167](https://github.com/uutils/coreutils/pull/16167).

```sh
behave diff examples/md5sum --impl ./uutils/md5sum
# 32/37 (stderr 5, stdout 1, exit 1)
```

## `tail/`: 33 recorded cases for GNU `tail`

`-n` and `-c` counts from the end and from a `+N` start (including `0`, `+0`, the `b`
multiplier, and more lines than the input has), obsolete `-2`, `+11` and `-3c`, `-z`, a
missing trailing newline, a multibyte character cut by `-c`, headers with several files,
`-q`/`-v`, `-` for stdin, empty files, the warning for `--pid` without `-f`, and error paths
(missing file, directory, bad counts, bad `-s` and `--pid` values, unknown options). The cases
never follow a file, so they finish instantly. Recorded with GNU coreutils 9.12. On 9.7, 32/33
pass: 9.7 prints the stdin header as `==> standard input <==` where 9.12 quotes it.

```sh
behave run examples/tail
# tail vs tail: 33/33 passed, 0 failed, 0 skipped
```

Against uutils coreutils built from main (0.13.0), output and exit codes all match; the seven
failures are error wording only (ASCII instead of curly quotes, clap usage errors, and an extra
`invalid digit found in string` after a bad `--pid`):

```sh
behave diff examples/tail --impl ./coreutils-main/tail
# 26/33, all stderr
behave diff examples/tail --impl ./coreutils-main/tail --ignore-stderr
# 33/33
```

## `base64/`: 34 recorded cases for GNU `base64`

Encoding from stdin, a file and `-`, padding for one, two and three bytes, binary input,
wrapping (default 76, `-w 0`, `-w 1`, `-w 20`, `--wrap=10`, and a size too large to
represent, which GNU treats as no wrapping), decoding wrapped input, missing and misplaced
padding, truncated input, spaces and CRLF line ends, `-i` with garbage, and error paths
(invalid characters, bad and negative `-w` values, missing file, directory, extra operand,
unknown option). Recorded with GNU coreutils 9.12; 9.7 passes all 34.

Unpadded input decodes since GNU coreutils 9.5; older releases stop with `invalid input`. The
cases that depend on it (`decode-truncated` here, `decode-missing-padding` and
`decode-truncated` in `base32/`) are tagged `gnu-9.5`, and CI, which runs on 9.4, skips them
with `--skip-tag gnu-9.5`.

```sh
behave run examples/base64
# base64 vs base64: 34/34 passed, 0 failed, 0 skipped
```

Against uutils coreutils built from main (0.13.0), 26/34 pass. Besides error wording, three
behaviour differences: a too-large `-w` is rejected instead of turning wrapping off (reported
in [#16180](https://github.com/uutils/coreutils/issues/16180), fix in
[#16181](https://github.com/uutils/coreutils/pull/16181)), `\r` before a newline is accepted
when decoding where GNU stops with `invalid input`, and on invalid input GNU writes the bytes
decoded so far while uutils writes nothing:

```sh
behave diff examples/base64 --impl ./coreutils-main/base64
# 26/34
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

## `basename/` and `dirname/`: 35 + 26 recorded cases for GNU `basename` and `dirname`
Path edge cases that are easy to get wrong in a rewrite: trailing and repeated slashes,
`/`, `//` and `///`, the empty string, `.` and `..`, suffix removal (`-s`, a suffix equal
to the whole name is kept, suffix after a trailing slash), `-a`/`--multiple`, `-z`, names
with a leading dash, UTF-8 and a newline inside a name, plus the error paths (missing and
extra operand, missing `-s` argument, unknown options). Recorded with GNU coreutils 9.12,
and both also pass on 9.7, so they run in CI.

```sh
behave run examples/basename
# basename vs basename: 35/35 passed, 0 failed, 0 skipped
behave run examples/dirname
# dirname vs dirname: 26/26 passed, 0 failed, 0 skipped
```

Against uutils coreutils main, every path result matches (basename 31/35, dirname 24/26);
the differences are only error wording on stderr (ASCII quotes and clap's messages for
unknown options and a missing `-s` value).

```sh
behave diff examples/basename --impl ./uutils/basename
behave diff examples/dirname --impl ./uutils/dirname
```


## `fmt/`: 36 recorded cases for GNU `fmt`

Line filling at the default width and with `-w`, `-WIDTH` and `--width`, `-g` goal width,
`-s` split only, `-u` uniform spacing, `-c` crown and `-t` tagged paragraphs, `-p` prefix,
paragraph breaks on blank lines and indentation changes, long words, empty input, a missing
final newline, and the error paths (bad, huge and missing widths, a goal above the width,
a missing file, unknown options). Recorded with GNU coreutils 9.12. Not in CI: GNU 9.7
fills two cases differently (multibyte words and `-w 20 -g 20`).

```sh
behave run examples/fmt
# fmt vs fmt: 36/36 passed, 0 failed, 0 skipped
```

Against uutils coreutils main, 27/36, and 33/36 with `--ignore-stderr`. Two of the three
output differences are the known `--goal` difference ([uutils/coreutils#5162](https://github.com/uutils/coreutils/issues/5162)).
The third was a real bug: an indented paragraph could get a line longer than `-w`
(`    at a narrow width` at `-w 20`), reported in [uutils/coreutils#16190](https://github.com/uutils/coreutils/issues/16190)
with a fix in [#16191](https://github.com/uutils/coreutils/pull/16191).

```sh
behave diff examples/fmt --impl ./uutils/fmt --ignore-stderr
```

## `realpath/`: 38 recorded cases for GNU `realpath`

Files in nested directories, `.` and `..` components, missing files with the default mode,
`-e` and `-m`, `-q`, `-z`, `-s`/`--no-symlinks`, `-L` and `-P`, `--relative-to` and
`--relative-base` (alone and together), root and double-slash paths, a trailing slash on a
file, an empty operand, `--`, and the error paths. Paths inside the case's temp directory are
normalized, so the spec replays anywhere. Recorded with GNU coreutils 9.12 (9.7 passes too).

```sh
behave run examples/realpath
# realpath vs realpath: 38/38 passed, 0 failed, 0 skipped
```


## `pr/`: 40 recorded cases for GNU `pr`

Columns (`-COLS` / `--columns`, `-a` across), `-d` double-space, `-t` / `-T` omit headers,
pinned `-h` / `-D` headers, `-l` length (including length that implies `-t`), `-o` indent,
`-n` / `-N` line numbers, `-m` merge, `-s` / `-S` separators, `-w` / `-W` width, `-f` / `-F`
form feeds, `+FIRST_PAGE` page ranges, `-J` join, `-c` / `-v` / `-e` / `-i`, multiple files,
stdin with `-`, empty input, `--`, and the error paths. Header dates are pinned with `-D` so
the cases stay deterministic. Recorded with GNU coreutils 9.12 (9.7 passes too).

```sh
behave run examples/pr
# pr vs pr: 40/40 passed, 0 failed, 0 skipped
```

Against uutils coreutils main, 25/40, and 27/40 with `--ignore-stderr`. Most
remaining misses are column separator/alignment (see
[uutils/coreutils#12313](https://github.com/uutils/coreutils/issues/12313)) and
unimplemented `-c`/`-v`/`-i`. One real bug was `-o` skipping the page header
([uutils/coreutils#16211](https://github.com/uutils/coreutils/issues/16211),
fix in [#16212](https://github.com/uutils/coreutils/pull/16212)).

```sh
behave diff examples/pr --impl ./uutils/pr --ignore-stderr
```

## `tee/`: 38 recorded cases for GNU `tee`

Stdin copy to stdout, one/multiple output files, overwrite vs `-a`/`--append`,
`-i`/`--ignore-interrupts`, `-p`, `--output-error` modes (`warn`, `warn-nopipe`,
`exit`, `exit-nopipe`) plus invalid values, flag combinations, empty and large
stdin, CRLF and unicode, `--`, and the error paths (directory operand, missing
parent directory, trailing slash on a file, unknown options). Recorded with GNU
coreutils 9.12 (9.7 passes too).

```sh
behave run examples/tee
# tee vs tee: 38/38 passed, 0 failed, 0 skipped
```


## `split/`: 40 recorded cases for GNU `split`

Line/byte/`-C` splits, `-n` chunk modes that write to stdout (`K/N`, `l/K/N`,
`r/K/N`) and to files, numeric/hex/`--additional-suffix`/`-a` suffixes,
custom prefix, `--filter=cat`, record separators (`-t`, NUL), `--verbose`
diagnostics, `-e` elide-empty, file and stdin/`-` inputs, `--`, and the common
error paths (zero/invalid sizes, missing file, bad chunk number, unknown
options). Recorded with GNU coreutils 9.12 (9.7 passes too).

```sh
behave run examples/split
# split vs split: 40/40 passed, 0 failed, 0 skipped
```


## `tsort/`: 33 recorded cases for GNU `tsort`

Topological sort of pairs from stdin, a file operand and `-`: chains, diamonds,
disconnected components, duplicate and self edges, pairs spread over lines or
packed on one line, tabs/CRLF/empty lines, numeric and UTF-8 tokens, a long
token, a 31-node chain, and the cycle paths (two-node, three-node, cycle with
a tail, two separate cycles) where GNU prints `input contains a loop` on stderr,
still emits every node and exits 1. Error paths: odd token count, missing file,
directory operand, extra operand, unknown options. `--help`, `--version` and
`-w` are left out on purpose: `-w` is accepted (and ignored) only since GNU
9.12, so older coreutils in CI would reject it. Recorded with GNU coreutils 9.12
(9.7 passes too).

```sh
behave run examples/tsort
# tsort vs tsort: 33/33 passed, 0 failed, 0 skipped
```

## `csplit/`: 37 recorded cases for GNU `csplit`

Context splitting by line numbers and regexes: `/re/` with `+N`/`-N` offsets,
`%re%` skips, `{N}` and `{*}` repeats, `-s`, `-z`, `-k`, `-f`, `-n` and `-b`
naming options, stdin via `-`, input without a final newline, and the error
paths (no match, line out of range or not increasing, bad repeat count, bad
regex, bad suffix format, missing file or operand). Only stdout (the byte count
of each piece) and stderr are compared, not the files csplit writes.

```sh
behave run examples/csplit
# csplit vs csplit: 37/37 passed, 0 failed, 0 skipped
```

## `xargs/`: 44 recorded cases for GNU `xargs`

Argument building from stdin or `-a`/`--arg-file`: default `echo`, `-n`/`-L`/`-l` batching,
`-d` and `-0` delimiters, `-I`/`-i`/`--replace` (including with the default command, which
prints a blank line per item), shell-like quoting and backslash escapes, `-E`/`-e`/`--eof`
(including `--eof -l` and `-e -l`, which still enable line mode because the optional EOF
value is not taken from the next option), `--`, a custom `printf` command, Unicode and CRLF
input, and a hyphen-leading arg-file name. Error paths: unknown options, bad `-n`/`-L`/`-s`
values, a missing `-I` value, an empty `-d`, a missing arg file, and a missing command.
`--help`, `--version`, `--show-limits`, `-t`, `-p`, `-o`, `-P` greater than 1, and `-s` size
limits that depend on ARG_MAX are left out so the suite stays deterministic and CI-safe.
Recorded with GNU findutils 4.10.0. This suite is what found [uutils/findutils#984](https://github.com/uutils/findutils/issues/984)
(`-I` with the default command) and [#985](https://github.com/uutils/findutils/issues/985) (option values that start with a dash).

```sh
behave run examples/xargs
# xargs vs xargs: 44/44 passed, 0 failed, 0 skipped
```


## `sysctl/`: 31 recorded cases for procps-ng `sysctl`

Read-only lookups of stable keys (`kernel.ostype`, `fs.overflowuid`, `fs.overflowgid`,
`kernel.overflowuid`): default `name = value` output, `-n`/`--values`, `-N`/`--names`
(including `-n -N` and `-nN`), slash keys such as `kernel/ostype`, `--`, and `-q` on a
read. Also the BSD no-op flags `-o` and `-x` (they must not consume the next key),
`-e`/`--ignore` with missing keys, a missing key that still prints later keys, and an
absolute `/proc/sys/...` path that GNU rejects. `--help`, `--version`, `-a`/`-A`/`-X`,
`-w`, `-p`/`--load`, `--pattern`, and `-b` are left out so the suite stays deterministic
and CI-safe (no host-specific values, no writes, no full dumps). Recorded with
procps-ng 4.0.4. This suite is what found [uutils/procps#835](https://github.com/uutils/procps/issues/835)
(`-o`/`-x` swallowing the next operand). CI puts `/usr/sbin` on `PATH` because `sysctl`
often lives there.

```sh
PATH="/usr/sbin:$PATH" behave run examples/sysctl
# sysctl vs sysctl: 31/31 passed, 0 failed, 0 skipped
```

Against uutils procps main (before #837): 24/31. Four of the misses are the `-o`/`-x`
bugs; the other three are error wording on a missing key and GNU rejecting an absolute
`/proc/sys` path.

```sh
PATH="/usr/sbin:$PATH" behave diff examples/sysctl --impl ./uutils/sysctl
```
