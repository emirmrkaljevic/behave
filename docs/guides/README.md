---
title: Guides
description: Walk-throughs for testing a rewrite, generating golden tests, and the uutils case study.
permalink: /guides/
redirect_from:
  - /guides/README.html
---

# Guides

Walk-throughs for [behave](https://github.com/emirmrkaljevic/behave/blob/main/README.md): how to test a rewrite, how to generate golden tests, and a case study of what it found in uutils coreutils.

## [Test a Rust or Go rewrite of a CLI against the original](rewrite-compatibility.md)

Record the original tool and diff a Rust or Go port against that recording.

## [Golden tests for CLI tools, generated from --help](golden-tests-cli.md)

Generate snapshot tests for stdout, stderr and exit codes from the tool's own help text.

## [Case study: uutils coreutils vs GNU coreutils](uutils-case-study.md)

What behave found when it recorded GNU coreutils 9.12 and replayed the cases against the Rust rewrite.

[Back to behave](../) · [Source on GitHub](https://github.com/emirmrkaljevic/behave)
