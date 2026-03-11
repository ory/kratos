
# Feasibility Report: Anonymous Sessions in Ory Kratos

## Executive summary

Anonymous sessions (sessions not tied to an authenticated identity) are **not natively supported** in Kratos today. The session model is deeply coupled to the identity model at the database, struct, and API level. Implementing anonymous sessions is feasible but requires changes across multiple layers. This report outlines the current architecture constraints, proposes API designs, and evaluates implementation approaches.

---

## 1. Current architecture analysis

### 1.1 Session–Identity coupling

The `Session` struct has a **non-nullable** `IdentityID uuid.UUID` field and a hard foreign key constraint at the database level:

```cloud/kratos/kratos-oss/persistence/sql/migrations/sql/20191100000003000000_sessions.postgres.up.sql#L1-10
CREATE TABLE "sessions" (
"id" UUID NOT NULL,
PRIMARY KEY("id"),
"issued_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
"expires_at" timestamp NOT NULL,
"authenticated_at" timestamp NOT NULL,
"identity_id" UUID NOT NULL,
"created_at" timestamp NOT NULL,
"updated_at" timestamp NOT NULL,
FOREIGN KEY ("identity_id") REFERENCES "identities" ("id") ON DELETE cascade
);
```

The persistence layer explicitly rejects sessions without an identity:

```cloud/kratos/kratos-oss/persistence/sql/persister_session.go#L244-248
	s.NID = p.NetworkID(ctx)
	if s.Identity != nil {
		s.IdentityID = s.Identity.ID
	} else if s.IdentityID.IsNil() {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("cannot upsert session without an identity or identity ID set"))
```

### 1.2 Session activation requires an identity

`ManagerHTTP.ActivateSession` hard-requires a non-nil, active identity:

```cloud/kratos/kratos-oss/session/manager_http.go#L314-324
func (s *ManagerHTTP) ActivateSession(r *http.Request, session *Session, i *identity.Identity, authenticatedAt time.Time) (err error) {
	// ...
	if i == nil {
		return errors.WithStack(x.PseudoPanic.WithReasonf("Identity must not be nil when activating a session."))
	}

	if !i.IsActive() {
		return errors.WithStack(ErrIdentityDisabled.WithDetail("identity_id", i.ID))
	}
```

### 1.3 The `/sessions/whoami` endpoint always returns identity data

The `whoami` handler unconditionally reads identity data and sets the `X-Kratos-Authenticated-Identity-Id` header:

```cloud/kratos/kratos-oss/session/handler.go#L260-261
	// Set userId as the X-Kratos-Authenticated-Identity-Id header.
	w.Header().Set("X-Kratos-Authenticated-Identity-Id", s.Identity.ID.String())
```

### 1.4 JWT tokenization uses identity as subject

The tokenizer requires the session's identity for the `sub` claim:

```cloud/kratos/kratos-oss/session/tokenizer.go#L71-82
func SetSubjectClaim(claims jwt.MapClaims, session *Session, subjectSource string) error {
	switch subjectSource {
	case "", "id":
		claims["sub"] = session.IdentityID.String()
	case "external_id":
		if session.Identity.ExternalID == "" {
			return errors.WithStack(herodot.ErrBadRequest.WithReasonf("The session's identity does not have an external ID set, but it is required for the subject claim."))
		}
		claims["sub"] = session.Identity.ExternalID.String()
```

### 1.5 Hooks assume identity existence

Multiple hooks (e.g., `SessionDestroyer`, `AddressVerifier`, `SessionIssuer`) dereference `s.Identity.ID` without nil checks:

```cloud/kratos/kratos-oss/selfservice/hook/session_destroyer.go#L37-44
func (e *SessionDestroyer) ExecuteLoginPostHook(_ http.ResponseWriter, r *http.Request, _ node.UiNodeGroup, _ *login.Flow, s *session.Session) error {
	return otelx.WithSpan(r.Context(), "selfservice.hook.SessionDestroyer.ExecuteLoginPostHook", func(ctx context.Context) error {
		if _, err := e.r.SessionPersister().RevokeSessionsIdentityExcept(ctx, s.Identity.ID, s.ID); err != nil {
			return err
		}
		return nil
	})
}
```

### 1.6 There is zero existing concept of "anonymous" or "guest" in the codebase

A codebase-wide search for `anonymous`, `guest`, `ephemeral` in the context of sessions returned **no relevant results**. This is a greenfield feature.

---

## 2. Proposed API design

### 2.1 New endpoint: create anonymous session

```/dev/null/api.yaml#L1-30
# POST /sessions/anonymous
# Creates an anonymous session without requiring authentication.

# Request (Browser flow):
#   No body required. Sets session cookie automatically.

# Request (API flow):
#   No body required.

# Response 200:
{
  "session": {
    "id": "uuid",
    "active": true,
    "expires_at": "2024-01-01T00:00:00Z",
    "issued_at": "2024-01-01T00:00:00Z",
    "authenticated_at": "2024-01-01T00:00:00Z",
    "authenticator_assurance_level": "aal0",
    "authentication_methods": [
      { "method": "anonymous", "aal": "aal0", "completed_at": "..." }
    ],
    "identity": null,
    "anonymous": true,
    "devices": [...]
  },
  "session_token": "ory_st_..." // Only for API flows
}
```

### 2.2 Modified `/sessions/whoami` behavior

The whoami endpoint should gracefully handle anonymous sessions:

```/dev/null/api.yaml#L1-22
# GET /sessions/whoami
# Returns the current session. For anonymous sessions, identity is null.

# Response 200 (anonymous session):
{
  "id": "uuid",
  "active": true,
  "anonymous": true,
  "authenticator_assurance_level": "aal0",
  "authentication_methods": [
    { "method": "anonymous", "aal": "aal0" }
  ],
  "identity": null,
  "devices": [...]
}

# Response 200 (authenticated session):
# Same as today, with "anonymous": false
```

### 2.3 Session promotion: anonymous → authenticated

When a user logs in or registers while holding an anonymous session, the session should be promotable:

```/dev/null/api.yaml#L1-15
# POST /self-service/login?flow=<flow_id>
# If the user has an active anonymous session cookie/token,
# the login flow promotes the anonymous session to an authenticated one.

# Behavior:
# 1. Anonymous session is revoked
# 2. New authenticated session is created
# 3. The anonymous session ID is available in the login hook context
#    so that application logic can migrate anonymous data (e.g., cart)

# New hook context field:
#   "previous_anonymous_session_id": "uuid"  (available in post-login webhooks)
```

### 2.4 Configuration

```/dev/null/config.yaml#L1-14
session:
  anonymous:
    # Enable anonymous session creation
    enabled: false
    # Lifespan of anonymous sessions (shorter than authenticated by default)
    lifespan: 1h
    # Maximum number of anonymous sessions per IP (rate limiting)
    max_per_ip: 100
    # Cookie name for anonymous sessions (separate from authenticated sessions)
    cookie:
      name: ory_kratos_anonymous_session
```

---

## 3. Implementation approaches

### Approach A: Phantom identity (recommended)

Create a lightweight "anonymous" identity behind the scenes for each anonymous session. This is the **lowest-risk** option.

| Aspect | Detail |
|---|---|
| **Core idea** | When an anonymous session is requested, create a special `Identity` with `state: active`, a dedicated `schema_id: "anonymous"`, and empty traits. The session's `IdentityID` FK is satisfied. |
| **Session struct change** | Add `Anonymous bool` field to `Session` (new DB column `is_anonymous`). |
| **Existing code impact** | Minimal. All existing code that reads `IdentityID` or `Identity` continues to work. `whoami` can check `s.Anonymous` and null out the identity in the response. |
| **Migration** | One new column: `ALTER TABLE sessions ADD COLUMN is_anonymous BOOL NOT NULL DEFAULT false`. |
| **Promotion** | On login/registration, update the anonymous session's `IdentityID` to the real identity and set `is_anonymous = false`. Or revoke and create new. |
| **Cleanup** | Expired anonymous sessions are cleaned up by existing `DeleteExpiredSessions`. The phantom identities can be garbage-collected when their sessions expire. |
| **Drawbacks** | Creates identity records that aren't "real" users. Inflates identity counts. Needs logic to exclude anonymous identities from list/count endpoints. |

### Approach B: Nullable IdentityID

Make `IdentityID` nullable across the entire stack.

| Aspect | Detail |
|---|---|
| **Core idea** | Change `IdentityID uuid.UUID` → `IdentityID uuid.NullUUID` in the `Session` struct. Change DB column to nullable. |
| **Blast radius** | **Very large.** Every code path that references `IdentityID` or `Identity` must handle nil: `UpsertSession`, `ActivateSession`, `GetSessionByToken`, `DoesSessionSatisfy`, `SetSessionDeviceInformation`, `Tokenizer`, all hooks, all self-service flows, OpenAPI spec, generated clients. |
| **Migration** | `ALTER TABLE sessions ALTER COLUMN identity_id DROP NOT NULL; ALTER TABLE sessions DROP CONSTRAINT sessions_identity_id_fkey; ADD CONSTRAINT ... ON DELETE SET NULL`. |
| **Drawbacks** | High risk of nil-pointer panics. Breaks the invariant that every session has an owner. Difficult to validate completeness of nil-handling. |

### Approach C: Separate anonymous session table and handler

Create a distinct `anonymous_sessions` table with its own handler.

| Aspect | Detail |
|---|---|
| **Core idea** | `anonymous_sessions` table with `id`, `token`, `expires_at`, `metadata`, `devices`. Completely separate from authenticated sessions. |
| **Blast radius** | Low on existing code. New code is isolated. |
| **Migration** | New table only, no changes to existing schema. |
| **Drawbacks** | Duplicates session management logic (cookie issuance, token handling, expiry, etc.). Two parallel systems to maintain. `whoami` must check both tables. Session promotion requires cross-table coordination. |

---

## 4. Impact matrix

| Component | Approach A (Phantom) | Approach B (Nullable) | Approach C (Separate) |
|---|---|---|---|
| `session.Session` struct | +1 field | Change `IdentityID` type | No change |
| `session.Persister` | Minor guard | Major refactor | New interface |
| `session.ManagerHTTP` | New method + guards | Refactor `ActivateSession`, `FetchFromRequest`, `DoesSessionSatisfy` | New manager |
| `session.Handler` | New route + `whoami` guard | Guards in every handler | New handler |
| `session.Tokenizer` | Guard for anonymous | Guard for nullable identity | Separate tokenizer logic |
| DB migration | 1 column | ALTER + FK change | New table |
| Self-service flows | Hook context extension | Nil-handling everywhere | Isolated |
| Hooks | Nil-guard in ~5 hooks | Nil-guard in ~5 hooks | N/A |
| OpenAPI spec | New endpoint + field | Modified `session` schema | New endpoints + schema |
| Identity handler/pool | Exclude anonymous from counts | No change | No change |
| Risk | **Low-Medium** | **High** | **Low** |
| Effort | **Medium** (~2-3 weeks) | **High** (~4-6 weeks) | **Medium** (~2-3 weeks) |

---

## 5. Recommendation

**Approach A (Phantom Identity)** is recommended. It satisfies the FK constraint naturally, minimizes blast radius on existing code, and leverages all existing session infrastructure (cookies, tokens, expiry, cleanup, caching). The main trade-off—phantom identities inflating counts—is manageable by filtering on the `is_anonymous` column or a dedicated `schema_id`.

### Key implementation steps for Approach A:

1. **Add `Anonymous` field** to `Session` struct + DB migration.
2. **Add new `CredentialsType`**: `CredentialsTypeAnonymous = "anonymous"` for the AMR.
3. **Add `POST /sessions/anonymous` endpoint** in `session.Handler` that:
   - Creates a phantom identity with `schema_id: "anonymous"` and empty traits.
   - Creates and activates a session with `Anonymous: true`, `AAL: aal0`.
   - Issues cookie or returns token.
4. **Guard `whoami`**: If `s.Anonymous`, null out identity in response and skip the `X-Kratos-Authenticated-Identity-Id` header.
5. **Guard hooks**: Add nil/anonymous checks in `SessionDestroyer` and other hooks.
6. **Session promotion**: In the login/registration post-hook, detect if an anonymous session exists, revoke it, and pass the old session ID to webhooks via `transient_payload` or a new hook context field.
7. **Configuration**: Add `session.anonymous.enabled` and `session.anonymous.lifespan`.
8. **Identity list filtering**: Exclude anonymous identities from `/admin/identities` by default (or add a filter parameter).
9. **Cleanup job**: Extend `DeleteExpiredSessions` to also garbage-collect orphaned phantom identities whose sessions are all expired/revoked.

---

## 6. Open questions

1. **Should anonymous sessions share the same cookie name?** Using a separate cookie avoids interference but complicates promotion. Using the same cookie makes promotion seamless but means authenticated sessions overwrite anonymous ones.
2. **Should anonymous sessions be tokenizable (JWT)?** The `sub` claim has no meaningful identity. A session-ID-only JWT could work, but consumers expecting an identity subject would break.
3. **Rate limiting**: Without authentication, anonymous session creation is a DoS vector. Per-IP rate limiting and short lifespans are essential.
4. **Multi-tenancy (NID)**: Anonymous sessions should respect network isolation like authenticated sessions. No additional work needed since phantom identities inherit the NID.
5. **Hydra/OAuth2 integration**: Anonymous sessions should likely **not** be usable as OAuth2 login sessions. The `AcceptLoginRequest` flow requires an authenticated identity.
