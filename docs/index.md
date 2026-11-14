---
title: behave, compatibility tests for CLI tools
description: Record how any command-line tool behaves and check a rewrite or new version against it. Found 41 bugs in uutils rewrites, 34 fix PRs merged.
---

# Record how a CLI behaves, then check a rewrite against it

behave treats a command-line tool as a black box. It reads `--help` and the man page, generates cases (every flag, flag pairs, listed values, edge case stdin, missing files, bad values), records stdout, stderr and the exit code, and replays the cases against another implementation.

![behave recording jq and checking gojq against it](demo.gif)

```sh
brew install emirmrkaljevic/tap/behave
behave probe sort
behave run behave-sort
behave diff behave-sort --impl ./target/release/sort
```

## Guides

- [Test a Rust or Go rewrite of a CLI against the original](rewrite-compatibility.html)
- [Golden tests for CLI tools without writing them by hand](golden-tests-cli.html)
- [Case study: what behave found in uutils coreutils](uutils-case-study.html)
- [Example specs](examples.html)
- [Upstream findings](upstream-findings.html)
- [Commands, spec format and CI notes](usage.html)
- [MCP server](mcp.html)

Source, releases and the MCP server: [github.com/emirmrkaljevic/behave](https://github.com/emirmrkaljevic/behave)
