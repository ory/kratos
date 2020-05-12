---
id: index
title: Overview
---

All SDKs use automated code generation provided by
[`openapi-generator`](https://github.com/OpenAPITools/openapi-generator).
Unfortunately, `openapi-generator` has serious breaking changes in the generated
code when upgrading versions. Therefore, we do not make backwards compatibility
promises with regards to the generated SDKs. We hope to improve this process in
the future.

Before you check out the SDKs, head over to the [REST API](api.md) documentation
which includes code samples for common programming languages for each REST
endpoint.

We publish our SDKs for popular languages in their respective package
repositories:

- [Python](https://pypi.org/project/ory-kratos-client/)
- [PHP](https://github.com/ory/kratos-client-php)
- [Go](https://github.com/ory/kratos-client-go)
- [NodeJS](https://www.npmjs.com/package/@oryd/kratos-client) (with TypeScript)
- [Java](https://search.maven.org/artifact/sh.ory.kratos/kratos-client)
- [Ruby](https://rubygems.org/gems/ory-kratos-client)

Missing your programming language?
[Create an issue](https://github.com/ory/kratos/issues) and help us build, test
and publish the SDK for your programming language!
