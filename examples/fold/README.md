# `fold/` — 28 recorded cases for GNU `fold`

Default width 80, `-w`/`--width` (including the obsolete `-5` form and width 1),
`-s`/`--spaces` breaking at blanks, `-b`/`--bytes`, tab stops, backspace and carriage
return column handling, multibyte input, stdin/`-`/files, empty input, missing
trailing newline, and error paths. Recorded with GNU coreutils 9.7.

```sh
behave run examples/fold
# fold vs fold: 28/28 passed, 0 failed, 0 skipped
```

Against uutils 0.12.0:

```sh
behave diff examples/fold --impl ./coreutils-0.12.0/fold
# 24/28 — three stderr-only errors, plus multibyte input (stdout)
behave diff examples/fold --impl ./coreutils-0.12.0/fold --ignore-stderr
# 27/28 — remaining miss is `fold -w 4` on UTF-8 text
```

The multibyte case is a version difference, not a uutils bug: GNU fold counted bytes
until coreutils 9.8, which made it multibyte aware (wrapping by display column, with
`-c`/`--characters` added). uutils already wraps by column, so it matches 9.8+ here.
Re-record that case with `behave record examples/fold` if your system has 9.8 or newer.
