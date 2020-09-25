---
id: kratos-jsonnet-lint
title: kratos jsonnet lint
description: kratos jsonnet lint
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos jsonnet lint

### Synopsis

Lints JSONNet files using the official JSONNet linter and exits with a status
code of 1 when issues are detected.

Glob Syntax:

    pattern:
        { term }

    term:
        '*'         matches any sequence of non-separator characters
        '**'        matches any sequence of characters
        '?'         matches any single non-separator character
        '[' [ '!' ] { character-range } ']'
                    character class (must be non-empty)
        '{' pattern-list '}'
                    pattern alternatives
        c           matches character c (c != '*', '**', '?', '\', '[', '{', '}')
        '\' c       matches character c

    character-range:
        c           matches character c (c != '\\', '-', ']')
        '\' c       matches character c
        lo '-' hi   matches character c for lo <= c <= hi

    pattern-list:
        pattern { ',' pattern }
                    comma-separated (without spaces) patterns

```
kratos jsonnet lint path/to/files/*.jsonnet [more/files.jsonnet, [supports/**/{foo,bar}.jsonnet]] [flags]
```

### Options

```
  -h, --help   help for lint
```

### SEE ALSO

- [kratos jsonnet](kratos-jsonnet) - Helpers for linting and formatting JSONNet
  code
