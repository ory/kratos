---
id: kratos-jsonnet-format
title: kratos jsonnet format
description: kratos jsonnet format
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos jsonnet format

### Synopsis

Formats JSONNet files using the official JSONNet formatter.

Use -w or --write to write output back to files instead of stdout.

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
kratos jsonnet format path/to/files/*.jsonnet [more/files.jsonnet, [supports/**/{foo,bar}.jsonnet]] [flags]
```

### Options

```
  -h, --help    help for format
  -w, --write   Write formatted output back to file.
```

### Options inherited from parent commands

```
  -c, --config string   Path to config file. Supports .json, .yaml, .yml, .toml. Default is "$HOME/.kratos.(yaml|yml|toml|json)"
```

### SEE ALSO

- [kratos jsonnet](kratos-jsonnet) - Helpers for linting and formatting JSONNet code
