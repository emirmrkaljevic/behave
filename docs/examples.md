---
title: Example specs shipped with behave
description: Ready-made behave recordings for GNU coreutils, findutils, procps-ng and jq.
---

# Example specs

The [examples/](https://github.com/emirmrkaljevic/behave/tree/main/examples) directory holds committed recordings you can run without probing anything first. Most were recorded against GNU coreutils 9.12. A few early ones (`cut`, `paste`, `join`, `nl`, `fold`, `jq`) were recorded by hand against 9.7 or jq 1.7. `xargs` is GNU findutils 4.10, and `sysctl` is procps-ng 4.0.4.

```sh
behave run examples/cut
behave diff examples/jq --impl gojq
```

On macOS the system tools are often BSD, so `behave run examples/cut` will show real differences. With Homebrew coreutils, compare against the GNU binary instead: `behave diff examples/cut --impl gcut`.

Per-tool notes, including what each suite covers and how uutils or gojq compared when the spec was added, live in [examples/README.md](https://github.com/emirmrkaljevic/behave/blob/main/examples/README.md).

| spec | cases | recorded from |
|---|---:|---|
| [cut](https://github.com/emirmrkaljevic/behave/tree/main/examples/cut) | 27 | GNU cut (9.7, hand-written) |
| [paste](https://github.com/emirmrkaljevic/behave/tree/main/examples/paste) | 18 | GNU paste (9.7, hand-written) |
| [join](https://github.com/emirmrkaljevic/behave/tree/main/examples/join) | 26 | GNU join (9.7, hand-written) |
| [nl](https://github.com/emirmrkaljevic/behave/tree/main/examples/nl) | 26 | GNU nl (9.7, hand-written) |
| [fold](https://github.com/emirmrkaljevic/behave/tree/main/examples/fold) | 28 | GNU fold (9.7, hand-written) |
| [head](https://github.com/emirmrkaljevic/behave/tree/main/examples/head) | 24 | GNU coreutils 9.12 |
| [expand](https://github.com/emirmrkaljevic/behave/tree/main/examples/expand) | 21 | GNU coreutils 9.12 |
| [unexpand](https://github.com/emirmrkaljevic/behave/tree/main/examples/unexpand) | 26 | GNU coreutils 9.12 |
| [tr](https://github.com/emirmrkaljevic/behave/tree/main/examples/tr) | 29 | GNU coreutils 9.12 |
| [comm](https://github.com/emirmrkaljevic/behave/tree/main/examples/comm) | 30 | GNU coreutils 9.12 |
| [uniq](https://github.com/emirmrkaljevic/behave/tree/main/examples/uniq) | 31 | GNU coreutils 9.12 |
| [sort](https://github.com/emirmrkaljevic/behave/tree/main/examples/sort) | 35 | GNU coreutils 9.12 |
| [cat](https://github.com/emirmrkaljevic/behave/tree/main/examples/cat) | 39 | GNU coreutils 9.12 |
| [wc](https://github.com/emirmrkaljevic/behave/tree/main/examples/wc) | 38 | GNU coreutils 9.12 |
| [tac](https://github.com/emirmrkaljevic/behave/tree/main/examples/tac) | 33 | GNU coreutils 9.12 |
| [numfmt](https://github.com/emirmrkaljevic/behave/tree/main/examples/numfmt) | 41 | GNU coreutils 9.12 |
| [od](https://github.com/emirmrkaljevic/behave/tree/main/examples/od) | 36 | GNU coreutils 9.12 |
| [basenc](https://github.com/emirmrkaljevic/behave/tree/main/examples/basenc) | 35 | GNU coreutils 9.12 |
| [seq](https://github.com/emirmrkaljevic/behave/tree/main/examples/seq) | 39 | GNU coreutils 9.12 |
| [expr](https://github.com/emirmrkaljevic/behave/tree/main/examples/expr) | 43 | GNU coreutils 9.12 |
| [printf](https://github.com/emirmrkaljevic/behave/tree/main/examples/printf) | 38 | GNU coreutils 9.12 |
| [factor](https://github.com/emirmrkaljevic/behave/tree/main/examples/factor) | 31 | GNU coreutils 9.12 |
| [cksum](https://github.com/emirmrkaljevic/behave/tree/main/examples/cksum) | 35 | GNU coreutils 9.12 |
| [md5sum](https://github.com/emirmrkaljevic/behave/tree/main/examples/md5sum) | 37 | GNU coreutils 9.12 |
| [tail](https://github.com/emirmrkaljevic/behave/tree/main/examples/tail) | 33 | GNU coreutils 9.12 |
| [base64](https://github.com/emirmrkaljevic/behave/tree/main/examples/base64) | 34 | GNU coreutils 9.12 |
| [base32](https://github.com/emirmrkaljevic/behave/tree/main/examples/base32) | 36 | GNU coreutils 9.12 |
| [basename](https://github.com/emirmrkaljevic/behave/tree/main/examples/basename) | 35 | GNU coreutils 9.12 |
| [dirname](https://github.com/emirmrkaljevic/behave/tree/main/examples/dirname) | 26 | GNU coreutils 9.12 |
| [fmt](https://github.com/emirmrkaljevic/behave/tree/main/examples/fmt) | 36 | GNU coreutils 9.12 |
| [realpath](https://github.com/emirmrkaljevic/behave/tree/main/examples/realpath) | 38 | GNU coreutils 9.12 |
| [pr](https://github.com/emirmrkaljevic/behave/tree/main/examples/pr) | 40 | GNU coreutils 9.12 |
| [tee](https://github.com/emirmrkaljevic/behave/tree/main/examples/tee) | 38 | GNU coreutils 9.12 |
| [split](https://github.com/emirmrkaljevic/behave/tree/main/examples/split) | 40 | GNU coreutils 9.12 |
| [tsort](https://github.com/emirmrkaljevic/behave/tree/main/examples/tsort) | 33 | GNU coreutils 9.12 |
| [csplit](https://github.com/emirmrkaljevic/behave/tree/main/examples/csplit) | 37 | GNU coreutils 9.12 |
| [ptx](https://github.com/emirmrkaljevic/behave/tree/main/examples/ptx) | 36 | GNU coreutils 9.12 |
| [echo](https://github.com/emirmrkaljevic/behave/tree/main/examples/echo) | 44 | GNU coreutils 9.12 |
| [readlink](https://github.com/emirmrkaljevic/behave/tree/main/examples/readlink) | 45 | GNU coreutils 9.12 |
| [env](https://github.com/emirmrkaljevic/behave/tree/main/examples/env) | 55 | GNU coreutils 9.12 |
| [dircolors](https://github.com/emirmrkaljevic/behave/tree/main/examples/dircolors) | 46 | GNU coreutils 9.12 |
| [xargs](https://github.com/emirmrkaljevic/behave/tree/main/examples/xargs) | 48 | GNU findutils 4.10 |
| [sysctl](https://github.com/emirmrkaljevic/behave/tree/main/examples/sysctl) | 31 | procps-ng 4.0.4 |
| [jq](https://github.com/emirmrkaljevic/behave/tree/main/examples/jq) | 33 | jq 1.7 (hand-written) |
| [behave](https://github.com/emirmrkaljevic/behave/tree/main/examples/behave) | 7 | behave itself |

CI runs most of these on Ubuntu with the runner's GNU tools. See [usage](usage.md#ci-notes) for the tags that are skipped there, and for the few specs that are not in the workflow yet.

[Back to behave](./) · [Source on GitHub](https://github.com/emirmrkaljevic/behave)
