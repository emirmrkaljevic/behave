---
title: behave MCP server
description: Run behave probe, record and diff over stdio from an MCP client.
---

# MCP server

`behave mcp` serves the same features over stdio. The main interface is still the CLI. This is for editor or MCP-client integrations that want to grow a spec while a port is being written, then check it after a change.

```json
{
  "mcpServers": {
    "behave": { "command": "behave", "args": ["mcp"] }
  }
}
```

The MCP Registry name is `io.github.emirmrkaljevic/behave`. The repo ships a multi-stage Dockerfile that builds the binary and starts `behave mcp` on stdio, for registries that build from a Dockerfile.

| tool | what it does |
|---|---|
| `behave_exec` | run a command once in a temp dir, return exit/stdout/stderr (exploration) |
| `behave_probe` | generate and record a spec from `--help` |
| `behave_add_case` | add a case and record the original's behavior |
| `behave_list_cases` | list cases, optionally by tag (`flag:--sort`, `manual`) |
| `behave_diff` | replay against an implementation, return failing cases with diffs |

[Back to behave](./) · [Source on GitHub](https://github.com/emirmrkaljevic/behave)
