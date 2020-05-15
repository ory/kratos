---
id: zero-trust-iap-proxy-identity-access-proxy
title: Zero Trust with IAP Proxy
---

import useBaseUrl from '@docusaurus/useBaseUrl'

The [Quickstart](../quickstart.mdx) covers a basic set up that uses a pipe in
SecureApp to forward requests to ORY Kratos.

Systems that have more than a few components often use Reverse Proxies such as
Nginx, Envoy, Kong to route and authorize traffic to applications. ORY Kratos
works very well in such a environment and the purpose of this guide is
clarifying how one can use a reverse proxy with IAP (Identity and Access Proxy)
capabilities to authorize incoming requests. In this tutorial, we will use ORY
Oathkeeper to achieve this.

This guide expects that you have familiarized yourself with ORY Kratos' concepts
and that builds on the components and flows established in the
[Quickstart](../quickstart.mdx).

To ensure that no one can access the dashboard without prior authentication
(login), we will use a reverse proxy
([ORY Oathkeeper](https://github.com/ory/oathkeeper)) to deny all
unauthenticated traffic to `http://secure-app/dashboard` and redirect the user
to the login page at `http://secure-app/auth/login`. Further, we will configure
access to `http://secure-app/auth/login` in such a way that access only works if
one is not yet authenticated.

## Running ORY Kratos and the ORY Oathkeeper Identity and Access Proxy

Clone the ORY Kratos repository and fetch the latest images:

```shell script
git clone https://github.com/ory/kratos.git
# or if you have git+ssh set up:
#  git clone git@github.com:ory/kratos.git
cd kratos
git checkout v0.3.0-alpha.1

docker pull oryd/kratos:latest-sqlite
docker pull oryd/kratos-selfservice-ui-node:latest
```

Next, run the quickstart and add the ORY Oathkeeper config:

```shell script
docker-compose \
  -f quickstart.yml \
  -f quickstart-oathkeeper.yml \
  up --build --force-recreate
```

This might take a minute or two. Once the output slows down and logs indicate a
healthy system you're ready to roll! A healthy system will show something along
the lines of (the order of messages might be reversed):

```
kratos_1                      | time="2020-01-20T14:52:13Z" level=info msg="Starting the admin httpd on: 0.0.0.0:4434"
kratos_1                      | time="2020-01-20T14:52:13Z" level=info msg="Starting the public httpd on: 0.0.0.0:4433"

oathkeeper_1                  | {"level":"info","msg":"TLS has not been configured for api, skipping","time":"2020-01-20T09:22:09Z"}
oathkeeper_1                  | {"level":"info","msg":"Listening on http://:4456","time":"2020-01-20T09:22:09Z"}
oathkeeper_1                  | {"level":"info","msg":"TLS has not been configured for proxy, skipping","time":"2020-01-20T09:22:09Z"}
oathkeeper_1                  | {"level":"info","msg":"Listening on http://:4455","time":"2020-01-20T09:22:09Z"}
```

:::note There are two important factors to get a fully functional system:

- You need to make sure that ports `4435`, `4455`, `4456`, `4433`, `4434`,
  `4436` >
  [are free](https://serverfault.com/questions/309052/check-if-port-is-open-or-closed-on-a-linux-server).
- Make sure to always use `127.0.0.1` as the hostname, never use `localhost`!
  This is important because browsers treat these two as separate domains and
  will therefore have issues with setting and using the right cookies. :::

### Network Architecture

This demo makes use of several services / Docker Images:

1. [ORY Kratos](https://github.com/ory/kratos)
2. The **SecureApp** - an
   [example application written in NodeJS](http://github.com/ory/kratos-selfservice-ui-node)
   that implements the login, registration, logout, ..., and dashboard screen.
3. A reverse proxy ([ORY Oathkeeper](https://github.com/ory/oathkeeper)) to
   protect the **SecureApp**.
4. An SMTP server with which ORY Kratos can send E-Mails with. We will use
   [MailHog](https://github.com/mailhog/MailHog), a minimalistic SMTP throaway
   server with an easy UI.

To better understand how everything is wired, let's take a look at the network
configuration. This assumes that you have at least some understanding of how
Docker (Compose) Networks work:

[![User Login and Registration Network Topology](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggVERcblxuc3ViZ3JhcGggaG5bSG9zdCBOZXR3b3JrXVxuICAgIEJbQnJvd3Nlcl1cbiAgICBCLS0-fENhbiBhY2Nlc3MgVVJMcyB2aWEgMTI3LjAuMC4xOjQ0NTV8T0tQSE5cbiAgICBCLS0-fENhbiBhY2Nlc3MgVUkgdmlhIDEyNy4wLjAuMTo0NDM2fFNNVFBVSVxuICAgIE9LUEhOKFtSZXZlcnNlIFByb3h5IGV4cG9zZWQgYXQgOjQ0NTVdKVxuICAgIFNNVFBVSShbTWFpbFNsdXJwZXIgVUkgZXhwb3NlZCBhdCA6NDQzNl0pXG5lbmRcblxuc3ViZ3JhcGggZG5bXCJJbnRlcm5hbCBEb2NrZXIgTmV0d29yayAoaW50cmFuZXQpXCJdXG4gICAgT0tQSE4tLT5PT1xuICAgIFNNVFBVSS0tPlNNVFBcbiAgICBPTy0tPnxQcm94aWVzIFVSTHNzIC8ub3J5L2tyYXRvcy9wdWJsaWMvKiB0b3xPS1xuICAgIE9PLS0-fFwiUHJveGllcyAvYXV0aC9sb2dpbiwgL2F1dGgvcmVnaXN0cmF0aW9uLCAvZGFzaGJvYXJkLCAuLi4gdG9cInxTQVxuICAgIFNBLS0-fFRhbGtzIHRvfE9LXG4gICAgT0stLT58U2VuZHMgbWFpbCB2aWF8U01UUFxuICAgIE9PLS0-fFZhbGlkYXRlcyBhdXRoIHNlc3Npb25zIHVzaW5nfE9LXG5cbiAgICBPS1tPUlkgS3JhdG9zXVxuICAgIE9PW1wiUmV2ZXJzZSBQcm94eSAoT1JZIE9hdGhrZWVwZXIpXCJdXG4gICAgU0FbXCJTZWN1cmVBcHAgKE9SWSBLcmF0b3MgU2VsZlNlcnZpY2UgVUkgTm9kZSBFeGFtcGxlKVwiXVxuICAgIFNNVFBbXCJTTVRQIFNlcnZlciAoTWFpbFNsdXJwZXIpXCJdXG5lbmRcbiIsIm1lcm1haWQiOnsidGhlbWUiOiJuZXV0cmFsIiwiZmxvd2NoYXJ0Ijp7InJhbmtTcGFjaW5nIjo2NSwibm9kZVNwYWNpbmciOjMwLCJjdXJ2ZSI6ImJhc2lzIn19fQ)](https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoiZ3JhcGggVERcblxuc3ViZ3JhcGggaG5bSG9zdCBOZXR3b3JrXVxuICAgIEJbQnJvd3Nlcl1cbiAgICBCLS0-fENhbiBhY2Nlc3MgVVJMcyB2aWEgMTI3LjAuMC4xOjQ0NTV8T0tQSE5cbiAgICBCLS0-fENhbiBhY2Nlc3MgVUkgdmlhIDEyNy4wLjAuMTo0NDM2fFNNVFBVSVxuICAgIE9LUEhOKFtSZXZlcnNlIFByb3h5IGV4cG9zZWQgYXQgOjQ0NTVdKVxuICAgIFNNVFBVSShbTWFpbFNsdXJwZXIgVUkgZXhwb3NlZCBhdCA6NDQzNl0pXG5lbmRcblxuc3ViZ3JhcGggZG5bXCJJbnRlcm5hbCBEb2NrZXIgTmV0d29yayAoaW50cmFuZXQpXCJdXG4gICAgT0tQSE4tLT5PT1xuICAgIFNNVFBVSS0tPlNNVFBcbiAgICBPTy0tPnxQcm94aWVzIFVSTHNzIC8ub3J5L2tyYXRvcy9wdWJsaWMvKiB0b3xPS1xuICAgIE9PLS0-fFwiUHJveGllcyAvYXV0aC9sb2dpbiwgL2F1dGgvcmVnaXN0cmF0aW9uLCAvZGFzaGJvYXJkLCAuLi4gdG9cInxTQVxuICAgIFNBLS0-fFRhbGtzIHRvfE9LXG4gICAgT0stLT58U2VuZHMgbWFpbCB2aWF8U01UUFxuICAgIE9PLS0-fFZhbGlkYXRlcyBhdXRoIHNlc3Npb25zIHVzaW5nfE9LXG5cbiAgICBPS1tPUlkgS3JhdG9zXVxuICAgIE9PW1wiUmV2ZXJzZSBQcm94eSAoT1JZIE9hdGhrZWVwZXIpXCJdXG4gICAgU0FbXCJTZWN1cmVBcHAgKE9SWSBLcmF0b3MgU2VsZlNlcnZpY2UgVUkgTm9kZSBFeGFtcGxlKVwiXVxuICAgIFNNVFBbXCJTTVRQIFNlcnZlciAoTWFpbFNsdXJwZXIpXCJdXG5lbmRcbiIsIm1lcm1haWQiOnsidGhlbWUiOiJuZXV0cmFsIiwiZmxvd2NoYXJ0Ijp7InJhbmtTcGFjaW5nIjo2NSwibm9kZVNwYWNpbmciOjMwLCJjdXJ2ZSI6ImJhc2lzIn19fQ)

As you can see, most requests are proxied through the Reverse Proxy
([ORY Oathkeeper](https://github.com/ory/oathkeeper)). The `quickstart.yml` file
also defines additional ports such as `4434`, `4456`, and others. These ports
are only there for debugging and playing around with and are not actually
required for the demo to work.

The next diagram shows how we've configured the routes in our Reverse Proxy
([ORY Oathkeeper](https://github.com/ory/oathkeeper)):

[![User Login and Registration Routes](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggVERcblxuc3ViZ3JhcGggcGlbUHVibGljIEludGVybmV0XVxuICAgIEJbQnJvd3Nlcl1cbmVuZFxuXG5zdWJncmFwaCB2cGNbVlBDIC8gQ2xvdWQgLyBEb2NrZXIgTmV0d29ya11cbnN1YmdyYXBoIFwiRGVtaWxpdGFyaXplZCBab25lIC8gRE1aXCJcbiAgICBPS1tPUlkgT2F0aGtlZXBlciA6NDQ1NV1cbiAgICBCIC0tPiBPS1xuZW5kXG5cbiAgICBPSyAtLT58XCJGb3J3YXJkcyB7LywvZGFzaGJvYXJkfSB0b1wifCBTQURcbiAgICBPSyAtLT58XCJGb3J3YXJkcyAvYXV0aC9sb2dvdXQgdG9cInwgU0FMVVxuICAgIE9LIC0tPnxcIkZvcndhcmRzIC9hdXRoL2xvZ2luIHRvXCJ8IFNBTElcbiAgICBPSyAtLT58XCJGb3J3YXJkcyAvYXV0aC9yZWdpc3RyYXRpb24gdG9cInwgU0FSXG4gICAgT0sgLS0-fFwiRm9yd2FyZHMgL2F1dGgvKiB0b1wifCBTQUFcbiAgICBPSyAtLT58XCJGb3J3YXJkcyAvLm9yeS9rcmF0b3MvcHVibGljLyogdG9cInwgS1BcblxuICAgIHN1YmdyYXBoIFwiUHJpdmF0ZSBTdWJuZXQgLyBJbnRyYW5ldFwiXG4gICAgS1sgT1JZIEtyYXRvcyBdXG5cbiAgICBLUChbIE9SWSBLcmF0b3MgUHVibGljIEFQSSBdKVxuICAgIEtBKFsgT1JZIEtyYXRvcyBBZG1pbiBBUEkgXSlcbiAgICBTQSAtLT4gS0FcbiAgICBLQSAtLmJlbG9uZ3MgdG8uLT4gS1xuICAgIEtQIC0uYmVsb25ncyB0by4tPiBLXG5cbiAgICBzdWJncmFwaCBzYVtcIlNlY3VyZUFwcCAvIGtyYXRvcy1zZXJsZnNlcnZpY2UtdWktbm9kZSBFeGFtcGxlXCJdXG5cbiAgICAgICAgU0FbU2VjdXJlQXBwXVxuICAgICAgICBTQUQgLS5iZWxvbmdzIHRvLi0-IFNBXG4gICAgICAgIFNBTFUgLS5iZWxvbmdzIHRvLi0-IFNBXG4gICAgICAgIFNBTEkgLS5iZWxvbmdzIHRvLi0-IFNBXG4gICAgICAgIFNBUiAtLmJlbG9uZ3MgdG8uLT4gU0FcbiAgICAgICAgU0FBIC0uYmVsb25ncyB0by4tPiBTQVxuXG4gICAgICAgIHN1YmdyYXBoIFwiSGFzIGFjdGl2ZSBsb2dpbiBzZXNzaW9uXCJcbiAgICAgICAgICAgIFNBRChbUm91dGUgL2Rhc2hib2FyZF0pXG4gICAgICAgICAgICBTQUxVKFtSb3V0ZSAvYXV0aC9sb2dvdXRdKVxuICAgICAgICBlbmRcblxuICAgICAgICBzdWJncmFwaCBcIk5vIGFjdGl2ZSBsb2dpbiBzZXNzaW9uXCJcbiAgICAgICAgICAgIFNBTEkoW1JvdXRlIC9hdXRoL2xvZ2luXSkgXG4gICAgICAgICAgICBTQVIoW1JvdXRlIC9hdXRoL3JlZ2lzdHJhdGlvbl0pIFxuICAgICAgICAgICAgU0FBKFtSb3V0ZSAvYXV0aC8uLi5dKVxuICAgICAgICBlbmRcbiAgICBlbmRcbiAgICBlbmRcblxuZW5kXG4iLCJtZXJtYWlkIjp7InRoZW1lIjoibmV1dHJhbCIsImZsb3djaGFydCI6eyJyYW5rU3BhY2luZyI6NzAsIm5vZGVTcGFjaW5nIjozMCwiY3VydmUiOiJiYXNpcyJ9fX0)](https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoiZ3JhcGggVERcblxuc3ViZ3JhcGggcGlbUHVibGljIEludGVybmV0XVxuICAgIEJbQnJvd3Nlcl1cbmVuZFxuXG5zdWJncmFwaCB2cGNbVlBDIC8gQ2xvdWQgLyBEb2NrZXIgTmV0d29ya11cbnN1YmdyYXBoIFwiRGVtaWxpdGFyaXplZCBab25lIC8gRE1aXCJcbiAgICBPS1tPUlkgT2F0aGtlZXBlciA6NDQ1NV1cbiAgICBCIC0tPiBPS1xuZW5kXG5cbiAgICBPSyAtLT58XCJGb3J3YXJkcyB7LywvZGFzaGJvYXJkfSB0b1wifCBTQURcbiAgICBPSyAtLT58XCJGb3J3YXJkcyAvYXV0aC9sb2dvdXQgdG9cInwgU0FMVVxuICAgIE9LIC0tPnxcIkZvcndhcmRzIC9hdXRoL2xvZ2luIHRvXCJ8IFNBTElcbiAgICBPSyAtLT58XCJGb3J3YXJkcyAvYXV0aC9yZWdpc3RyYXRpb24gdG9cInwgU0FSXG4gICAgT0sgLS0-fFwiRm9yd2FyZHMgL2F1dGgvKiB0b1wifCBTQUFcbiAgICBPSyAtLT58XCJGb3J3YXJkcyAvLm9yeS9rcmF0b3MvcHVibGljLyogdG9cInwgS1BcblxuICAgIHN1YmdyYXBoIFwiUHJpdmF0ZSBTdWJuZXQgLyBJbnRyYW5ldFwiXG4gICAgS1sgT1JZIEtyYXRvcyBdXG5cbiAgICBLUChbIE9SWSBLcmF0b3MgUHVibGljIEFQSSBdKVxuICAgIEtBKFsgT1JZIEtyYXRvcyBBZG1pbiBBUEkgXSlcbiAgICBTQSAtLT4gS0FcbiAgICBLQSAtLmJlbG9uZ3MgdG8uLT4gS1xuICAgIEtQIC0uYmVsb25ncyB0by4tPiBLXG5cbiAgICBzdWJncmFwaCBzYVtcIlNlY3VyZUFwcCAvIGtyYXRvcy1zZXJsZnNlcnZpY2UtdWktbm9kZSBFeGFtcGxlXCJdXG5cbiAgICAgICAgU0FbU2VjdXJlQXBwXVxuICAgICAgICBTQUQgLS5iZWxvbmdzIHRvLi0-IFNBXG4gICAgICAgIFNBTFUgLS5iZWxvbmdzIHRvLi0-IFNBXG4gICAgICAgIFNBTEkgLS5iZWxvbmdzIHRvLi0-IFNBXG4gICAgICAgIFNBUiAtLmJlbG9uZ3MgdG8uLT4gU0FcbiAgICAgICAgU0FBIC0uYmVsb25ncyB0by4tPiBTQVxuXG4gICAgICAgIHN1YmdyYXBoIFwiSGFzIGFjdGl2ZSBsb2dpbiBzZXNzaW9uXCJcbiAgICAgICAgICAgIFNBRChbUm91dGUgL2Rhc2hib2FyZF0pXG4gICAgICAgICAgICBTQUxVKFtSb3V0ZSAvYXV0aC9sb2dvdXRdKVxuICAgICAgICBlbmRcblxuICAgICAgICBzdWJncmFwaCBcIk5vIGFjdGl2ZSBsb2dpbiBzZXNzaW9uXCJcbiAgICAgICAgICAgIFNBTEkoW1JvdXRlIC9hdXRoL2xvZ2luXSkgXG4gICAgICAgICAgICBTQVIoW1JvdXRlIC9hdXRoL3JlZ2lzdHJhdGlvbl0pIFxuICAgICAgICAgICAgU0FBKFtSb3V0ZSAvYXV0aC8uLi5dKVxuICAgICAgICBlbmRcbiAgICBlbmRcbiAgICBlbmRcblxuZW5kXG4iLCJtZXJtYWlkIjp7InRoZW1lIjoibmV1dHJhbCIsImZsb3djaGFydCI6eyJyYW5rU3BhY2luZyI6NzAsIm5vZGVTcGFjaW5nIjozMCwiY3VydmUiOiJiYXNpcyJ9fX0)

You might notice that we're also proxying requests to ORY Kratos' Public API. We
are doing this because that way all requests are going to and coming from the
same hostname. This avoids common cross-domain issues with cookies.

## Perform registration, logout, login

Enough theory, it's time to get this thing going! Let's start by trying to open
the dashboard - **go to
[127.0.0.1:4455/dashboard](http://127.0.0.1:4455/dashboard)**.

Check the [Quickstart](../quickstart.mdx) for the other flows!

## Configuration

You can find all configuration files for this quickstart guide in
`./contrib/quickstart`, `./quickstart.yml`, `./quickstart-oathkeeper.yml`.

### ORY Oathkeeper: Identity and Access Proxy

All configuration for [ORY Oathkeeper](https://www.ory.sh/oathkeeper/) resides
in `./contrib/quickstart/oathkeeper`.

#### Configuration

We define several configuration options for ORY Oathkeeper, such as the port
where the proxy should run or where to load the access rules from.

##### Cookie Session Authenticator

The
[Cookie Session Authenticator](https://www.ory.sh/docs/oathkeeper/pipeline/authn#cookie_session)
is enabled and points to
[ORY Kratos' `/sessions/whoami` API](../reference/api.md). It uses the
`ory_kratos_session` cookie to identify if a request contains a session or not:

```yaml title="contrib/quickstart/oathkeeper/.oathkeeper.yml"
# ...
authenticators
  cookie_session:
    enabled: true
    config:
      check_session_url: http://kratos:4433/sessions/whoami
      preserve_path: true
      extra_from: "@this"
      subject_from: "identity.id"
      only:
        - ory_kratos_session
# ...
```

It's more or less doing what the `needsLogin` function does in the
[Quickstart](../quickstart.mdx).

#### Anonymous Authenticator

The
[Anonymous Authenticator](https://www.ory.sh/docs/oathkeeper/pipeline/authn#anonymous)
is useful for endpoints that do not need login, such as the registration screen:

```yaml title="contrib/quickstart/oathkeeper/.oathkeeper.yml"
# ...
authenticators
  anonymous:
    enabled: true
    config:
      subject: guest
# ...
```

#### Allowed Authorizer

The
[Allowed Authenticator](https://www.ory.sh/docs/oathkeeper/pipeline/authz#allowed)
simply allows all users to access the URL. Since we don't have RBAC or ACL in
place for this example, this will be enough.

```yaml title="contrib/quickstart/oathkeeper/.oathkeeper.yml"
# ...
authorizers
  allowed:
    enabled: true
# ...
```

### ID Token Mutator

The
[ID Token Mutator](https://www.ory.sh/docs/oathkeeper/pipeline/mutator#id_token)
takes all the available session information and puts it into a JSON Web Token
(JWT). The protected SecureApp will now receive `Authorization: bearer <jwt...>`
in the HTTP Header instead of `Cookie: ory_kratos_session=...`. The JWT is
signed using a RS256 key. To verify the JWT we can use the public key provided
by ORY Oathkeeper's JWKS API `http://127.0.0.1:4456/.well-known/jwks.json`. You
can generate the RS256 key yourself by running:
`oathkeeper credentials generate --alg RS256 > id_token.jwks.json`.

We also enabled the
[NoOp Mutator](https://www.ory.sh/docs/oathkeeper/pipeline/mutator#) for the
login, registration, ... endpoints:

```yaml title="contrib/quickstart/oathkeeper/.oathkeeper.yml"
mutators:
  noop:
    enabled: true

  id_token:
    enabled: true
    config:
      issuer_url: http://127.0.0.1:4455/
      jwks_url: file:///etc/config/oathkeeper/id_token.jwks.json
      claims: |
        {
          "session": {{ .Extra | toJson }}
        }
```

You could obviously also use other mutators such as the
[Header Mutator](https://www.ory.sh/docs/oathkeeper/pipeline/mutator#header) and
use headers such as `X-User-ID` instead of the JWT.

### Error Handling

We configure the error handling in such a way that a missing or invalid login
session (when accessed from a browser) leads to a redirect to `/auth/login`:

```yaml title="contrib/quickstart/oathkeeper/.oathkeeper.yml"
errors:
  fallback:
    - json

  handlers:
    redirect:
      enabled: true
      config:
        to: http://127.0.0.1:4455/auth/login
        when:
          - error:
              - unauthorized
              - forbidden
            request:
              header:
                accept:
                  # We don't want this for application/json requests, only browser requests!
                  - text/html
    json:
      enabled: true
      config:
        verbose: true
```

### Access Rules

We use [glob matching](https://github.com/gobwas/glob) to match the HTTP
requests for our access rules:

```yaml title="contrib/quickstart/oathkeeper/.oathkeeper.yml"
access_rules:
  matching_strategy: glob
  repositories:
    - file:///etc/config/oathkeeper/`access-rules.yml`
```

In `access-rules.yml` we define three rules. The first rule forwards all traffic
matching `http://127.0.0.1:4455/.ory/kratos/public/` to ORY Kratos' Public API:

```yaml title="contrib/quickstart/oathkeeper/access-rules.yml"
- id: 'ory:kratos:public'
  upstream:
    preserve_host: true
    url: 'http://kratos:4433'
    strip_path: /.ory/kratos/public
  match:
    url: 'http://127.0.0.1:4455/.ory/kratos/public/<**>'
    methods:
      - GET
      - POST
      - PUT
      - DELETE
      - PATCH
  authenticators:
    - handler: noop
  authorizer:
    handler: allow
  mutators:
    - handler: noop
```

The second rule allows anonymous requests to login, registration, re-send
verification email, and the error page plus any assets:

```yaml title="contrib/quickstart/oathkeeper/access-rules.yml"
# ...
- id: 'ory:kratos-selfservice-ui-node:anonymous'
  upstream:
    preserve_host: true
    url: 'http://kratos-selfservice-ui-node:4435'
  match:
    url: 'http://127.0.0.1:4455/<{error,verify,auth/*,**.css,**.js}{/,}>'
    methods:
      - GET
  authenticators:
    - handler: anonymous
  authorizer:
    handler: allow
  mutators:
    - handler: noop
```

And the final rule requires a valid session before allowing requests to the
dashboard and user settings:

```yaml title="contrib/quickstart/oathkeeper/access-rules.yml"
# ...
- id: 'ory:kratos-selfservice-ui-node:protected'
  upstream:
    preserve_host: true
    url: 'http://kratos-selfservice-ui-node:4435'
  match:
    url: 'http://127.0.0.1:4455/<{,debug,dashboard,settings}{/,}>'
    methods:
      - GET
  authenticators:
    - handler: cookie_session
  authorizer:
    handler: allow
  mutators:
    - handler: id_token
  errors:
    - handler: redirect
      config:
        to: http://127.0.0.1:4455/auth/login
```

## Cleaning up Docker

To clean everything up, you need to bring down the Docker Compose environment
and remove all mounted volumes.

```shell script
docker-compose -f quickstart.yml -f quickstart-oathkeeper.yml down -v
docker-compose -f quickstart.yml -f quickstart-oathkeeper.yml rm -f -s -v
```
