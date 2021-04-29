---
id: kratos-serve
title: kratos serve
description: kratos serve Run the Ory Kratos server
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos serve

Run the Ory Kratos server

```
kratos serve [flags]
```

### Options

```
  -c, --config strings   Path to one or more .json, .yaml, .yml, .toml config files. Values are loaded in the order provided, meaning that the last config file overwrites values from the previous config file.
      --dev              Disables critical security features to make development easier
  -h, --help             help for serve
      --sqa-opt-out      Disable anonymized telemetry reports - for more information please visit https://www.ory.sh/docs/ecosystem/sqa
      --watch-courier    Run the message courier as a background task, to simplify single-instance setup
```

### SEE ALSO

- [kratos](kratos) -
