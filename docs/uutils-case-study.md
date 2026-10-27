---
title: "Case study: compatibility bugs found in uutils coreutils"
description: Black-box testing of uutils coreutils (Rust) against GNU coreutils 9.12 with behave led to 17 bug reports and 24 fix PRs.
---

# Case study: uutils coreutils vs GNU coreutils

[uutils coreutils](https://github.com/uutils/coreutils) is a Rust rewrite of GNU coreutils that aims to be a drop in replacement. behave recorded GNU coreutils 9.12 and replayed the cases against uutils built from main.

As of Oct 27, 2026 this led to **17 bug reports and 24 fix PRs upstream, 11 of them merged**. Every issue and PR says it was found with behave and each fix comes with a test.

## Examples of what it found

- `sort --check --merge` ignored the check ([#15380](https://github.com/uutils/coreutils/issues/15380), fixed in [#15381](https://github.com/uutils/coreutils/pull/15381))
- `tac -s -1` rejected a separator starting with `-` ([#15497](https://github.com/uutils/coreutils/issues/15497), fixed in [#15498](https://github.com/uutils/coreutils/pull/15498))
- `tail 007` printed the missing file as `'7'` ([#15533](https://github.com/uutils/coreutils/issues/15533), fixed in [#15534](https://github.com/uutils/coreutils/pull/15534), then in tac, nl, sort, shred and the checksum tools)
- bare `tail -c` and `tail +NUM` ignored `_POSIX2_VERSION` ([#15528](https://github.com/uutils/coreutils/issues/15528), fixed in [#15529](https://github.com/uutils/coreutils/pull/15529))
- `sort -c dir` exited 0 with no message ([#15615](https://github.com/uutils/coreutils/issues/15615), fixed in [#15616](https://github.com/uutils/coreutils/pull/15616))
- `cut --output-delimiter -x` was rejected ([#15341](https://github.com/uutils/coreutils/issues/15341), fixed in #15345)
- error messages without the file name or context in cp, mv, ln and chmod (open PRs [#15692](https://github.com/uutils/coreutils/pull/15692), [#15722](https://github.com/uutils/coreutils/pull/15722), [#15714](https://github.com/uutils/coreutils/pull/15714), [#15884](https://github.com/uutils/coreutils/pull/15884))

## What kind of bugs a black-box probe finds

Mostly the long tail: argument parsing of values that look like flags, error wording and context, exit codes on read errors, and locale formatting leaking into file names. These are exactly the things scripts depend on and unit tests rarely cover.

## Try it on your own port

See [Test a Rust rewrite of a CLI against the original](rewrite-compatibility.html).

[Back to behave](./) · [Source on GitHub](https://github.com/emirmrkaljevic/behave)
