# Specmatic Contract Testing for Ory Kratos

This guide explains how [Specmatic](https://specmatic.io/) is integrated into
Ory Kratos to provide automated API contract testing, backward compatibility
enforcement, and intelligent service virtualization.

## Overview

Specmatic transforms the Kratos OpenAPI specification (`spec/api.json`) into an
**executable contract** that is used in three ways:

1. **Backward Compatibility Checks** — Every pull request is automatically
   checked for breaking API changes. If a PR modifies the OpenAPI spec in a way
   that would break existing consumers (removing fields, changing types, removing
   endpoints), the CI pipeline blocks the merge.

2. **Contract Testing (Provider-Side)** — Specmatic auto-generates test requests
   from the OpenAPI spec and sends them to a running Kratos instance. It
   validates that every response matches the documented schema, catching drift
   between the implementation and the specification.

3. **Service Virtualization (Consumer-Side)** — Specmatic generates an
   intelligent mock server from the spec. Frontend and backend applications that
   integrate with Kratos can test against this mock without running the actual
   Kratos server, PostgreSQL, or running migrations.

## Quick Start

### Prerequisites

- [Node.js](https://nodejs.org/) v18+ (for npx)
- [Java Runtime](https://adoptium.net/) v17+ (Specmatic runs on the JVM)
- [Docker](https://www.docker.com/) (optional, for containerized testing)

### Run Backward Compatibility Check

Compare your local spec changes against the `master` branch:

```bash
npx --yes specmatic backward-compatibility-check \
  --target-path=spec/api.json \
  --base-branch=origin/master
```

This command:
- Parses the current spec and the version on `master`
- Identifies any breaking changes (removed fields, type changes, removed endpoints)
- Exits with a non-zero code if breaking changes are detected

### Run Contract Tests Locally

Start Kratos locally (see the main README for setup), then:

```bash
# Run contract tests against a running Kratos instance
npx --yes specmatic test \
  --spec-file=spec/api.json \
  --testBaseURL=http://localhost:4433
```

### Start a Mock Kratos Server

For consumer-side testing, spin up a Specmatic mock:

```bash
# Start a mock Kratos server on port 4433
npx --yes specmatic stub spec/api.json --port=4433
```

Or using Docker:

```bash
docker run --rm -p 4433:9000 \
  -v "$(pwd)/spec/api.json:/app/spec.json" \
  specmatic/specmatic stub /app/spec.json --port=9000
```

The mock server:
- Responds with valid data matching the OpenAPI schema
- Rejects requests that don't match the spec (invalid paths, wrong types)
- Requires no database, migrations, or configuration

## CI/CD Integration

Specmatic is integrated into the CI pipeline via
`.github/workflows/specmatic-contract-tests.yml`. The workflow runs three jobs:

### 1. Backward Compatibility Check (on every PR)

Triggered on pull requests that touch `spec/`, `.schema/`, or Go source files.
Compares the PR's spec against `master` and blocks the merge if breaking changes
are found.

### 2. OpenAPI Spec Validation (on every push/PR)

Validates that `spec/api.json` is a well-formed OpenAPI 3.0 spec and can be used
by Specmatic for contract testing and mock generation. Includes a smoke test that
starts a mock server and hits the `/health/alive` endpoint.

### 3. Contract Test (on master pushes)

Builds Kratos, starts it with a PostgreSQL database, and runs full contract tests
to verify the implementation matches the spec. This only runs on pushes to
`master` to avoid heavy CI costs on every PR.

## Configuration Files

### `specmatic.yaml`

The central configuration file at the repo root. It tells Specmatic where to find
the OpenAPI spec and sets defaults for testing and stubbing.

### `specmatic_dictionary.json`

A domain-specific dictionary that provides realistic test values for
identity-related fields (emails, passwords, UUIDs, etc.). Specmatic uses these
values when auto-generating test requests instead of random data.

## How It Works

### Why Contract Testing for Kratos?

Kratos is a **security-critical identity server**. Its API contract is consumed by:
- Frontend applications for login/registration flows
- Backend services for identity management
- SDKs generated from the OpenAPI spec (Go, JavaScript, Python, Java, etc.)

A breaking API change in Kratos is not just a bug — it's a potential security
incident. If a session response schema changes silently, downstream services may
fail to properly authenticate users.

### What Specmatic Catches

| Issue | Traditional Testing | Specmatic |
|-------|-------------------|-----------|
| Removed API field | ❌ May pass unit tests | ✅ Caught by backward compatibility |
| Type change (string → number) | ❌ May pass if not tested | ✅ Caught by contract test |
| New required field | ❌ May pass provider tests | ✅ Caught by backward compatibility |
| Missing error response | ⚠️ Only if explicitly tested | ✅ Auto-generated negative tests |
| Spec-implementation drift | ❌ No automated check | ✅ Contract test validates every endpoint |

## Using the Specmatic MCP Server

For AI-assisted development, you can use the
[Specmatic MCP Server](https://github.com/specmatic/specmatic-mcp-server) to
expose contract testing capabilities to AI coding agents:

```bash
# Install globally
npm install -g specmatic-mcp

# Or add to your MCP client configuration:
# {
#   "servers": {
#     "specmatic": {
#       "command": "npx",
#       "args": ["specmatic-mcp"]
#     }
#   }
# }
```

This enables natural language commands like:
- "Run contract tests against my local Kratos instance"
- "Start a mock Kratos server on port 4433"
- "Check if my spec changes break backward compatibility"

## Further Reading

- [Specmatic Documentation](https://specmatic.io/documentation)
- [Specmatic GitHub](https://github.com/specmatic/specmatic)
- [Contract-Driven Development](https://specmatic.io/contract_driven_development)
- [Ory Kratos API Reference](https://www.ory.sh/docs/kratos/reference/api)
