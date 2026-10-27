# `jq/`: 33 recorded cases for GNU/`jq` 1.7

Identity and field access, nested paths, array index/iterate, object and array
construction, builtins (`length`, `keys`, `type`, `map`, `select`, `add`, `sort`),
flags (`-r`, `-c`, `-n`, `-s`, `-e`, `--arg`, `--argjson`), stdin edge cases, a file
argument, and error paths. Recorded with jq 1.7.

```sh
behave run examples/jq
# jq vs jq: 33/33 passed, 0 failed, 0 skipped
```

Against [gojq](https://github.com/itchyny/gojq) 0.12.19 (stderr wording differs on errors;
output and exits match; gojq documents its error messages as intentionally different):

```sh
behave diff examples/jq --impl gojq
# 29/33: all four failures are stderr only (invalid JSON, bad filter, unknown option, 1/0 wording)
behave diff examples/jq --impl gojq --ignore-stderr
# 33/33
```

Against [jaq](https://github.com/01mf02/jaq) 3.1.1:

```sh
behave diff examples/jq --impl jaq
# 29/33: three stderr only errors, plus division by zero
behave diff examples/jq --impl jaq --ignore-stderr
# 32/33, remaining: `1/0` → `Infinity` exit 0 (jq errors exit 5)
```

The `/0` case is a known intentional difference in jaq (IEEE-style floats; see
[jaq#230](https://github.com/01mf02/jaq/issues/230)), not a bug to fix. Error-message
wording also differs across implementations; use `--ignore-stderr` when comparing ports.
