# behave

**Record how any CLI behaves. Prove your rewrite behaves the same.**

`behave` explores a command-line tool as a black box (flags, stdin, files, exit codes, edge
cases), records hundreds of golden test cases, and replays them against a new implementation,
so you can show that your Rust/Go port, refactor, or upgrade is behavior-compatible.

> Status: early development. Not usable yet.

## Planned

- `behave probe jq` — discover flags from help/man and generate edge-case inputs
- `behave record` — store cases as plain, reviewable YAML
- `behave diff spec/ --impl ./my-jq` — side-by-side diff + coverage report
- `behave mcp` — MCP server so coding assistants can drive exploration
- Normalizers for timestamps, temp paths, PIDs, ANSI colors

## Install (soon)

```sh
go install github.com/emirmrkaljevic/behave/cmd/behave@latest
```

## License

MIT
