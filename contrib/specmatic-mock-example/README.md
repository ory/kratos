# Specmatic Mock Server Example for Ory Kratos

This directory contains examples showing how consumer applications (frontend
UIs, backend services) can use [Specmatic](https://specmatic.io/) to run an
intelligent mock of the Kratos identity server for local development and
testing.

## Why Use a Specmatic Mock Instead of Running Kratos?

Running the real Kratos server locally requires:
- PostgreSQL (or MySQL/CockroachDB)
- Database migrations
- Identity schema configuration
- Various environment variables

With Specmatic, you get a **contract-faithful mock** that:
- ✅ Responds with valid data matching the OpenAPI schema
- ✅ Rejects invalid requests (wrong paths, bad types, missing fields)
- ✅ Requires zero infrastructure (no database, no migrations)
- ✅ Starts in seconds
- ✅ Stays automatically in sync with the official spec

## Quick Start

### Option 1: Using npx (Recommended)

```bash
# From the Kratos repo root
npx --yes specmatic stub spec/api.json --port=4433
```

### Option 2: Using Docker

```bash
# From the Kratos repo root
docker run --rm -p 4433:9000 \
  -v "$(pwd)/spec/api.json:/app/spec.json" \
  specmatic/specmatic stub /app/spec.json --port=9000
```

### Option 3: Using the Specmatic MCP Server

If you have an AI coding agent with MCP support:

```
"Start a mock Kratos server on port 4433 using spec/api.json"
```

## Example: Testing a Login Flow

Once the mock is running on port 4433, you can test login flows:

```bash
# 1. Create a login flow
curl -s http://localhost:4433/self-service/login/api | jq .

# 2. Get the flow ID and submit credentials
curl -s -X POST http://localhost:4433/self-service/login \
  -H "Content-Type: application/json" \
  -d '{
    "method": "password",
    "identifier": "test@example.org",
    "password": "S3cur3P@ssw0rd!"
  }' | jq .
```

## Example: Testing with Custom Expectations

Create an `_expectations` directory next to your spec to provide custom response
examples:

```
spec/
  api.json
  api_expectations/
    login_flow.json
```

Then start the stub with expectations:

```bash
npx --yes specmatic stub spec/api.json --port=4433
```

## Example: Using in a Frontend Test Suite

### With Playwright

```javascript
// playwright.config.ts
import { defineConfig } from '@playwright/test';

export default defineConfig({
  webServer: [
    {
      // Start the Specmatic mock as the "backend"
      command: 'npx --yes specmatic stub spec/api.json --port=4433',
      port: 4433,
      reuseExistingServer: true,
    },
    {
      // Start your frontend
      command: 'npm run dev',
      port: 3000,
    },
  ],
});
```

### With Jest

```javascript
// jest.globalSetup.ts
const { exec } = require('child_process');

module.exports = async () => {
  // Start Specmatic mock before all tests
  const mockProcess = exec('npx --yes specmatic stub spec/api.json --port=4433');
  global.__SPECMATIC_MOCK__ = mockProcess;

  // Wait for the mock to be ready
  await new Promise(resolve => setTimeout(resolve, 5000));
};
```

## Example: Using in a Go Service Test

If you have a Go service that calls Kratos APIs:

```go
package myservice_test

import (
    "os/exec"
    "testing"
    "time"
    "net/http"
)

func TestMain(m *testing.M) {
    // Start Specmatic mock
    cmd := exec.Command("npx", "--yes", "specmatic", "stub",
        "spec/api.json", "--port=4433")
    cmd.Start()
    defer cmd.Process.Kill()

    // Wait for mock to be ready
    for i := 0; i < 30; i++ {
        resp, err := http.Get("http://localhost:4433/health/alive")
        if err == nil && resp.StatusCode == 200 {
            break
        }
        time.Sleep(time.Second)
    }

    // Run tests
    m.Run()
}

func TestUserRegistration(t *testing.T) {
    // Your test code here — calls go to the Specmatic mock
    // instead of a real Kratos instance
}
```

## Troubleshooting

### Mock server won't start
- Ensure Java 17+ is installed: `java --version`
- Ensure Node.js 18+ is installed: `node --version`
- Check that port 4433 is not already in use

### Mock returns unexpected responses
- The mock generates responses based on the OpenAPI schema
- Use the `specmatic_dictionary.json` file at the repo root to control test values
- Create custom expectations for specific endpoints

### Mock rejects valid requests
- Ensure your request matches the spec exactly (required headers, content type)
- Check the mock server logs for detailed error messages

## Further Reading

- [Specmatic Stub/Mock Documentation](https://specmatic.io/documentation/service_virtualization_tutorial.html)
- [Specmatic Examples](https://github.com/specmatic)
- [Ory Kratos API Reference](https://www.ory.sh/docs/kratos/reference/api)
