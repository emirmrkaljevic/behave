---
title: behave commands, spec format and CI notes
description: Command reference, the YAML spec layout, how recordings stay deterministic, and notes on the example CI jobs.
---

# Commands, spec format and CI notes

## Commands

| command | what it does |
|---|---|
| `behave probe <cmd>` | discover flags, generate and record cases (`--arg`, `--inputs text\|json\|both`, `--max`, `--skip-flag`, `--sample-files`, `--dry-run`) |
| `behave record <spec>` | re-record expectations from the original |
| `behave add <spec> -- ARGS` | add one hand-written case |
| `behave run <spec>` | replay against the original |
| `behave diff <spec> --impl <bin>` | replay against another implementation (`--ignore-stderr`, `--ignore-exit`, `--tag`, `--skip-tag`, `--json`, `--junit file`, `-v`) |
| `behave coverage <spec>` | which documented flags (from `--help`/man) the cases exercise (`--min PCT`, `--all`, `--json`) |
| `behave normalizers` | list built-in normalizers |
| `behave mcp` | serve the same features over stdio (see [MCP](mcp.md)) |

`run` and `diff` exit with 1 when any case fails, so they drop straight into CI. Pass `--junit report.xml` to write a JUnit XML report for CI dashboards.

When the help text lists the values a flag accepts, `probe` adds one case per value next to the invalid-value cases. It understands GNU style (`WHEN can be: auto, always, never`), clap's `[possible values: ...]`, argparse's `{json,yaml}` and `<fast|safe>`. It only keeps plain lowercase lists it can tie to that flag. `--dry-run` shows what it found:

```
$ behave probe ls --dry-run
       --color                      WHEN  {always,auto,never}
       --sort                       WORD  {none,size,time,version,extension,name,width}
```

`coverage` tells you what a spec doesn't test yet. It reads flags the same way `probe` does, understands `--long=value`, unambiguous long prefixes and clustered short flags (`-sf1`), and exits 1 with `--min 80` when coverage drops below 80%:

```
$ behave coverage examples/cut
cut: 27 cases exercise 7 of 12 documented flags (58%)
not exercised (5): -n --zero-terminated --help --version -M
used but not in help: --no-such-flag
```

## Spec format

A spec is a directory you can commit and review:

```
behave-cut/
  behave.yaml           # command, timeout, normalizers, recorded_from (--version line)
  cases/
    0001-help.yaml
    0135-flag-output-delimiter-value-1-arg-f1.yaml
    ...
```

```yaml
name: flag --output-delimiter, value -1, arg -f1
args: [--output-delimiter, "-1", -f1]
stdin: |
  hello
tags: [flag-value, "flag:--output-delimiter"]
expect:
  exit: 0
  stdout: |
    hello
  stderr: ""
```

Cases can also set `env`, `files` (created in a fresh temp dir), and `timeout_ms`. Add your own with `behave add behave-cut --stdin 'a,b' -- -d, -f2`.

## How recordings stay deterministic

- every case runs in a new temp directory with a fixed environment (`LC_ALL=C.UTF-8`, `TZ=UTC`, `NO_COLOR=1`, `HOME`=temp dir)
- input files get fixed modification times, and tools that read the current directory (`ls`, `du`, `rg`; help text mentions "current directory" or recursion) get a small sample tree (`--sample-files auto|on|off`) whose names, sizes and extensions sort differently, so `ls --sort=size`, `ls -t`, `du --time` and `du -S` produce output worth comparing
- each case is recorded 3 times; anything that changes between runs is marked `flaky` and skipped (for example `sort --random-sort`)
- normalizers remove noise before comparing: `ansi`, `workdir`, `cmd` (`jq:` vs `gojq:`), `timestamp`, `hexaddr`, `duration`, `pid`, plus your own regex rules in `behave.yaml`
- timeouts kill the whole process group

## CI notes

GitHub Actions runs `go vet`, `go test -race`, a smoke `probe`/`run` on `wc`, then most of the committed example specs.

A few cases depend on a newer GNU than the runner ships (currently 9.4):

- since GNU coreutils 9.5, `base32`/`base64` `-d` accepts input without padding. Those cases are tagged `gnu-9.5` and skipped in CI: `behave run examples/base32 --skip-tag gnu-9.5` (same for `base64`)
- `ptx -t` output changed after GNU 9.7, so that case is tagged `gnu-9.12` and skipped in CI

`examples/sysctl` needs the sbin directories on `PATH`. There is also a small self-check: `behave run examples/behave`.

The `fmt`, `cksum`, `md5sum` and `tail` example specs are in the repo but are not part of the CI workflow yet.

[Back to behave](./) · [Source on GitHub](https://github.com/emirmrkaljevic/behave)
