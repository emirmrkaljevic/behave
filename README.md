# behave

<p align="center"><a href="https://skintoscan.com/#packages"><img src=".github/assets/skintoscan-support.svg" width="684" alt="Support my open-source work: get a SkinToScan pack"></a><br><sub>If you find this useful, you can support me by getting a SkinToScan pack, my other project. It's a one-time purchase that directly funds my open-source work.</sub></p>
<p align="center"><a href="https://skintoscan.com/checkout/?add-to-cart=138"><img src=".github/assets/skintoscan-pack-1.svg" width="220" alt="Buy Single Picture Check, €4.49"></a> <a href="https://skintoscan.com/checkout/?add-to-cart=139"><img src=".github/assets/skintoscan-pack-5.svg" width="220" alt="Buy 5 Picture Pack, €19.99 (popular)"></a> <a href="https://skintoscan.com/checkout/?add-to-cart=140"><img src=".github/assets/skintoscan-pack-15.svg" width="220" alt="Buy 15 Picture Pack, €54.99"></a></p>

behave records how a command-line tool actually behaves, then replays those cases against another implementation. I built it for CLI compatibility testing: checking a Rust or Go port against the original, or catching a refactor that changed visible behaviour.

It treats the tool as a black box. It reads `--help` and the man page, generates cases (flags, flag pairs, listed values, odd stdin, missing files, bad values), records stdout, stderr and the exit code, and diffs a second binary against that recording.

![behave recording jq and checking gojq against it](docs/demo.gif)

I have used it on the Rust rewrites of GNU coreutils, findutils, sed and procps. So far that has led to 44 upstream bug reports and 51 fix PRs, 34 of them merged (as of 15 Nov 2026). The full list is in [upstream findings](docs/upstream-findings.md).

## Install

Homebrew (macOS and Linux), a binary from [Releases](https://github.com/emirmrkaljevic/behave/releases), or `go install`:

```sh
brew install emirmrkaljevic/tap/behave
# or: go install github.com/emirmrkaljevic/behave/cmd/behave@latest
```

## Example

Record GNU `sort`, check that the recording is stable, then compare a rewrite:

```sh
behave probe sort
behave run behave-sort
behave diff behave-sort --impl ./target/release/sort --skip-tag meta
```

`--skip-tag meta` drops `--help` and `--version`, which a rewrite is allowed to word differently.

A failing case looks like this (abridged):

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

The recording is a directory of YAML files you can read and commit. `behave.yaml` stores the version of the tool you recorded, and `run` / `diff` print it next to the version being replayed.

The repo also ships ready-made specs under [examples/](examples/). `behave run examples/cut` is a small GNU `cut` suite; `behave diff examples/jq --impl gojq` checks gojq against jq 1.7. The full list, with case counts, is in [examples](docs/examples.md).

For tools that need a positional argument, pass samples with `--arg`:

```sh
behave probe jq --arg . --arg '.a' --arg keys
```

Add a hand-written case with `behave add behave-cut --stdin 'a,b' -- -d, -f2`. `behave coverage` shows which documented flags the spec does not exercise yet. `behave probe --dry-run` lists the flags and values it would cover, without running the tool; `--skip-flag` and `--max` cut the set down when a tool is large or unsafe to probe in full. The [commands and spec format](docs/usage.md) page covers the rest of the CLI, including JUnit output for CI.

Each case runs in a fresh temp directory with a fixed environment (`LC_ALL=C.UTF-8`, `TZ=UTC`). Anything that changes between runs is marked flaky and skipped.

A longer walk-through of this flow, including how to keep the recording in CI, is in [Test a rewrite against the original](docs/guides/rewrite-compatibility.md). If you want snapshot tests for a tool you already own, rather than a comparison against a second binary, the same recording is enough: see [Golden tests from --help](docs/guides/golden-tests-cli.md).

## MCP (optional)

`behave mcp` exposes the same probe, record and diff flow over stdio, if you want to drive it from an editor. Setup and tools are in [MCP server](docs/mcp.md).

[![emirmrkaljevic/behave MCP server](https://glama.ai/mcp/servers/emirmrkaljevic/behave/badges/score.svg)](https://glama.ai/mcp/servers/emirmrkaljevic/behave)

## Limitations

behave only runs the binaries you point it at, in a fresh temp directory. Do not probe tools that change things outside their working directory (package managers, `rm -rf` style tools) without `--skip-flag` and a sandbox.

Recordings are only as good as the original you pointed at. Record from the release the rewrite actually targets. GNU coreutils 9.12 and 9.7 already disagree on some `tail -v` headers. Cases that change between runs, such as `sort --random-sort`, are marked flaky and skipped. Some differences are deliberate design choices, not bugs.

Status: early (v0.1). The CLI works; feedback is welcome.

## Docs

A browsable copy of these pages is at [emirmrkaljevic.github.io/behave](https://emirmrkaljevic.github.io/behave/).

- [Examples](docs/examples.md): ready-made recordings for GNU coreutils, findutils, procps-ng and jq
- [Commands, spec format and CI notes](docs/usage.md): CLI reference, the YAML layout, and how to run the specs in CI
- [MCP server](docs/mcp.md): the same probe, record and diff flow over stdio
- [What it found upstream](docs/upstream-findings.md): bug reports and fix PRs in uutils coreutils, findutils, sed and procps
- [Guides](docs/guides/): testing a rewrite, generating golden tests, and a uutils case study
  - [Test a rewrite against the original](docs/guides/rewrite-compatibility.md): record the original CLI and diff a Rust or Go port
  - [Golden tests from --help](docs/guides/golden-tests-cli.md): generate snapshot tests for stdout, stderr and exit codes
  - [Case study: uutils coreutils](docs/guides/uutils-case-study.md): what behave found when comparing the Rust rewrite to GNU

MCP Registry name: `io.github.emirmrkaljevic/behave`

## License

MIT
