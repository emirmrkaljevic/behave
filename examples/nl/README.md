# `nl/` — 26 recorded cases for GNU `nl`

Body numbering (`-b a`/`t`/`n`/`pPATTERN`), number formats (`-n ln`/`rn`/`rz`),
`-w` width, `-s` separator, `-v` start, `-i` increment, `-l` blank-line grouping,
section delimiters (`-d`, `-p`), stdin/file/multi-file, empty input, and error paths.
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
