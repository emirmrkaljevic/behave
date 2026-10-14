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

## Grow it

```sh
behave add examples/cut --name 'fields: unicode' --stdin $'é,ü\n' -- -d, -f2
behave run examples/cut
```

`behave add` runs the original right away and stores what it did. To record a much
bigger spec automatically from `--help`, use `behave probe cut` (about 175 cases).
