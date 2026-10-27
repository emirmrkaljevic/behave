---
title: Golden tests for CLI tools, generated from --help
description: Generate golden (snapshot) tests for any command-line tool by recording stdout, stderr and exit codes for hundreds of generated cases.
---

# Golden tests for CLI tools, generated from --help

Golden tests (snapshot tests) compare a program's output with a stored known good output. For a CLI that means stdout, stderr and the exit code for each invocation. Writing them by hand is slow, so most tools only cover the happy path.

behave generates them:

```sh
behave probe mytool          # generate and record cases into ./behave-mytool/
behave run behave-mytool     # replay; fails if anything changed
```

## What gets generated

- every flag from `--help`, alone and in pairs
- every value the help text lists for a flag
- edge case stdin (empty, no trailing newline, long lines, binary)
- file arguments, missing files, directories where a file is expected
- bad values and unknown flags

## Why this works for refactors and upgrades

After a refactor, a dependency bump or a new argument parser, `behave run` tells you within seconds whether any observable behaviour changed, including error messages that unit tests rarely check.

## Comparing two implementations

The same recording can be replayed against a different binary with `behave diff --impl`. That is how the repo's [jq example](https://github.com/emirmrkaljevic/behave/tree/main/examples/jq) checks gojq against jq 1.7: 29 of 33 cases match and the 4 differences are error wording on stderr.

[Back to behave](./) · [Source on GitHub](https://github.com/emirmrkaljevic/behave)
