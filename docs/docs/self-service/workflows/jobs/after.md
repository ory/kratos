---
id: after
title: After Jobs
---

## User Login

Jobs running after successful user authentication can be defined per
Self-Service Login Strategy in ORY Kratos' configuration file, for example:

```yaml
selfservice:
  login:
    after:
      passwordless:
        - job: redirect
          config:
            redirect_to: https://url-to-redirect/to
      oidc:
        - job: redirect
          config:
            redirect_to: https://url-to-redirect/to
      password:
        - job: redirect
          config:
            redirect_to: https://url-to-redirect/to
```

### `session`

The `session` job will send a `Set-Cookie` header which contains the session
cookie. To use it, you must first define one or more (for secret rotation)
session secrets and then use it in one of the `after` work flows:

```yaml
secrets:
  session:
    - something-super-secret # The first entry will be used to sign and verify session cookies

    # All other entries will be used to verify session cookies that were signed before "something-super-secret" became
    # the current signing secret.
    - old-session-secret
    - older-session-secret
    - ancient-session-secret

selfservice:
  registration:
    after:
      <strategy>:
        - job: session
          # can not be configured
```

> This job is required for login to work, otherwise no session will be created.

### `redirect`

The `redirect` job will send a HTTP 302 Found response and redirect the client
to the specified endpoint. There are two configuration options available:

```yaml
selfservice:
  login:
    after:
      <strategy>:
        - job: redirect
          config:
            # redirect_to sets the URL the client will be redirected to.
            redirect_to: https://url-to-redirect/to

            # allow_user_defined, if enabled, will check for a `?return_to` query parameter in the original request URL.
            # If the parameter is set and the URL is whitelisted in `urls.whitelisted_return_to_domains`
            allow_user_defined: true
```

> It is highly recommended to set up a redirect job after login, otherwise the
> user might get stuck on an empty, white screen.

### `revoke_active_sessions`

The `revoke_active_sessions` will delete all active sessions for that user on
successful login:

```yaml
selfservice:
  login:
    after:
      <strategy>:
        - job: revoke_active_sessions
          # can not be configured
```

## User Registration

Jobs running after successful user registration can be defined per Self-Service
Registration Strategy in ORY Kratos' configuration file, for example:

```yaml
selfservice:
  registration:
    after:
      passwordless:
        - job: redirect
          config:
            redirect_to: https://url-to-redirect/to
      oidc:
        - job: redirect
          config:
            redirect_to: https://url-to-redirect/to
      password:
        - job: redirect
          config:
            redirect_to: https://url-to-redirect/to
```

### `session`

The `session` job will send a `Set-Cookie` header which contains the session
cookie. To use it, you must first define one or more (for secret rotation)
session secrets and then use it in one of the `after` work flows:

```yaml
secrets:
  session:
    - something-super-secret # The first entry will be used to sign and verify session cookies

    # All other entries will be used to verify session cookies that were signed before "something-super-secret" became
    # the current signing secret.
    - old-session-secret
    - older-session-secret
    - ancient-session-secret

selfservice:
  registration:
    after:
      <strategy>:
        - job: session
          # can not be configured
```

The `session` job is useful if you want users to be signed in immediately after
registration, without further account activation or an additional login flow.

> Using this job as part of your post-registration workflow makes your system
> vulnerable to
> [Account Enumeration Attacks](../../../concepts/security.md#account-enumeration-attacks)
> because a threat agent can distinguish between existing and non-existing
> accounts by checking if `Set-Cookie` was sent as part of the registration
> response.

### `redirect`

The `redirect` job will send a HTTP 302 Found response and redirect the client
to the specified endpoint. There are two configuration options available:

```yaml
selfservice:
  registration:
    after:
      <strategy>:
        - job: redirect
          config:
            # redirect_to sets the URL the client will be redirected to.
            redirect_to: https://url-to-redirect/to

            # allow_user_defined, if enabled, will check for a `?return_to` query parameter in the original request URL.
            # If the parameter is set and the URL is whitelisted in `urls.whitelisted_return_to_domains`
            allow_user_defined: true
```

> It is highly recommended to set up a redirect job after registration,
> otherwise the user might get stuck on an empty, white screen.
