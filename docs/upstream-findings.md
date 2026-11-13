---
title: "Upstream findings: bugs behave found in CLI rewrites"
description: Bug reports and fix PRs found by recording GNU tools with behave and replaying the cases against Rust rewrites.
---

# Upstream findings

Running behave against Rust rewrites of GNU tools has led to 41 bug reports and 48 fix PRs, 33 of them merged, as of 13 Nov 2026. Each issue and PR notes that it was found with behave, and each fix includes a test.

| project | reports | fix PRs | merged |
|---|---:|---:|---:|
| [uutils/coreutils](https://github.com/uutils/coreutils) | 34 | 41 | 29 |
| [uutils/findutils](https://github.com/uutils/findutils) | 3 | 3 | 1 |
| [uutils/sed](https://github.com/uutils/sed) | 2 | 2 | 1 |
| [uutils/procps](https://github.com/uutils/procps) | 2 | 2 | 2 |
| total | 41 | 48 | 33 |

The extra coreutils PRs are follow-up fixes in other utilities for the same class of bug, for example numeric-looking file names after [tail #15533](https://github.com/uutils/coreutils/issues/15533).

Many recorded differences are deliberate design choices (error wording, clap vs GNU usage text), not bugs. The lists below are the ones that were reported.

## Early comparison snapshot

These were generated-case runs on one machine, with `--help`/`--version` excluded. Numbers are a snapshot from early use, not a live scoreboard.

| original | candidate | cases | pass | notable differences |
|---|---|---|---|---|
| GNU wc 9.7 | uutils wc 0.12.0 | 63 | 63 | none |
| GNU head | uutils head | 51 | 51 | none |
| GNU uniq | uutils uniq | 100 | 100 | none |
| GNU tr | uutils tr | 63 | 63 | none |
| GNU tac 9.12 | uutils tac (main) | 36 | 35 | `-s -1` rejected (value starting with `-`, [#15497](https://github.com/uutils/coreutils/issues/15497), now fixed) |
| GNU tail 9.12 | uutils tail (main) | 60 | 58 | `-c` with no count accepted ([#15528](https://github.com/uutils/coreutils/issues/15528), now fixed); file names like `007` printed as `7` ([#15533](https://github.com/uutils/coreutils/issues/15533), fixed in tail and several other tools); a hand-added `-cf FILE` case is rejected ([#15517](https://github.com/uutils/coreutils/issues/15517), now fixed) |
| GNU cut | uutils cut | 175 | 172 | `--output-delimiter -1` rejected (value starting with `-`) |
| GNU base64 | uutils base64 | 36 | 31 | `-d` on invalid input: GNU prints the bytes decoded so far, uutils prints nothing |
| GNU sort | uutils sort | 290 | ~260 | usage errors exit 1 instead of 2; `--debug` underline width; `--check --merge` |
| jq 1.7 | gojq 0.12.19 | 596 | 331 | unsupported flags, key order, big-number formatting; pass count is with `--ignore-stderr` |

Record from the release the rewrite actually targets. uutils follows the newest GNU release, and GNU 9.12 prints `==> 'standard input' <==` in `tail -v` headers where 9.7 prints no quotes, so a spec recorded from an older GNU reports differences that are really GNU changes. `behave.yaml` keeps the `--version` line of the recorded tool, and `run`/`diff` print it next to the version being replayed.

## uutils coreutils

Fixed:

- `cut --output-delimiter -x` was rejected ([#15341](https://github.com/uutils/coreutils/issues/15341), fixed in [#15345](https://github.com/uutils/coreutils/pull/15345))
- `sort --check --merge` ignored the check ([#15380](https://github.com/uutils/coreutils/issues/15380), fixed in [#15381](https://github.com/uutils/coreutils/pull/15381))
- `tac -s -1` rejected a separator starting with `-` ([#15497](https://github.com/uutils/coreutils/issues/15497), fixed in [#15498](https://github.com/uutils/coreutils/pull/15498))
- `tail 007` printed the missing file as `'7'` ([#15533](https://github.com/uutils/coreutils/issues/15533), fixed for tail in [#15534](https://github.com/uutils/coreutils/pull/15534); the same bug in other utilities was fixed one by one: `tac` in [#15541](https://github.com/uutils/coreutils/pull/15541), `nl` in [#15544](https://github.com/uutils/coreutils/pull/15544), `md5sum`/`cksum` in [#15548](https://github.com/uutils/coreutils/pull/15548), `sort` in [#15565](https://github.com/uutils/coreutils/pull/15565), `shred` in [#15604](https://github.com/uutils/coreutils/pull/15604)). The issue is still open for any remaining utilities.
- bare `tail -c` and `tail +NUM` ignored `_POSIX2_VERSION` ([#15528](https://github.com/uutils/coreutils/issues/15528), fixed in [#15529](https://github.com/uutils/coreutils/pull/15529)); `tail -cf` obsolete syntax ([#15517](https://github.com/uutils/coreutils/issues/15517), fixed in [#15518](https://github.com/uutils/coreutils/pull/15518))
- `sort -c dir` exited 0 without a message ([#15615](https://github.com/uutils/coreutils/issues/15615), fixed in [#15616](https://github.com/uutils/coreutils/pull/15616))
- `paste` open and read errors omitted the file name ([#15254](https://github.com/uutils/coreutils/issues/15254), fixed in [#15642](https://github.com/uutils/coreutils/pull/15642))
- `fmt d a.txt` stopped at the directory with a bare `read error` and never formatted `a.txt` ([#15922](https://github.com/uutils/coreutils/issues/15922), fixed in [#15923](https://github.com/uutils/coreutils/pull/15923))
- `od -t x0` (and `d0`, `u0`, `f0`) silently used the default size, where GNU rejects a 0-byte type ([#16085](https://github.com/uutils/coreutils/issues/16085), fixed in [#16086](https://github.com/uutils/coreutils/pull/16086))
- `od -A q` and `od -t x3` error wording ([#16071](https://github.com/uutils/coreutils/issues/16071), fixed in [#16072](https://github.com/uutils/coreutils/pull/16072))
- `cp f/ dest` printed the raw OS error without the file name ([#15691](https://github.com/uutils/coreutils/issues/15691), fixed in [#15692](https://github.com/uutils/coreutils/pull/15692))
- `ln f/ link` and `ln missing link` said `failed to create hard link` instead of `failed to access` ([#15713](https://github.com/uutils/coreutils/issues/15713), fixed in [#15714](https://github.com/uutils/coreutils/pull/15714))
- `mv a missing/b` and `mv a ro/b` printed only the bare OS error ([#15721](https://github.com/uutils/coreutils/issues/15721), fixed in [#15722](https://github.com/uutils/coreutils/pull/15722))
- `mv -T f d` omitted the source name ([#15724](https://github.com/uutils/coreutils/issues/15724), fixed in [#15725](https://github.com/uutils/coreutils/pull/15725))
- `mv -n z z` (and `--update=none`) failed with `are the same file`, where GNU skips the move and exits 0 ([#15824](https://github.com/uutils/coreutils/issues/15824), fixed in [#15825](https://github.com/uutils/coreutils/pull/15825))
- `chown 1000 f/` and `chgrp 1000 missing` said `cannot dereference` instead of `cannot access` ([#15867](https://github.com/uutils/coreutils/issues/15867), fixed in [#15868](https://github.com/uutils/coreutils/pull/15868))
- `chmod -v 644 missing` printed nothing, or a made-up mode line for a dangling symlink ([#15883](https://github.com/uutils/coreutils/issues/15883), fixed in [#15884](https://github.com/uutils/coreutils/pull/15884))
- `seq --foo` said `invalid floating point argument` and `seq -inf 1` printed `-inf` forever ([#16113](https://github.com/uutils/coreutils/issues/16113), fixed in [#16114](https://github.com/uutils/coreutils/pull/16114))
- `factor '30 '` and `factor $'\t30'` printed factors where GNU rejects the argument ([#16142](https://github.com/uutils/coreutils/issues/16142), fixed in [#16143](https://github.com/uutils/coreutils/pull/16143))
- `md5sum -c` rejected lines it writes for names with a backslash or a newline, and unescaped names without that marker ([#16166](https://github.com/uutils/coreutils/issues/16166), fixed in [#16167](https://github.com/uutils/coreutils/pull/16167))
- `base64 -w 99999999999999999999` failed with `invalid wrap size` where GNU turns wrapping off ([#16180](https://github.com/uutils/coreutils/issues/16180), fixed in [#16181](https://github.com/uutils/coreutils/pull/16181))
- `fmt -w 20` could print a line longer than 20 columns ([#16190](https://github.com/uutils/coreutils/issues/16190), fixed in [#16191](https://github.com/uutils/coreutils/pull/16191))
- `ptx` input and output file errors quoted names that GNU leaves bare ([#16362](https://github.com/uutils/coreutils/issues/16362), fixed in [#16363](https://github.com/uutils/coreutils/pull/16363))

Open (each has a fix PR):

- `shred dir` says `Not a file` instead of `failed to open for writing: Is a directory` ([#15630](https://github.com/uutils/coreutils/issues/15630), [#15631](https://github.com/uutils/coreutils/pull/15631))
- `mv missing/ g` and `mv d/f g` with an unsearchable `d` give the wrong reason ([#15704](https://github.com/uutils/coreutils/issues/15704), [#15705](https://github.com/uutils/coreutils/pull/15705))
- `tr a '[:alpha:]'` words the set2 class error differently from GNU ([#15948](https://github.com/uutils/coreutils/issues/15948), [#15949](https://github.com/uutils/coreutils/pull/15949))
- `uniq -f x` (and `-s`, `-w`) words the invalid number error differently ([#15970](https://github.com/uutils/coreutils/issues/15970), [#15971](https://github.com/uutils/coreutils/pull/15971))
- `sort -t ab`, `sort -t ''` and `sort -c a b` error wording ([#15975](https://github.com/uutils/coreutils/issues/15975), [#15976](https://github.com/uutils/coreutils/pull/15976))
- `realpath '' /usr` aborted with a garbled error and never printed `/usr` ([#16206](https://github.com/uutils/coreutils/issues/16206), [#16207](https://github.com/uutils/coreutils/pull/16207))
- `pr -o` indented body lines but not the page header ([#16211](https://github.com/uutils/coreutils/issues/16211), [#16212](https://github.com/uutils/coreutils/pull/16212))
- `base32 -d` (and `base64`/`basenc`) accepts a carriage return in encoded input where GNU rejects it unless `-i` ([#16223](https://github.com/uutils/coreutils/issues/16223), [#16224](https://github.com/uutils/coreutils/pull/16224))
- `csplit --suffix-format` errors differ from GNU, and `%x` with a quote flag is accepted ([#16350](https://github.com/uutils/coreutils/issues/16350), [#16351](https://github.com/uutils/coreutils/pull/16351))
- `readlink -f/-e/-m ''` resolves to the current directory and exits 0 ([#16418](https://github.com/uutils/coreutils/issues/16418), [#16419](https://github.com/uutils/coreutils/pull/16419))
- `env` suggests `-S` for shebang lines after every missing program or option error, where GNU only does when the name contains whitespace ([#16423](https://github.com/uutils/coreutils/issues/16423), [#16424](https://github.com/uutils/coreutils/pull/16424))
- `env -C` appends Rust's `(os error 2)` to the error when the directory can't be entered ([#16432](https://github.com/uutils/coreutils/issues/16432), [#16433](https://github.com/uutils/coreutils/pull/16433))

## uutils findutils (`xargs`, GNU findutils 4.10)

- `xargs -I`/`-i`/`--replace` with the default command printed the input items, where GNU prints a blank line per item ([#984](https://github.com/uutils/findutils/issues/984), fixed in [#986](https://github.com/uutils/findutils/pull/986))
- `xargs -a -foo` (and `-n`/`-s`/`-L`/`-d`/`-P` with a value starting with a dash) rejected the hyphen-leading value ([#985](https://github.com/uutils/findutils/issues/985), fix in [#987](https://github.com/uutils/findutils/pull/987))
- `xargs --eof` swallowed the next option as the EOF string ([#992](https://github.com/uutils/findutils/issues/992), fix in [#993](https://github.com/uutils/findutils/pull/993))

## uutils sed

- `v` panicked at runtime when the version check succeeded ([#669](https://github.com/uutils/sed/issues/669), fix in [#671](https://github.com/uutils/sed/pull/671))
- `#` after a command was rejected instead of starting a comment, and `}` could not end a flags list ([#670](https://github.com/uutils/sed/issues/670), fixed in [#672](https://github.com/uutils/sed/pull/672))

## uutils procps

- `sysctl -o` and `sysctl -x` swallowed the next key; those are BSD no-op flags ([#835](https://github.com/uutils/procps/issues/835), fixed in [#837](https://github.com/uutils/procps/pull/837))
- `pmap -XX` was rejected as a repeated `-X`, where GNU uses `-XX` for the most detailed maps format ([#836](https://github.com/uutils/procps/issues/836), fixed in [#838](https://github.com/uutils/procps/pull/838))

[Back to behave](./) · [Source on GitHub](https://github.com/emirmrkaljevic/behave)
