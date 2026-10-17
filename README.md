# behave

<p align="center"><a href="https://skintoscan.com/#packages"><img src="https://img.shields.io/badge/%E2%98%95_Support_me-Get_a_SkinToScan_pack-ff5f5f?style=for-the-badge" alt="Support me: get a SkinToScan pack"></a><br><sub>If you find this useful, you can support me by getting a SkinToScan pack, my other project. It's a one-time purchase that directly funds my open-source work.</sub></p>

**Record how any CLI behaves. Prove your rewrite behaves the same.**

`behave` treats a command-line tool as a black box. It reads `--help` (and the man page),
generates hundreds of test cases (every flag, flag pairs, every value the help lists for a flag,
edge-case stdin, file arguments, bad values, missing files), records what the tool actually does (stdout, stderr, exit code),
and replays those cases against another implementation.

![behave recording jq and checking gojq against it](docs/demo.gif)

Use it when you are:

- porting a tool to Rust or Go and want proof that the port is compatible,
- refactoring or upgrading a CLI and want a regression net in a few seconds,
- comparing two implementations of the same tool.

> Status: early (v0.1). CLI and MCP server work; feedback welcome.

## Quick start

Download a binary for Linux or macOS from [Releases](https://github.com/emirmrkaljevic/behave/releases), or:

```sh
go install github.com/emirmrkaljevic/behave/cmd/behave@latest

# 1. record how GNU sort behaves (writes ./behave-sort/)
behave probe sort

# 2. check that the recording is stable
behave run behave-sort

# 3. compare another implementation against it
behave diff behave-sort --impl ./target/release/sort --skip-tag meta
```

Or try the committed examples first: `behave run examples/cut` (27 GNU `cut` cases),
`behave run examples/paste` (18 GNU `paste` cases), `behave run examples/join` (26 GNU `join` cases), `behave run examples/nl` (26 GNU `nl` cases), `behave run examples/fold` (28 GNU `fold` cases), or `behave run examples/jq` (33 jq 1.7 cases); see [examples/](examples/).

For tools that need a positional argument (a filter, a pattern, a field list), give samples:

```sh
behave probe jq --arg . --arg '.a' --arg '.[]' --arg 'map(.+1)' --arg keys
behave diff behave-jq --impl gojq --ignore-stderr
```

When the help text lists the values a flag accepts, `probe` adds one case per value next to the
invalid-value cases. It understands GNU style (`WHEN can be: auto, always, never`,
`--sort=WORD ... to WORD: none (-U), size (-S), ...`, `The WHEN argument defaults to 'always' and can
also be 'auto' or 'never'.`), clap's `[possible values: ...]`, argparse's `{json,yaml}` and
`<fast|safe>`. It only keeps plain lowercase lists it can tie to that flag. `--dry-run` shows what it found:

```
$ behave probe ls --dry-run
       --color                      WHEN  {always,auto,never}
       --sort                       WORD  {none,size,time,version,extension,name,width}
```

Example output (abridged):

```
✗ flag --output-delimiter, value -1, arg -f1  0135-flag-output-delimiter-value-1-arg-f1.yaml
  $ cut --output-delimiter -1 -f1
  exit:
    exit: want 0, got 1
  stdout:
    - hello

flags with differences:
  --output-delimiter             2/16  cases differ
flags fully compatible: 9 of 10

cut vs ./uu/cut: 172/175 passed, 3 failed, 0 skipped  (differs in: exit 2, stdout 3)  [0.1s]
```

## What it found on day one

Run on this machine against real tools, generated cases only, `--help`/`--version` excluded:

| original | candidate | cases | pass | notable differences |
|---|---|---|---|---|
| GNU wc 9.7 | uutils wc 0.12.0 | 63 | 63 | none |
| GNU head | uutils head | 51 | 51 | none |
| GNU uniq | uutils uniq | 100 | 100 | none |
| GNU tr | uutils tr | 63 | 63 | none |
| GNU cut | uutils cut | 175 | 172 | `--output-delimiter -1` rejected (value starting with `-`) |
| GNU base64 | uutils base64 | 36 | 31 | `-d` on invalid input: GNU prints the bytes decoded so far, uutils prints nothing |
| GNU sort | uutils sort | 290 | ~260† | usage errors exit 1 instead of 2; `--debug` underline width; `--check --merge` |
| jq 1.7 | gojq 0.12.19 | 596 | 331* | unsupported flags (`--seq`, `--ascii-output`, …), key order, big-number formatting |

\* with `--ignore-stderr`; error message wording differs almost everywhere.
† varies by one or two: some `--random-sort` cases are only sometimes detected as flaky with tiny inputs.

Many of these are deliberate design choices, not bugs. behave just makes them visible.
Some were real bugs and are now fixed upstream:

- uutils `cut --output-delimiter -x` ([uutils/coreutils#15341](https://github.com/uutils/coreutils/issues/15341), fixed in #15345)
- uutils `sort --check --merge` ignored the check ([uutils/coreutils#15380](https://github.com/uutils/coreutils/issues/15380), fixed in [#15381](https://github.com/uutils/coreutils/pull/15381))

## MCP server

`behave mcp` serves the same features over stdio, so a coding assistant can explore the
original tool, grow the spec while it works, and check its port after every change.

```json
{
  "mcpServers": {
    "behave": { "command": "behave", "args": ["mcp"] }
  }
}
```

| tool | what it does |
|---|---|
| `behave_exec` | run a command once in a temp dir, return exit/stdout/stderr (exploration) |
| `behave_probe` | generate and record a spec from `--help` |
| `behave_add_case` | add a case and record the original's behavior |
| `behave_list_cases` | list cases, optionally by tag (`flag:--sort`, `manual`) |
| `behave_diff` | replay against an implementation, return failing cases with diffs |

## Spec format

A spec is a directory you can commit and review:

```
behave-cut/
  behave.yaml           # command, timeout, normalizers
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

Cases can also set `env`, `files` (created in a fresh temp dir), and `timeout_ms`.
Add your own with `behave add behave-cut --stdin 'a,b' -- -d, -f2`.

## How it stays deterministic

- every case runs in a new temp directory with a fixed environment
  (`LC_ALL=C.UTF-8`, `TZ=UTC`, `NO_COLOR=1`, `HOME`=temp dir)
- each case is recorded 3 times; anything that changes between runs is marked `flaky`
  and skipped (e.g. `sort --random-sort`)
- normalizers remove noise before comparing: `ansi`, `workdir`, `cmd` (`jq:` vs `gojq:`),
  `timestamp`, `hexaddr`, `duration`, `pid`, plus your own regex rules in `behave.yaml`
- timeouts kill the whole process group

## Commands

| command | what it does |
|---|---|
| `behave probe <cmd>` | discover flags, generate and record cases (`--arg`, `--inputs text\|json\|both`, `--max`, `--skip-flag`, `--dry-run`) |
| `behave record <spec>` | re-record expectations from the original |
| `behave add <spec> -- ARGS` | add one hand-written case |
| `behave run <spec>` | replay against the original |
| `behave diff <spec> --impl <bin>` | replay against another implementation (`--ignore-stderr`, `--ignore-exit`, `--tag`, `--skip-tag`, `--json`, `--junit file`, `-v`) |
| `behave coverage <spec>` | which documented flags (from `--help`/man) the cases exercise (`--min PCT`, `--all`, `--json`) |
| `behave normalizers` | list built-in normalizers |

`run` and `diff` exit with 1 when any case fails, so they drop straight into CI. Pass `--junit report.xml` to write a JUnit XML report for CI dashboards (GitHub Actions, Jenkins, etc.).

`coverage` tells you what a spec doesn't test yet. On the hand-picked `cut` example:

```
$ behave coverage examples/cut
cut: 27 cases exercise 7 of 12 documented flags (58%)
not exercised (5): -n --zero-terminated --help --version -M
used but not in help: --no-such-flag
```

It reads flags the same way `probe` does, understands `--long=value`, unambiguous long prefixes and clustered short flags (`-sf1`), and exits 1 with `--min 80` when coverage drops below 80%.

The repo ships a multi-stage `Dockerfile` that builds the binary and starts `behave mcp` on stdio (for Glama and other MCP registries that build from a Dockerfile).

## Roadmap

- more committed example specs (`head`, `tail`, `wc`, a `grep` subset vs ripgrep)
- Homebrew tap

## Safety

behave only runs the binaries you point it at, in a temp directory. Don't probe tools that
change things outside their working directory (package managers, `rm -rf`-style tools)
without `--skip-flag` and a sandbox.

## Links

- MCP Registry name: `io.github.emirmrkaljevic/behave`

## License

MIT
