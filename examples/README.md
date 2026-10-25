# Examples

## `cut/` — 27 recorded cases for GNU `cut`

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


## `paste/` — 18 recorded cases for GNU `paste`

Parallel and serial paste, custom delimiter lists (including empty and cycling),
stdin with files, empty inputs, `-z`, and error paths. Recorded with GNU coreutils 9.7.

```sh
behave run examples/paste
# paste vs paste: 18/18 passed, 0 failed, 0 skipped
```

Against uutils 0.12.0 (stderr wording differs on errors; output and exits match):

```sh
behave diff examples/paste --impl ./coreutils-0.12.0/paste
# 16/18 — both failures are stderr only
behave diff examples/paste --impl ./coreutils-0.12.0/paste --ignore-stderr
# 18/18
```



## `join/` — 26 recorded cases for GNU `join`

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
# 23/26 — two stderr-only usage errors, plus -o fields as separate operands (exit/stdout)
behave diff examples/join --impl ./coreutils-0.12.0/join --ignore-stderr
# 25/26 — remaining miss is GNU accepting `join -o 1.1 1.2 2.2 …` (uutils rejects)
```



## `nl/` — 26 recorded cases for GNU `nl`

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
# 24/26 — both failures are stderr only
behave diff examples/nl --impl ./coreutils-0.12.0/nl --ignore-stderr
# 26/26
```

## `fold/` — 28 recorded cases for GNU `fold`

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
# 24/28 — three stderr-only errors, plus multibyte input (stdout)
behave diff examples/fold --impl ./coreutils-0.12.0/fold --ignore-stderr
# 27/28 — GNU 9.7 wraps UTF-8 by bytes; 9.8+ (and uutils) wrap by column
```


## `jq/` — 33 recorded cases for GNU/`jq`

Filters, builtins, common flags (`-r`/`-c`/`-n`/`-s`/`-e`/`--arg`), stdin/file input,
and error paths. Recorded with jq 1.7.

```sh
behave run examples/jq
# jq vs jq: 33/33 passed, 0 failed, 0 skipped
```

Against gojq 0.12.19 (stderr wording differs on errors; output and exits match):

```sh
behave diff examples/jq --impl gojq
# 29/33 — stderr only
behave diff examples/jq --impl gojq --ignore-stderr
# 33/33
```

Against jaq 3.1.1 (`1/0` → `Infinity` is intentional; see jaq#230):

```sh
behave diff examples/jq --impl jaq --ignore-stderr
# 32/33
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
