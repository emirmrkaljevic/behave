---
title: Test a Rust rewrite of a CLI against the original
description: How to check that a Rust or Go port of a command-line tool matches the original's output, errors and exit codes, using recorded black-box cases.
---

# Test a Rust (or Go) rewrite of a CLI against the original

Rewriting a well known command-line tool is mostly about the long tail: the exact error message, the exit code for a missing file, how a flag that takes a value starting with `-` is parsed. Users and scripts depend on all of it, and most of it is not written down anywhere.

behave builds that specification for you by observing the original tool.

## 1. Record the original

```sh
behave probe sort
```

This reads `sort --help` and the man page, generates cases for every flag, flag pairs and invalid input, runs each one in a fresh temp directory with a fixed environment, and writes them to `./behave-sort/` as plain files you can commit.

## 2. Check the recording is stable

```sh
behave run behave-sort
```

Cases that depend on time, randomness or the machine get flagged so they do not create noise later.

## 3. Diff your port

```sh
behave diff behave-sort --impl ./target/release/sort
```

Every case where your build differs is listed with the expected and actual stdout, stderr and exit code. Use `--skip-tag` to ignore groups such as `meta` (version and help text) that are expected to differ.

## 4. Keep it in CI

The recorded case directory is ordinary text. Commit it and run `behave run` or `behave diff` in CI so a regression in an edge case fails the build.

## Does this find real bugs?

Run against uutils coreutils (the Rust rewrite of GNU coreutils), it led to 17 upstream bug reports and 24 fix PRs, 11 merged so far. See the [uutils case study](uutils-case-study.html).

[Back to behave](./) · [Source on GitHub](https://github.com/emirmrkaljevic/behave)
