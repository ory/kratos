# Feature Implementation: Impossible Travel Detection

## 1\. Approach & Research

* Getting up to speed with the Ory Kratos codebase and architecture was the first challenge.
* To ensure my implementation followed existing patterns and conventions, I analyzed the repository for similar functionalities/PR
* I identified [PR \#2715 (feat: adding device information to the session)](https://github.com/ory/kratos/pull/2715) as a highly relevant blueprint.
* This PR established how to attach metadata to a session device.
* I used this as a foundation, specifically adapting the logic to extract geolocation data from Cloudflare headers (`CF-IPLatitude`, `CF-IPLongitude`), which aligns with the task requirements.

## 2\. Architecture & Design Decisions

The implementation is divided into two distinct components: the core calculation logic and the state management within the login flow.

### A. Core/Calculation Logic: [x/geo.go](./x/geo.go)

- I isolated the mathematical logic into a helper package [x/geo.go](./x/geo.go).
- This component calculates the distance and travel speed between two coordinates using the Haversine formula.
- By isolating this logic, I ensured it was decoupled from the Kratos session flow, making it a textbook candidate for "black-box" unit testing.
- Comprehensive [unit tests](./x/geo_test.go) cover various scenarios (zero distance, impossible speed, edge cases).

### B. Integration Strategy: The "Flagging" Mechanism

- The most critical architectural challenge was adhering to the performance constraints.
- The assessment highlighted that **DB writes are slow**, so avoiding additional synchronous write operations during the login flow was paramount.

**My Strategy:**

1.  **Flag, Don't Block:** The requirement was to "flag" the session. I assumed that enforcement (e.g., triggering MFA, revoking the session) would be handled by a downstream asynchronous worker or policy engine.
2.  **Data Model:**
  * **Device:** Added `latitude` and `longitude` columns to the `session_devices` table: this semantically belongs with devices.
  * **Session:** Added an `impossible_travel` boolean column to the `sessions` table.
3.  **Performance Optimization (Piggybacking):**
  * Instead of a separate DB write to flag the session, I integrated the check into the `PostLoginHook`.
  * The logic calculates the flag state *in-memory* and assigns it to the session object.
  * This flag is persisted during the **existing** session write operation.
  * **Result:** **Zero additional database writes** are introduced, ensuring the login latency remains unaffected.

### C. The Detection Algorithm

Inside the: [PostLoginHook](./selfservice/flow/login/hook.go)

1.  Query the database for the **latest previous device** associated with the identity (optimized query).
2.  Compare the previous device's coordinates/timestamp with the current login's coordinates (from headers).
3.  If the calculated speed exceeds the configured threshold (default: 900 km/h), set `session.ImpossibleTravel = true`.

## 3\. Verification & Testing

### Unit Testing

- The core logic in `x/geo.go` is fully unit tested
- the unit test passes without errors: (truncated for conciseness)
```bash
make test-short
go test -tags sqlite -count=1 -failfast -short ./...
?       github.com/ory/kratos   [no test files]
ok      github.com/ory/kratos/cipher    0.434s
..
..
ok      github.com/ory/kratos/x 0.109s <<[[[ the geo unit testing is included here ]]]
...
...
ok      github.com/ory/kratos/x/webauthnx/js    0.015s
```


### Manual Integration Testing

Due to time constraints, I verified the feature end-to-end using a shell script [t.sh](./t.sh) that simulates Cloudflare headers and inspects the resulting session state via the Admin API.

**Scenario:**

1.  **Login 1:** Close to previous location (False).
2.  **Login 2:** Distant coordinates immediately after (True).
3.  **Login 3:** Same Distant coordinates (False).

**Test Output:**

```bash
$ ./t.sh test@test.com Change_me_123 8.1475 11.5645
Starting Login Flow ...
Simulating Login from geo-coordinates (8.1475, 11.5645)...
Login Successful!
impossible_travel: false

$ ./t.sh test@test.com Change_me_123 108.1475 11.5645
Starting Login Flow ...
Simulating Login from geo-coordinates (108.1475, 11.5645)...
Login Successful!
impossible_travel: true  <-- CORRECTLY FLAGGED

$ ./t.sh test@test.com Change_me_123 108.1475 11.5645
Starting Login Flow ...
Simulating Login from geo-coordinates (108.1475, 11.5645)...
Login Successful!
impossible_travel: false <-- CORRECT (Same location as previous)
```

## 4\. Next Steps & Future Work

1.  **Automated Integration Tests:** The manual shell script logic should be converted into a formal Go integration test suite to ensure regression stability in CI/CD.
2.  **Performance Benchmarking:** While the design theoretically minimizes overhead, I would ideally run load tests to measure the exact latency impact of the Haversine calculation and the extra "read" query during high-throughput scenarios.
3.  **Enforcement:** Implement the downstream logic to act upon the `impossible_travel=true` flag.
