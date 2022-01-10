---
id: upgrade
title: Ory Kratos Upgrade Guide
---

This guide covers essentials to consider when upgrading Ory Kratos, please also
refer to [UPGRADE.md](https://github.com/ory/kratos/blob/master/UPGRADE.md).

Generally after upgrading just run `kratos migrate`.

Please note the following breaking changes when upgrading from Ory Kratos 0.5 to
0.8:

## Ory Kratos v0.6 Breaking changes

Full list of changes:
https://github.com/ory/kratos/blob/v0.6.0-alpha.1/CHANGELOG.md#breaking-changes

1. **hashing:** BCrypt is now the default hashing algorithm. If you wish to
   continue using Argon2id please set `hashers.algorithm` to `argon2`.

1. To upgrade to v0.6 you must apply SQL migrations. These migrations will drop
   the flow method table implying that all self-service flows that are ongoing
   will become invalid. We recommend purging the flow table manually as well
   after this migration has been applied, if you have users doing at least one
   self-service flow per minute. This implies a significant breaking change in
   the self-service flows payload. Please consult the new ui documentation. In
   essence, the login flow's `methods` key was replaced with a generic `ui` key
   which provides information for the UI that needs to be rendered.

1. This change introduces a new feature: UI Nodes. Previously, all self-service
   flows (login, registration, ...) included form fields (e.g.
   `methods.password.config.fields`). However, these form fields lacked support
   for other types of UI elements such as links (for e.g. "Sign in with
   Google"), images (e.g. QR codes), javascript (e.g. WebAuthn), or text (e.g.
   recovery codes). With v0.6, these new features have been introduced. Please
   be aware that this introduces significant breaking changes which you will
   need to adopt to in your UI. Please refer to the most recent documentation to
   see what has changed. Conceptionally, most things stayed the same - you do
   however need to update how you access and render the form fields.

1. Please be also aware that v0.6 includes SQL migrations which **purge existing
   self-service forms** from the database. This means that users will need to
   re-start the login/registration/... flow after the SQL migrations have been
   applied! If you wish to keep these records, make a back up of your database
   prior!

1. The configuration value for `hashers.argon2.memory` is now a string
   representation of the memory amount including the unit of measurement. To
   convert the value divide your current setting (KB) by 1024 to get a result in
   MB or 1048576 to get a result in GB. Example: `131072` would now become
   `128MB`.

1. The following configuration keys were updated:

   ```diff
   selfservice.methods.password.config.max_breaches
   ```

   - `password.max_breaches` ->
     `selfservice.methods.password.config.max_breaches`
   - `password.ignore_network_errors` ->
     `selfservice.methods.password.config.ignore_network_errors`

1. After battling with [spf13/viper](https://github.com/spf13/viper) for several
   years we finally found a viable alternative with
   [knadh/koanf](https://github.com/knadh/koanf). The complete internal
   configuration infrastructure has changed, with several highlights:

   1. Configuration sourcing works from all sources (file, env, cli flags) with
      validation against the configuration schema, greatly improving developer
      experience when changing or updating configuration.
   2. Configuration reloading has improved significantly and works flawlessly on
      Kubernetes.
   3. Performance increased dramatically, completely removing the need for a
      cache layer between the configuration system and ORY Hydra.
   4. It is now possible to load several config files using the `--config` flag.
   5. Configuration values are now sent to the tracer (e.g. Jaeger) if tracing
      is enabled.

Please be aware that Ory Kratos might complain about an invalid configuration,
because the validation process has improved significantly.

## Ory Kratos v0.7 Breaking changes

Full list of changes:
https://github.com/ory/kratos/blob/v0.7.0-alpha.1/CHANGELOG.md#breaking-changes

1.  Prior to this change it was not possible to specify the
    verification/recovery link lifetime. Instead, it was bound to the flow
    expiry. This patch changes that and adds the ability to configure the
    lifespan of the link individually:

    ```diff
     selfservice:
       methods:
         link:
           enabled: true
           config:
    +        # Defines how long a recovery link is valid for (default 1h)
    +        lifespan: 15m
    ```

    This is a breaking change because the link strategy no longer respects the
    recovery / verification flow expiry time and, unless set, will default to
    one hour.

1.  This change introduces a better SDK. As part of this change, several
    breaking changes with regards to the SDK have been introduced. We recommend
    reading this section carefully to understand the changes and how they might
    affect you. Before, the SDK was structured into tags `public` and `admin`.
    This stems from the fact that we have two ports in Ory Kratos - one
    administrative and one public port. While serves as a good overview when
    working with Ory Kratos, it does not express:

        - What module the API belongs to (e.g. self-service, identity, ...)
        - What maturity the API has (e.g. experimental, alpha, beta, ...)
        - What version the API has (e.g. v0alpha0, v1beta0, ...)

        This patch replaces the current `admin` and `public` tags with a versioned approach indicating the maturity of the API used. For example, `initializeSelfServiceSettingsForBrowsers` would no longer be under the `public` tag but instead under the `v0alpha1` tag:

        ```diff
        import {
          Configuration,
        - PublicApi
        + V0Alpha1
        } from '@ory/kratos-client';

        - const kratos = new PublicApi(new Configuration({ basePath: config.kratos.public }));
        + const kratos = new V0Alpha1(new Configuration({ basePath: config.kratos.public }));
        ```

        To avoid confusion when setting up the SDK, and potentially using the wrong endpoints in your codebase and ending up with strange 404 errors, Ory Kratos now redirects you to the correct port, given that `serve.(public|admin).base_url` are configured correctly.

1.  Further, all administrative functions require authorization using e.g. an
    Ory Personal Access Token in Ory Cloud. For self-hosted deployments of Ory
    Kratos, we do not know what developers use to protect their APIs. As such,
    we believe that it is ok to have admin and public functions under one common
    API and differentiate with an `admin` prefix. Therefore, the following
    patches should be made in your codebase:

    ```diff
    import {
    - AdminApi,
    + V0Alpha1,
      Configuration
    } from '@ory/kratos-client';

    -const kratos = new AdminApi(new Configuration({ basePath: config.kratos.admin }));
    +const kratos = new V0Alpha1(new Configuration({ basePath: config.kratos.admin }));

    -kratos.createIdentity({
    +kratos.adminCreateIdentity({
      schema_id: 'default',
      traits: { /* ... */ }
    })
    ```

1.  We streamlined how credentials are used. We now differentiate between:

    - Per-request credentials such as the Ory Session Token / Cookie
      ```diff
      - public getSelfServiceRegistrationFlow(id: string, cookie?: string, options?: any) {}
      + public getSelfServiceSettingsFlow(id: string, xSessionToken?: string, cookie?: string, options?: any) {}
      ```
    - Global credentials such as the Ory Cloud PAT.

      ```typescript
      const kratos = new V0Alpha0(
        new Configuration({
          basePath: config.kratos.admin,
          accessToken: 'some-token'
        })
      )

      kratosAdmin.adminCreateIdentity({
        schema_id: 'default',
        traits: {
          /* ... */
        }
      })
      ```

1.  This patch introduces CSRF countermeasures for fetching all self-service
    flows. This ensures that users can not accidentally leak sensitive
    information when copy/pasting e.g. login URLs (see #1282). If a self-service
    flow for browsers is requested, the CSRF cookie must be included in the
    call, regardless if it is a client-side browser app or a server-side browser
    app calling. This **does not apply** for API-based flows.

    As part of this change, the following endpoints have been removed:

    - `GET <ory-kratos-admin>/self-service/login/flows`;
    - `GET <ory-kratos-admin>/self-service/registration/flows`;
    - `GET <ory-kratos-admin>/self-service/verification/flows`;
    - `GET <ory-kratos-admin>/self-service/recovery/flows`;
    - `GET <ory-kratos-admin>/self-service/settings/flows`.

    Please ensure that your server-side applications use the public port (e.g.
    `GET <ory-kratos-public>/self-service/login/flows`) for fetching
    self-service flows going forward.

    If you use the SDKs, upgrading is easy by adding the `cookie` header when
    fetching the flows. This is only required when **using browser flows on the
    server side**.

    The following example illustrates a ExpressJS (NodeJS) server-side
    application fetching the self-service flows.

    ```diff
    app.get('some-route', (req: Request, res: Response) => {
    -   kratos.getSelfServiceLoginFlow(flow).then((flow) => /* ... */ )
    +   kratos.getSelfServiceLoginFlow(flow, req.header('cookie')).then((flow) => /* ... */ )

    -   kratos.getSelfServiceRecoveryFlow(flow).then((flow) => /* ... */ )
    +   kratos.getSelfServiceRecoveryFlow(flow, req.header('cookie')).then((flow) => /* ... */ )

    -   kratos.getSelfServiceRegistrationFlow(flow).then((flow) => /* ... */ )
    +   kratos.getSelfServiceRegistrationFlow(flow, req.header('cookie')).then((flow) => /* ... */ )

    -   kratos.getSelfServiceVerificationFlow(flow).then((flow) => /* ... */ )
    +   kratos.getSelfServiceVerificationFlow(flow, req.header('cookie')).then((flow) => /* ... */ )

    -   kratos.getSelfServiceSettingsFlow(flow).then((flow) => /* ... */ )
    +   kratos.getSelfServiceSettingsFlow(flow, undefined, req.header('cookie')).then((flow) => /* ... */ )
    })
    ```

    For concrete details, check out
    [the changes in the NodeJS app](https://github.com/ory/kratos-selfservice-ui-node/commit/e7fa292968111e06401fcfc9b1dd0e8e285a4d87).

1.  This patch refactors the logout functionality for browsers and APIs. It adds
    increased security and DoS-defenses to the logout flow. Previously, calling
    `GET /self-service/browser/flows/logout` would remove the session cookie and
    redirect the user to the logout endpoint. Now you have to make a call to
    `GET /self-service/logout/browser` which returns a JSON response including a
    `logout_url` URL to be used for logout. The call to
    `/self-service/logout/browser` must be made using AJAX with cookies enabled
    or by including the Ory Session Cookie in the `X-Session-Cookie` HTTP
    Header. You may also use the SDK method
    `createSelfServiceLogoutUrlForBrowsers` to do that.

    Additionally, the endpoint `DELETE /sessions` has been moved to
    `DELETE /self-service/logout/api`. Payloads and responses stay equal. The
    SDK method `revokeSession` has been renamed to
    `submitSelfServiceLogoutFlowWithoutBrowser`.

1.  Several SDK methods have been renamed:

    - `initializeSelfServiceRecoveryForNativeApps` to
      `initializeSelfServiceRecoveryWithoutBrowser`.
    - `initializeSelfServiceVerificationForNativeApps` to
      `initializeSelfServiceVerificationWithoutBrowser`
    - `initializeSelfServiceSettingsForNativeApps` to
      `initializeSelfServiceSettingsWithoutBrowser`.
    - `initializeSelfServiceregistrationForNativeApps` to
      `initializeSelfServiceregistrationWithoutBrowser`.
    - `initializeSelfServiceLoginForNativeApps` to
      `initializeSelfServiceLoginWithoutBrowser`.

    As in the previous release you may still use the old SDK if you do not want
    to deal with the SDK breaking changes for now.

## Ory Kratos v0.8 Breaking changes

Full list of changes:
https://github.com/ory/kratos/blob/v0.8.0-alpha.1/CHANGELOG.md#breaking-changes

1. The location of the homebrew tap has changed from `ory/ory/kratos` to
   `ory/tap/kratos`.

1. The self-service login flow's `forced` key has been renamed to `refresh`.

1. The SDKs are now generated with tag v0alpha2 to reflect that some signatures
   have changed in a breaking fashion. Please update your imports from
   `v0alpha1` to `v0alpha2`.

1. To support 2FA on non-browser (e.g. native mobile) apps we have added the Ory
   Session Token as a possible parameter to both
   `initializeSelfServiceLoginFlowWithoutBrowser` and
   `submitSelfServiceLoginFlow`. Depending on the SDK generator, the order of
   the arguments may have changed. In JavaScript:

   ```diff
   - .submitSelfServiceLoginFlow(flow.id, payload)
   + .submitSelfServiceLoginFlow(flow.id, sessionToken, payload)
   // or if the user has no session yet:
   + .submitSelfServiceLoginFlow(flow.id, undefined, payload)
   ```

1. To improve the overall API design we have changed the result of
   `POST /self-service/settings`. Instead of having flow be a key, the flow is
   now the response. The updated identity payload stays the same!

   ```diff
    {
   -  "flow": {
   -    "id": "flow-id-..."
   -    ...
   -  },
   +  "id": "flow-id-..."
   +  ...
      "identity": {
        "id": "identity-id-..."
      }
    }
   ```

1. The SMTPS scheme used in courier config url with cleartext/StartTLS/TLS SMTP
   connection types is now only supporting implicit TLS. For StartTLS and
   cleartext SMTP, please use the smtp scheme instead.

   Example:

   - SMTP Cleartext: `smtp://foo:bar@my-mailserver:1234/?disable_starttls=true`
   - SMTP with StartTLS: `smtps://foo:bar@my-mailserver:1234/` ->
     `smtp://foo:bar@my-mailserver:1234/`
   - SMTP with implicit TLS:
     `smtps://foo:bar@my-mailserver:1234/?legacy_ssl=true` ->
     `smtps://foo:bar@my-mailserver:1234/`

1. This patch changes the naming and number of prometheus metrics (see:
   https://github.com/ory/x/pull/379). In short: all metrics will have now
   `http_` prefix to conform to Prometheus best practices.
