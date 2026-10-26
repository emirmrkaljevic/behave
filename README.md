# behave

<p align="center"><a href="https://skintoscan.com/#packages"><img src=".github/assets/skintoscan-support.svg" width="684" alt="Support my open-source work: get a SkinToScan pack"></a><br><sub>If you find this useful, you can support me by getting a SkinToScan pack, my other project. It's a one-time purchase that directly funds my open-source work.</sub></p>
<p align="center"><a href="https://skintoscan.com/checkout/?add-to-cart=138"><img src=".github/assets/skintoscan-pack-1.svg" width="220" alt="Buy Single Picture Check, €4.49"></a> <a href="https://skintoscan.com/checkout/?add-to-cart=139"><img src=".github/assets/skintoscan-pack-5.svg" width="220" alt="Buy 5 Picture Pack, €19.99 (popular)"></a> <a href="https://skintoscan.com/checkout/?add-to-cart=140"><img src=".github/assets/skintoscan-pack-15.svg" width="220" alt="Buy 15 Picture Pack, €54.99"></a></p>

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

Install with Homebrew (macOS and Linux), download a binary from [Releases](https://github.com/emirmrkaljevic/behave/releases), or use `go install`:

```sh
brew install emirmrkaljevic/tap/behave
# or: go install github.com/emirmrkaljevic/behave/cmd/behave@latest

# 1. record how GNU sort behaves (writes ./behave-sort/)
behave probe sort

# 2. check that the recording is stable
behave run behave-sort

# 3. compare another implementation against it
behave diff behave-sort --impl ./target/release/sort --skip-tag meta
```

Or try the committed examples first: `behave run examples/cut` (27 GNU `cut` cases),
`behave run examples/paste` (18 GNU `paste` cases), `behave run examples/join` (26 GNU `join` cases), `behave run examples/nl` (26 GNU `nl` cases), `behave run examples/fold` (28 GNU `fold` cases), `behave run examples/head` (24 GNU `head` cases), or `behave run examples/jq` (33 jq 1.7 cases); see [examples/](examples/).

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
| GNU tac 9.12 | uutils tac (main) | 36 | 35 | `-s -1` rejected (value starting with `-`, [#15497](https://github.com/uutils/coreutils/issues/15497), now fixed) |
| GNU tail 9.12 | uutils tail (main) | 60 | 58 | `-c` with no count accepted ([#15528](https://github.com/uutils/coreutils/issues/15528), now fixed); file names like `007` printed as `7` ([#15533](https://github.com/uutils/coreutils/issues/15533), now fixed in tail); a hand-added `-cf FILE` case is rejected ([#15517](https://github.com/uutils/coreutils/issues/15517)) |
| GNU cut | uutils cut | 175 | 172 | `--output-delimiter -1` rejected (value starting with `-`) |
| GNU base64 | uutils base64 | 36 | 31 | `-d` on invalid input: GNU prints the bytes decoded so far, uutils prints nothing |
| GNU sort | uutils sort | 290 | ~260† | usage errors exit 1 instead of 2; `--debug` underline width; `--check --merge` |
| jq 1.7 | gojq 0.12.19 | 596 | 331* | unsupported flags (`--seq`, `--ascii-output`, …), key order, big-number formatting |

\* with `--ignore-stderr`; error message wording differs almost everywhere.
† varies by one or two: some `--random-sort` cases are only sometimes detected as flaky with tiny inputs.

Many of these are deliberate design choices, not bugs. behave just makes them visible.
Record from the release the rewrite actually targets. uutils follows the newest GNU release, and
GNU 9.12 prints `==> 'standard input' <==` in `tail -v` headers where 9.7 prints no quotes, so a
spec recorded from an older GNU reports differences that are really GNU changes. `behave.yaml`
keeps the `--version` line of the recorded tool, and `run`/`diff` print it next to the version
being replayed.
Some were real bugs and are now fixed upstream:

- uutils `cut --output-delimiter -x` ([uutils/coreutils#15341](https://github.com/uutils/coreutils/issues/15341), fixed in #15345)
- uutils `sort --check --merge` ignored the check ([uutils/coreutils#15380](https://github.com/uutils/coreutils/issues/15380), fixed in [#15381](https://github.com/uutils/coreutils/pull/15381))
- uutils `tac -s -1` rejected a separator starting with `-` ([uutils/coreutils#15497](https://github.com/uutils/coreutils/issues/15497), fixed in [#15498](https://github.com/uutils/coreutils/pull/15498))
- uutils `tail 007` printed the missing file as `'7'` ([uutils/coreutils#15533](https://github.com/uutils/coreutils/issues/15533), fixed for tail in [#15534](https://github.com/uutils/coreutils/pull/15534); the same bug in other utilities is being fixed one by one: `tac` in [#15541](https://github.com/uutils/coreutils/pull/15541) (merged), `nl` in [#15544](https://github.com/uutils/coreutils/pull/15544) (merged), `md5sum`/`cksum` in [#15548](https://github.com/uutils/coreutils/pull/15548) (merged), `sort` in [#15565](https://github.com/uutils/coreutils/pull/15565) (merged), `shred` in [#15604](https://github.com/uutils/coreutils/pull/15604) (merged))
- uutils bare `tail -c` and `tail +NUM` ignored `_POSIX2_VERSION` ([uutils/coreutils#15528](https://github.com/uutils/coreutils/issues/15528), fixed in [#15529](https://github.com/uutils/coreutils/pull/15529))
- uutils `sort -c dir` exited 0 without a message instead of reporting the read error ([uutils/coreutils#15615](https://github.com/uutils/coreutils/issues/15615), fixed in [#15616](https://github.com/uutils/coreutils/pull/15616))

Still open: a few error messages that lose the file name, mostly caught by the "file with trailing slash" probe case added in v0.1.7:

- `cp f/ dest` prints `Not a directory (os error 20)` instead of `cannot stat 'f/': Not a directory` ([#15691](https://github.com/uutils/coreutils/issues/15691), fix in [#15692](https://github.com/uutils/coreutils/pull/15692))
- `paste` open and read errors without the file name ([#15254](https://github.com/uutils/coreutils/issues/15254), fix in [#15642](https://github.com/uutils/coreutils/pull/15642))
- `shred dir` says `Not a file` instead of `failed to open for writing: Is a directory` ([#15630](https://github.com/uutils/coreutils/issues/15630), fix in [#15631](https://github.com/uutils/coreutils/pull/15631))
- `mv missing/ g` says `Not a directory` and `mv d/f g` with an unsearchable `d` says `No such file or directory`, where GNU gives the real reason ([#15704](https://github.com/uutils/coreutils/issues/15704), fix in [#15705](https://github.com/uutils/coreutils/pull/15705))
- `ln f/ link` and `ln missing link` say `failed to create hard link 'link' => ...` instead of naming the source with `failed to access 'f/'` ([#15713](https://github.com/uutils/coreutils/issues/15713), fix in [#15714](https://github.com/uutils/coreutils/pull/15714))
- `mv a missing/b` and `mv a ro/b` print only `No such file or directory (os error 2)` or `Permission denied (os error 13)`, where GNU says `cannot move 'a' to 'missing/b': ...` ([#15721](https://github.com/uutils/coreutils/issues/15721), fix in [#15722](https://github.com/uutils/coreutils/pull/15722))
- `mv -T f d` (file onto an existing directory) prints `cannot overwrite directory 'd' with non-directory` without the source, where GNU adds `'f'` ([#15724](https://github.com/uutils/coreutils/issues/15724), fix in [#15725](https://github.com/uutils/coreutils/pull/15725))
- `mv -n z z` (and `--update=none`) fails with `'z' and 'z' are the same file`, where GNU skips the move and exits 0 ([#15824](https://github.com/uutils/coreutils/issues/15824), fix in [#15825](https://github.com/uutils/coreutils/pull/15825))

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

Cases can also set `env`, `files` (created in a fresh temp dir), and `timeout_ms`.
Add your own with `behave add behave-cut --stdin 'a,b' -- -d, -f2`.

## How it stays deterministic

- every case runs in a new temp directory with a fixed environment
  (`LC_ALL=C.UTF-8`, `TZ=UTC`, `NO_COLOR=1`, `HOME`=temp dir)
- input files get fixed modification times, and tools that read the current directory
  (`ls`, `du`, `rg`; help text mentions "current directory" or recursion) get a small sample
  tree (`--sample-files auto|on|off`) whose names, sizes and extensions sort differently,
  so `ls --sort=size`, `ls -t`, `du --time` and `du -S` produce output worth comparing
- each case is recorded 3 times; anything that changes between runs is marked `flaky`
  and skipped (e.g. `sort --random-sort`)
- normalizers remove noise before comparing: `ansi`, `workdir`, `cmd` (`jq:` vs `gojq:`),
  `timestamp`, `hexaddr`, `duration`, `pid`, plus your own regex rules in `behave.yaml`
- timeouts kill the whole process group

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
