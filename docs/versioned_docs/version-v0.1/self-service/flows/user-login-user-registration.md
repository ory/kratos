---
id: user-login-user-registration
title: User Login and User Registration
---

ORY Kratos supports two type of login and registration flows:

- Browser-based (easy): This flow works for all applications running on top of a
  browser. Websites, single-page apps, Cordova/Ionic, and so on.
- API-based (advanced): This flow works for native applications like iOS
  (Swift), Android (Java), Microsoft (.NET), React Native, Electron, and others.

## Self-Service User Login and User Registration for Browser Applications

ORY Kratos supports browser applications that run on server-side (e.g. Java,
NodeJS, PHP) as well as client-side (e.g. JQuery, ReactJS, AngularJS, ...).

Browser-based login and registration makes use of three core HTTP technologies:

- HTTP Redirects
- HTTP POST (`application/json`, `application/x-www-urlencoded`) and RESTful GET
  requests.
- HTTP Cookies to prevent CSRF and Session Hijaking attack vectors.

The browser flow is the easiest and most secure to set up and integrated with.
ORY Kratos takes care of all required session and CSRF cookies and ensures that
all security requirements are fulfilled.

Future versions of ORY Kratos will be able to deal with multi-domain
environments that require SSO. For example, one account would be used to sign
into both `mydomain.com` and `anotherdomain.org`. A common real-world example is
using your Google account to seamlessly be signed into YouTube and Google at the
same time.

This flow is not suitable for scenarios where you use purely programmatic
clients that do not work well with HTTP Cookies and HTTP Redirects.

### The Login and Registration User Interface

In Browser Applications, the User Interface is typically rendered as an HTML
Form:

```html
<!-- Login -->
<form action="..." method="POST">
  <input type="text" name="identifier" placeholder="Enter your username" />
  <input type="password" name="password" placeholder="Enter your password" />
  <input type="hidden" name="csrf_token" value="cdef..." />
  <input type="submit" />
</form>

<!-- Registration -->
<form action="..." method="POST">
  <input type="email" name="email" placeholder="Enter your E-Mail Address" />
  <input type="password" name="password" placeholder="Enter your password" />
  <input
    type="first_name"
    name="password"
    placeholder="Enter your First Name"
  />
  <input type="last_name" name="password" placeholder="Enter your Last Name" />
  <input type="hidden" name="csrf_token" value="cdef..." />
  <input type="submit" />
</form>
```

Depending on the type of login flows you want to support, this might also be a
"Sign up/in with GitHub" flow:

```html
<!-- Login and Registration -->
<form action="..." method="POST">
  <input type="hidden" name="csrf_token" value="cdef..." />

  <!-- Basically <a href="https://github.com/login/oauth/authorize?...">Sign up/in with GitHub</a> -->
  <input type="submit" name="provider" value="GitHub" />
</form>
```

In stark contrast to other Identity Systems, ORY Kratos does not render this
HTML. Instead, you need to implement the HTML code in your application (e.g.
NodeJS + ExpressJS, Java, PHP, ReactJS, ...), which gives you extreme
flexibility and customizability in your user interface flows and designs.

Each Login and Registration Strategy (e.g.
[Username and Password](../strategies/username-email-password.md),
[Social Sign In](../strategies/openid-connect-social-sign-in-oauth2.md),
Passwordless, ...) works a bit differently but they all boil down to the same
abstract sequence:

[![Abstract Login and Registration User Flow](https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gIHBhcnRpY2lwYW50IEIgYXMgQnJvd3NlclxuICBwYXJ0aWNpcGFudCBLIGFzIE9SWSBLcmF0b3NcbiAgcGFydGljaXBhbnQgQSBhcyBZb3VyIEFwcGxpY2F0aW9uXG5cblxuICBCLT4-SzogSW5pdGlhdGUgTG9naW5cbiAgSy0-PkI6IFJlZGlyZWN0cyB0byB5b3VyIEFwcGxpY2F0aW9uJ3MgL2xvZ2luIGVuZHBvaW50XG4gIEItPj5BOiBDYWxscyAvbG9naW5cbiAgQS0tPj5LOiBGZXRjaGVzIGRhdGEgdG8gcmVuZGVyIGZvcm1zIGV0Y1xuICBCLS0-PkE6IEZpbGxzIG91dCBmb3JtcywgY2xpY2tzIGUuZy4gXCJTdWJtaXQgTG9naW5cIlxuICBCLT4-SzogUE9TVHMgZGF0YSB0b1xuICBLLS0-Pks6IFByb2Nlc3NlcyBMb2dpbiBJbmZvXG5cbiAgYWx0IExvZ2luIGRhdGEgdmFsaWRcbiAgICBLLS0-PkI6IFNldHMgc2Vzc2lvbiBjb29raWVcbiAgICBLLT4-QjogUmVkaXJlY3RzIHRvIGUuZy4gRGFzaGJvYXJkXG4gIGVsc2UgTG9naW4gZGF0YSBpbnZhbGlkXG4gICAgSy0tPj5COiBSZWRpcmVjdHMgdG8geW91ciBBcHBsaWNhaXRvbidzIC9sb2dpbiBlbmRwb2ludFxuICAgIEItPj5BOiBDYWxscyAvbG9naW5cbiAgICBBLS0-Pks6IEZldGNoZXMgZGF0YSB0byByZW5kZXIgZm9ybSBmaWVsZHMgYW5kIGVycm9yc1xuICAgIEItLT4-QTogRmlsbHMgb3V0IGZvcm1zIGFnYWluLCBjb3JyZWN0cyBlcnJvcnNcbiAgICBCLT4-SzogUE9TVHMgZGF0YSBhZ2FpbiAtIGFuZCBzbyBvbi4uLlxuICBlbmRcbiIsIm1lcm1haWQiOnsidGhlbWUiOiJuZXV0cmFsIiwic2VxdWVuY2VEaWFncmFtIjp7ImRpYWdyYW1NYXJnaW5YIjoxNSwiZGlhZ3JhbU1hcmdpblkiOjE1LCJib3hUZXh0TWFyZ2luIjowLCJub3RlTWFyZ2luIjoxNSwibWVzc2FnZU1hcmdpbiI6NDUsIm1pcnJvckFjdG9ycyI6dHJ1ZX19fQ)](https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gIHBhcnRpY2lwYW50IEIgYXMgQnJvd3NlclxuICBwYXJ0aWNpcGFudCBLIGFzIE9SWSBLcmF0b3NcbiAgcGFydGljaXBhbnQgQSBhcyBZb3VyIEFwcGxpY2F0aW9uXG5cblxuICBCLT4-SzogSW5pdGlhdGUgTG9naW5cbiAgSy0-PkI6IFJlZGlyZWN0cyB0byB5b3VyIEFwcGxpY2F0aW9uJ3MgL2xvZ2luIGVuZHBvaW50XG4gIEItPj5BOiBDYWxscyAvbG9naW5cbiAgQS0tPj5LOiBGZXRjaGVzIGRhdGEgdG8gcmVuZGVyIGZvcm1zIGV0Y1xuICBCLS0-PkE6IEZpbGxzIG91dCBmb3JtcywgY2xpY2tzIGUuZy4gXCJTdWJtaXQgTG9naW5cIlxuICBCLT4-SzogUE9TVHMgZGF0YSB0b1xuICBLLS0-Pks6IFByb2Nlc3NlcyBMb2dpbiBJbmZvXG5cbiAgYWx0IExvZ2luIGRhdGEgdmFsaWRcbiAgICBLLS0-PkI6IFNldHMgc2Vzc2lvbiBjb29raWVcbiAgICBLLT4-QjogUmVkaXJlY3RzIHRvIGUuZy4gRGFzaGJvYXJkXG4gIGVsc2UgTG9naW4gZGF0YSBpbnZhbGlkXG4gICAgSy0tPj5COiBSZWRpcmVjdHMgdG8geW91ciBBcHBsaWNhaXRvbidzIC9sb2dpbiBlbmRwb2ludFxuICAgIEItPj5BOiBDYWxscyAvbG9naW5cbiAgICBBLS0-Pks6IEZldGNoZXMgZGF0YSB0byByZW5kZXIgZm9ybSBmaWVsZHMgYW5kIGVycm9yc1xuICAgIEItLT4-QTogRmlsbHMgb3V0IGZvcm1zIGFnYWluLCBjb3JyZWN0cyBlcnJvcnNcbiAgICBCLT4-SzogUE9TVHMgZGF0YSBhZ2FpbiAtIGFuZCBzbyBvbi4uLlxuICBlbmRcbiIsIm1lcm1haWQiOnsidGhlbWUiOiJuZXV0cmFsIiwic2VxdWVuY2VEaWFncmFtIjp7ImRpYWdyYW1NYXJnaW5YIjoxNSwiZGlhZ3JhbU1hcmdpblkiOjE1LCJib3hUZXh0TWFyZ2luIjowLCJub3RlTWFyZ2luIjoxNSwibWVzc2FnZU1hcmdpbiI6NDUsIm1pcnJvckFjdG9ycyI6dHJ1ZX19fQ)

The exact data being fetched and the step _"Processes Login / Registration
Info"_ depend, of course, on the actual Strategy being used. But it is important
to understand that **"Your Application"** is responsible for rendering the
actual Login and Registration HTML Forms. You can of course implement one app
for rendering all the Login, Registration, ... screens, and another app (think
"Service Oriented Architecture", "Micro-Services" or "Service Mesh") is
responsible for rendering your Dashboards, Management Screens, and so on.

> It is highly RECOMMENDED to all the applications (or "services"), including
> ORY Kratos, behind a common API Gateway or Reverse Proxy. This greatly reduces
> the amount of work you have to do to get all the Cookies working properly. We
> RECOMMEND using [ORY Oathkeeper](http://github.com/ory/oathkeeper) for this as
> it integrates best with the ORY Ecosystem and because all of our examples use
> ORY Oathkeeper. You MAY of course use any other reverse proxy (Envoy, AWS API
> Gateway, Ambassador, Nginx, Kong, ...), but we do not have examples or guides
> for those at this time.

### Code

Because Login and Registration are so similar, we can use one common piece of
code to cover both. A functioning example of the code and approach used here can
be found on
[github.com/ory/kratos-selfservice-ui-node](https://github.com/ory/kratos-selfservice-ui-node).

The code example used here is universal and does not use an SDK because we want
you to understand the fundamentals of how this flow works.

While this example assumes a Server-Side Application, a Client-Side (e.g.
ReactJS) Application would work the same, but use ORY Kratos' Public API
instead.

```js
const fetch = require('node-fetch');

const config = {
  kratos: {
    // The browser config key is used to redirect the user. It reflects where ORY Kratos' Public API
    // is accessible from. Here, we're assuming traffic going to `http://example.org/.ory/kratos/public/`
    // will be forwarded to ORY Kratos' Public API.
    browser: 'http://example.org/.ory/kratos/public/',

    // This is where ORY Kratos' Admin API is accessible.
    admin: 'https://ory-kratos-admin.example-org.vpc/',
  },
};

// The parameter "flow" can be "login" and "registration".
// You would register the two routes in express js like this:
//
//  app.get('/auth/registration', authHandler('registration'))
//  app.get('/auth/login', authHandler('login'))

export const authHandler = (flow) => (req, res, next) => {
  // The request ID is used to identify the login and registraion request and
  // return data like the csrf_token and so on.
  const request = req.query.request;
  if (!request) {
    console.log('No request found in URL, initializing ${flow} flow.');
    res.redirect(`${config.kratos.browser}/auth/browser/${flow}`);
    return;
  }

  // This is the ORY Kratos URL. If this app and ORY Kratos are running
  // on the same (e.g. Kubernetes) cluster, this should be ORY Kratos's internal hostname.
  const url = new URL(`${config.kratos.admin}/auth/browser/requests/${flow}`);
  url.searchParams.set('request', request);

  fetch(url.toString())
    .then((response) => {
      // A 404 code means that this code does not exist. We'll retry by re-initiating the flow.
      if (response.status == 404) {
        res.redirect(`${config.kratos.browser}/auth/browser/${flow}`);
        return;
      }

      return response.json();
    })
    .then((request) => {
      // Request contains all the request data for this Registration request.
      // You can process that data here, if you want.

      // Lastly, you probably want to render the data using a view (e.g. Jade Template):
      res.render(flow, request);
    });
};
```

For details on payloads and potential HTML snippets consult the individual
Self-Service Strategies for:

- [Username and Password Strategy](../strategies/username-email-password.md)
- [Social Sign In Strategy](../strategies/openid-connect-social-sign-in-oauth2.md)

### Server-Side Browser Applications

Let's take a look at the concrete network topologies, calls, and payloads. Here,
we're assuming that you're running a server-side browser application (written in
e.g. PHP, Java, NodeJS) to render the login and registration screen on the
server and make all API calls from that server code. The counterpart to this
would be a client-side browser application (written in e.g. Vanilla JavaScript,
JQuery, ReactJS, AngularJS, ...) that uses AJAX requests to fetch data. For
these type of applications, read this section first and go to section
[Client-Side Browser Applications](#client-side-browser-applications) next.

#### Network Topology

Your Server-Side Application and ORY Kratos are deployed in a Virtual Private
Cluster that can not be accessed from the public internet directly. Instead,
only ORY Oathkeeper can be accessed from the public internet and proxies
incoming requests to the appropriate service:

- Public internet traffic to domain `example.org` is sent to ORY Oathkeeper
  which in turn:
  - proxies URLs matching `https://example.org/auth/login` to your Server-Side
    Application available at
    `https://your-server-side-application.example-org.vpc/auth/login`
  - proxies URLs matching `https://example.org/auth/registration` to your
    Server-Side Application available at
    `https://your-server-side-application.example-org.vpc/auth/registration`
  - `https://example.org/.ory/kratos/public/*` is proxied to
    `https://ory-kratos-public.example-org.vpc/`
- `https://ory-kratos-admin.example-org.vpc/` exposes ORY Kratos' Admin API and
  is not accessible by the open internet and ideally only by Your Server-Side
  Application.
- `https://ory-kratos-public.example-org.vpc` exposes ORY Kratos' Public API and
  is ideally only accessible by ORY Oathkeeper.
- `https://your-server-side-application.example-org.vpc` exposes your
  Server-Side Application and is ideally only accessible by ORY Oathkeeper.

The ORY Kratos Admin API is exposed only in the intranet and only the
Server-Side Application should be able to talk to it.

[![User Login and Registration Network Topology for Server-Side Applications](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggVERcblxuc3ViZ3JhcGggcGlbUHVibGljIEludGVybmV0XVxuICAgIEJbQnJvd3Nlcl1cbmVuZFxuXG5zdWJncmFwaCB2cGNbVlBDIC8gQ2xvdWQgLyBEb2NrZXIgTmV0d29ya11cbnN1YmdyYXBoIFwiRGVtaWxpdGFyaXplZCBab25lIC8gRE1aXCJcbiAgICBPS1tPUlkgT2F0aGtlZXBlciA6NDQ1NV1cbiAgICBCIC0tPiBPS1xuZW5kXG5cbiAgICBPSyAtLT58XCJGb3J3YXJkcyAvYXV0aC9sb2dpbiB0b1wifCBTQUxJXG4gICAgT0sgLS0-fFwiRm9yd2FyZHMgL2F1dGgvcmVnaXN0cmF0aW9uIHRvXCJ8IFNBTFJcbiAgICBPSyAtLT58XCJGb3J3YXJkcyAvLm9yeS9rcmF0b3MvcHVibGljLyogdG9cInwgS1BcblxuICAgIHN1YmdyYXBoIFwiUHJpdmF0ZSBTdWJuZXQgLyBJbnRyYW5ldFwiXG4gICAgS1sgT1JZIEtyYXRvcyBdXG5cbiAgICBLUChbIE9SWSBLcmF0b3MgUHVibGljIEFQSSBdKVxuICAgIEtBKFsgT1JZIEtyYXRvcyBBZG1pbiBBUEkgXSlcbiAgICBTQSAtLT58XCJ0YWxrcyB0b1wifCBLQVxuICAgIEtBIC0uYmVsb25ncyB0by4tPiBLXG4gICAgS1AgLS5iZWxvbmdzIHRvLi0-IEtcblxuICAgIHN1YmdyYXBoIHNhW1wiWW91ciBBcHBsaWNhdGlvblwiXVxuXG4gICAgICAgIFNBW1wiWW91ciBTZXJ2ZXItU2lkZSBBcHBsaWNhdGlvblwiXVxuICAgICAgICBTQUxJIC0uYmVsb25ncyB0by4tPiBTQVxuICAgICAgICBTQUxSIC0uYmVsb25ncyB0by4tPiBTQVxuICAgICAgICBTQUxJKFtSb3V0ZSAvYXV0aC9sb2dpbl0pIFxuICAgICAgICBTQUxSKFtSb3V0ZSAvYXV0aC9yZWdpc3RyYXRpb25dKSBcbiAgICBlbmRcbiAgICBlbmRcblxuZW5kXG4iLCJtZXJtYWlkIjp7InRoZW1lIjoibmV1dHJhbCIsImZsb3djaGFydCI6eyJyYW5rU3BhY2luZyI6NzAsIm5vZGVTcGFjaW5nIjozMCwiY3VydmUiOiJiYXNpcyJ9fSwidXBkYXRlRWRpdG9yIjpmYWxzZX0)](https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoiZ3JhcGggVERcblxuc3ViZ3JhcGggcGlbUHVibGljIEludGVybmV0XVxuICAgIEJbQnJvd3Nlcl1cbmVuZFxuXG5zdWJncmFwaCB2cGNbVlBDIC8gQ2xvdWQgLyBEb2NrZXIgTmV0d29ya11cbnN1YmdyYXBoIFwiRGVtaWxpdGFyaXplZCBab25lIC8gRE1aXCJcbiAgICBPS1tPUlkgT2F0aGtlZXBlciA6NDQ1NV1cbiAgICBCIC0tPiBPS1xuZW5kXG5cbiAgICBPSyAtLT58XCJGb3J3YXJkcyAvYXV0aC9sb2dpbiB0b1wifCBTQUxJXG4gICAgT0sgLS0-fFwiRm9yd2FyZHMgL2F1dGgvcmVnaXN0cmF0aW9uIHRvXCJ8IFNBTFJcbiAgICBPSyAtLT58XCJGb3J3YXJkcyAvLm9yeS9rcmF0b3MvcHVibGljLyogdG9cInwgS1BcblxuICAgIHN1YmdyYXBoIFwiUHJpdmF0ZSBTdWJuZXQgLyBJbnRyYW5ldFwiXG4gICAgS1sgT1JZIEtyYXRvcyBdXG5cbiAgICBLUChbIE9SWSBLcmF0b3MgUHVibGljIEFQSSBdKVxuICAgIEtBKFsgT1JZIEtyYXRvcyBBZG1pbiBBUEkgXSlcbiAgICBTQSAtLT58XCJ0YWxrcyB0b1wifCBLQVxuICAgIEtBIC0uYmVsb25ncyB0by4tPiBLXG4gICAgS1AgLS5iZWxvbmdzIHRvLi0-IEtcblxuICAgIHN1YmdyYXBoIHNhW1wiWW91ciBBcHBsaWNhdGlvblwiXVxuXG4gICAgICAgIFNBW1wiWW91ciBTZXJ2ZXItU2lkZSBBcHBsaWNhdGlvblwiXVxuICAgICAgICBTQUxJIC0uYmVsb25ncyB0by4tPiBTQVxuICAgICAgICBTQUxSIC0uYmVsb25ncyB0by4tPiBTQVxuICAgICAgICBTQUxJKFtSb3V0ZSAvYXV0aC9sb2dpbl0pIFxuICAgICAgICBTQUxSKFtSb3V0ZSAvYXV0aC9yZWdpc3RyYXRpb25dKSBcbiAgICBlbmRcbiAgICBlbmRcblxuZW5kXG4iLCJtZXJtYWlkIjp7InRoZW1lIjoibmV1dHJhbCIsImZsb3djaGFydCI6eyJyYW5rU3BhY2luZyI6NzAsIm5vZGVTcGFjaW5nIjozMCwiY3VydmUiOiJiYXNpcyJ9fSwidXBkYXRlRWRpdG9yIjpmYWxzZX0)

#### User Login and User Registration Process Sequence

The Login and Registration User Flow is composed of several high-level steps
summarized in this state diagram:

[![User Login and Registration State Machine](https://mermaid.ink/img/eyJjb2RlIjoic3RhdGVEaWFncmFtXG4gIHMxOiBVc2VyIGJyb3dzZXMgYXBwXG4gIHMyOiBFeGVjdXRlIFwiQmVmb3JlIExvZ2luL1JlZ2lzdHJhdGlvbiBKb2IocylcIlxuICBzMzogVXNlciBJbnRlcmZhY2UgQXBwbGljYXRpb24gcmVuZGVycyBcIkxvZ2luL1JlZ2lzdHJhdGlvbiBSZXF1ZXN0XCJcbiAgczQ6IEV4ZWN1dGUgXCJBZnRlciBMb2dpbi9SZWdpc3RyYXRpb24gSm9iKHMpXCJcbiAgczU6IFVwZGF0ZSBcIkxvZ2luL1JlZ2lzdHJhdGlvbiBSZXF1ZXN0XCIgd2l0aCBFcnJvciBDb250ZXh0KHMpXG4gIHM2OiBMb2dpbi9SZWdpc3RyYXRpb24gc3VjY2Vzc2Z1bFxuXG5cblxuXHRbKl0gLS0-IHMxXG4gIHMxIC0tPiBzMiA6IFVzZXIgY2xpY2tzIFwiTG9nIGluIC8gU2lnbiB1cFwiXG4gIHMyIC0tPiBFcnJvciA6IEEgam9iIGZhaWxzXG4gIHMyIC0tPiBzMyA6IFVzZXIgaXMgcmVkaXJlY3RlZCB0byBMb2dpbi9SZWdpc3RyYXRpb24gVUkgVVJMXG4gIHMzIC0tPiBzNCA6IFVzZXIgcHJvdmlkZXMgdmFsaWQgY3JlZGVudGlhbHMvcmVnaXN0cmF0aW9uIGRhdGFcbiAgczMgLS0-IHM1IDogVXNlciBwcm92aWRlcyBpbnZhbGlkIGNyZWRlbnRpYWxzL3JlZ2lzdHJhdGlvbiBkYXRhXG4gIHM1IC0tPiBzMyA6IFVzZXIgaXMgcmVkaXJlY3RlZCB0byBMb2dpbi9SZWdpc3RyYXRpb24gVUkgVVJMXG4gIHM0IC0tPiBFcnJvciA6IEEgam9iIGZhaWxzXG4gIHM0IC0tPiBzNlxuICBzNiAtLT4gWypdXG5cbiAgRXJyb3IgLS0-IFsqXVxuXG5cbiIsIm1lcm1haWQiOnsidGhlbWUiOiJkZWZhdWx0In0sInVwZGF0ZUVkaXRvciI6ZmFsc2V9)](https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic3RhdGVEaWFncmFtXG4gIHMxOiBVc2VyIGJyb3dzZXMgYXBwXG4gIHMyOiBFeGVjdXRlIFwiQmVmb3JlIExvZ2luL1JlZ2lzdHJhdGlvbiBKb2IocylcIlxuICBzMzogVXNlciBJbnRlcmZhY2UgQXBwbGljYXRpb24gcmVuZGVycyBcIkxvZ2luL1JlZ2lzdHJhdGlvbiBSZXF1ZXN0XCJcbiAgczQ6IEV4ZWN1dGUgXCJBZnRlciBMb2dpbi9SZWdpc3RyYXRpb24gSm9iKHMpXCJcbiAgczU6IFVwZGF0ZSBcIkxvZ2luL1JlZ2lzdHJhdGlvbiBSZXF1ZXN0XCIgd2l0aCBFcnJvciBDb250ZXh0KHMpXG4gIHM2OiBMb2dpbi9SZWdpc3RyYXRpb24gc3VjY2Vzc2Z1bFxuXG5cblxuXHRbKl0gLS0-IHMxXG4gIHMxIC0tPiBzMiA6IFVzZXIgY2xpY2tzIFwiTG9nIGluIC8gU2lnbiB1cFwiXG4gIHMyIC0tPiBFcnJvciA6IEEgam9iIGZhaWxzXG4gIHMyIC0tPiBzMyA6IFVzZXIgaXMgcmVkaXJlY3RlZCB0byBMb2dpbi9SZWdpc3RyYXRpb24gVUkgVVJMXG4gIHMzIC0tPiBzNCA6IFVzZXIgcHJvdmlkZXMgdmFsaWQgY3JlZGVudGlhbHMvcmVnaXN0cmF0aW9uIGRhdGFcbiAgczMgLS0-IHM1IDogVXNlciBwcm92aWRlcyBpbnZhbGlkIGNyZWRlbnRpYWxzL3JlZ2lzdHJhdGlvbiBkYXRhXG4gIHM1IC0tPiBzMyA6IFVzZXIgaXMgcmVkaXJlY3RlZCB0byBMb2dpbi9SZWdpc3RyYXRpb24gVUkgVVJMXG4gIHM0IC0tPiBFcnJvciA6IEEgam9iIGZhaWxzXG4gIHM0IC0tPiBzNlxuICBzNiAtLT4gWypdXG5cbiAgRXJyb3IgLS0-IFsqXVxuXG5cbiIsIm1lcm1haWQiOnsidGhlbWUiOiJkZWZhdWx0In0sInVwZGF0ZUVkaXRvciI6ZmFsc2V9)

1. The **Login/Registration User Flow** is initiated because a link was clicked
   or an action was performed that requires an active user session.
1. ORY Kratos executes Jobs defined in the **Before Login/Registration
   Workflow**. If a failure occurs, the whole flow is aborted.
1. The user's browser is redirected to
   `https://example.org/.ory/kratos/public/auth/browser/(login|registration)`
   (the notation `(login|registration)` expresses the two possibilities of
   `../auth/browser/login` or `../auth/browser/registration`).
1. ORY Kratos does some internal processing (e.g. checks if a session cookie is
   set, generates payloads for form fields, sets CSRF token, ...) and redirects
   the user's browser to the Login UI URL which is defined using the
   `urls.login_ui` (or `urls.registration_ui`) config or `URLS_LOGIN_UI` (or
   `URLS_REGISTRATION_UI`) environment variable, which is set to the ui
   endpoints - for example `https://example.org/auth/login` and
   `https://example.org/auth/registration`). The user's browser is thus
   redirected to `https://example.org/auth/(login|registration)?request=abcde`.
   The `request` query parameter includes a unique ID which will be used to
   fetch contextual data for this login request.
1. Your Server-Side Application makes a `GET` request to
   `https://ory-kratos-admin.example-org.vpc/auth/browser/requests/(login|registration)?request=abcde`.
   ORY Kratos responds with a JSON Payload that contains data (form fields,
   error messages, ...) for all enabled User Login Strategies:
   `json5 { "id": "abcde", "expires_at": "2020-01-27T09:34:39.3249566Z", "issued_at": "2020-01-27T09:24:39.3249689Z", "request_url": "https://example.org/.ory/kratos/public/auth/browser/(login|registration)", "methods": { "oidc": { "method": "oidc", "config": { /* ... */ } }, "password": { "method": "password", "config": { /* ... */ } } // ... } }`
1. Your Server-Side applications renders the data however you see fit. The User
   interacts with it an completes the Login by clicking, for example, the
   "Login", the "Login with Google", ... button.
1. The User's browser makes a request to one of ORY Kratos' Strategy URLs (e.g.
   `https://example.org/.ory/kratos/public/auth/browser/methods/password/(login|registration)`
   or
   `https://example.org/.ory/kratos/public/auth/browser/methods/oidc/auth/abcde`).
   ORY Kratos validates the User's credentials (when logging in - e.g. Username
   and Password, by performing an OpenID Connect flow, ...) or the registration
   form data (when signing up - e.g. is the E-Mail address valid, is the person
   at least 21 years old, ...):
   - If the credentials / form data is invalid, the Login Request's JSON Payload
     is updated - for example with
     ```json5
     {
       id: 'abcde',
       expires_at: '2020-01-27T10:05:50.1678228Z',
       issued_at: '2020-01-27T09:55:50.1678348Z',
       request_url: 'http://127.0.0.1:4455/auth/browser/(login|registration)',
       methods: {
         oidc: {
           method: 'oidc',
           config: {
             /* ... */
           },
         },
         password: {
           method: 'password',
           config: {
             /* ... */
             errors: [
               {
                 message: 'The provided credentials are invalid. Check for spelling mistakes in your password or username, email address, or phone number.',
               },
             ],
           },
         },
       },
     }
     ```
     and the user's Browser is redirected back to the Login UI:
     `https://example.org/auth/(login|registration)?request=abcde`.
   - If credentials / data is valid, ORY Kratos proceeds with the next step.
   - If the flow is a registration request and the registration data is valid,
     the identity is created.
1. ORY Kratos executes Jobs (e.g. redirect somewhere) defined in the **After
   Login/Registration Workflow**. If a failure occurs, the whole flow is
   aborted.

[![User Login Sequence Diagram for Server-Side Applications](https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gIHBhcnRpY2lwYW50IEIgYXMgQnJvd3NlclxuICBwYXJ0aWNpcGFudCBPIGFzIE9SWSBPYXRoa2VlcGVyXG4gIHBhcnRpY2lwYW50IEtQIGFzIE9SWSBLcmF0b3MgUHVibGljIEFQSVxuICBwYXJ0aWNpcGFudCBBIGFzIFlvdXIgU2VydmVyLVNpZGUgQXBwbGljYXRpb25cbiAgcGFydGljaXBhbnQgS0EgYXMgT1JZIEtyYXRvcyBBZG1pbiBBUElcblxuICBCLT4-K086IEdFVCAvLm9yeS9rcmF0b3MvcHVibGljL2F1dGgvYnJvd3Nlci8obG9naW58cmVnaXN0cmF0aW9uKVxuICBPLT4-K0tQOiBHRVQgL2F1dGgvYnJvd3Nlci8obG9naW58cmVnaXN0cmF0aW9uKVxuICBLUC0tPj5LUDogRXhlY3V0ZSBKb2JzIGRlZmluZWQgaW4gXCJCZWZvcmUgTG9naW4vUmVnaXN0cmF0aW9uIFdvcmtmbG93KHMpXCJcbiAgS1AtLT4-LU86IEhUVFAgMzAyIEZvdW5kIC9hcHAvYXV0aC8obG9naW58cmVnaXN0cmF0aW9uKT9yZXF1ZXN0PWFiY2RlXG4gIE8tLT4-LUI6IEhUVFAgMzAyIEZvdW5kIC9hcHAvYXV0aC8obG9naW58cmVnaXN0cmF0aW9uKT9yZXF1ZXN0PWFiY2RlXG5cbiAgQi0-PitPOiBHRVQgL2FwcC9hdXRoLyhsb2dpbnxyZWdpc3RyYXRpb24pP3JlcXVlc3Q9YWJjZGVcbiAgTy0-PitBOiBHRVQgL2F1dGgvKGxvZ2lufHJlZ2lzdHJhdGlvbik_cmVxdWVzdD1hYmNkZVxuICBBLT4-K0tBOiBHRVQgL2F1dGgvYnJvd3Nlci9yZXF1ZXN0cy8obG9naW58cmVnaXN0cmF0aW9uKT9yZXF1ZXN0PWFiY2RlXG4gIEtBLT4-LUE6IFNlbmRzIExvZ2luL1JlZ2lzdHJhdGlvbiBSZXF1ZXN0IEpTT04gUGF5bG9hZFxuICBOb3RlIG92ZXIgQSxLQTogIHtcIm1ldGhvZHNcIjp7XCJwYXNzd29yZFwiOi4uLixcIm9pZGNcIjouLn19XG4gIEEtLT4-QTogR2VuZXJhdGUgYW5kIHJlbmRlciBIVE1MXG4gIEEtLT4-LU86IFJldHVybiBIVE1MIChGb3JtLCAuLi4pXG4gIE8tLT4-LUI6IFJldHVybiBIVE1MIChGb3JtLCAuLi4pXG5cbiAgQi0tPj5COiBGaWxsIG91dCBIVE1MXG5cbiAgQi0-PitPOiBQT1NUIEhUTUwgRm9ybVxuICBPLT4-K0tQOiBQT1NUIEhUTUwgRm9ybVxuICBLUC0tPj5LUDogQ2hlY2tzIGxvZ2luIC8gcmVnaXN0cmF0aW9uIGRhdGFcblxuXG4gIGFsdCBMb2dpbiBkYXRhIGlzIHZhbGlkXG4gICAgS1AtLT4-LUtQOiBFeGVjdXRlIEpvYnMgZGVmaW5lZCBpbiBcIkFmdGVyIExvZ2luIFdvcmtmbG93KHMpXCJcbiAgICBLUC0tPj5POiBIVFRQIDMwMiBGb3VuZCAvYXBwL2Rhc2hib2FyZFxuICAgIE5vdGUgb3ZlciBLUCxCOiBTZXQtQ29va2llOiBhdXRoX3Nlc3Npb249Li4uXG4gICAgTy0tPj4tQjogSFRUUCAzMDIgRm91bmQgL2FwcC9kYXNoYm9hcmRcbiAgICBCLT4-TzogR0VUIC9hcHAvZGFzaGJvYXJkXG4gICAgTy0tPktBOiBWYWxpZGF0ZXMgU2Vzc2lvbiBDb29raWVcbiAgICBPLT4-QTogR0VUIC9kYXNoYm9hcmRcbiAgZWxzZSBMb2dpbiBkYXRhIGlzIGludmFsaWRcbiAgICBOb3RlIG92ZXIgS1AsQjogVXNlciByZXRyaWVzIGxvZ2luIC8gcmVnaXN0cmF0aW9uXG4gICAgS1AtLT4-TzogSFRUUCAzMDIgRm91bmQgL2FwcC9hdXRoLyhsb2dpbnxyZWdpc3RyYXRpb24pP3JlcXVlc3Q9YWJjZGVcbiAgICBPLS0-PkI6IEhUVFAgMzAyIEZvdW5kIC9hcHAvYXV0aC8obG9naW58cmVnaXN0cmF0aW9uKT9yZXF1ZXN0PWFiY2RlXG4gIGVuZFxuICAiLCJtZXJtYWlkIjp7InRoZW1lIjoibmV1dHJhbCIsInNlcXVlbmNlRGlhZ3JhbSI6eyJkaWFncmFtTWFyZ2luWCI6MTUsImRpYWdyYW1NYXJnaW5ZIjoxNSwiYm94VGV4dE1hcmdpbiI6MSwibm90ZU1hcmdpbiI6MTAsIm1lc3NhZ2VNYXJnaW4iOjU1LCJtaXJyb3JBY3RvcnMiOnRydWV9fSwidXBkYXRlRWRpdG9yIjpmYWxzZX0)](https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gIHBhcnRpY2lwYW50IEIgYXMgQnJvd3NlclxuICBwYXJ0aWNpcGFudCBPIGFzIE9SWSBPYXRoa2VlcGVyXG4gIHBhcnRpY2lwYW50IEtQIGFzIE9SWSBLcmF0b3MgUHVibGljIEFQSVxuICBwYXJ0aWNpcGFudCBBIGFzIFlvdXIgU2VydmVyLVNpZGUgQXBwbGljYXRpb25cbiAgcGFydGljaXBhbnQgS0EgYXMgT1JZIEtyYXRvcyBBZG1pbiBBUElcblxuICBCLT4-K086IEdFVCAvLm9yeS9rcmF0b3MvcHVibGljL2F1dGgvYnJvd3Nlci8obG9naW58cmVnaXN0cmF0aW9uKVxuICBPLT4-K0tQOiBHRVQgL2F1dGgvYnJvd3Nlci8obG9naW58cmVnaXN0cmF0aW9uKVxuICBLUC0tPj5LUDogRXhlY3V0ZSBKb2JzIGRlZmluZWQgaW4gXCJCZWZvcmUgTG9naW4vUmVnaXN0cmF0aW9uIFdvcmtmbG93KHMpXCJcbiAgS1AtLT4-LU86IEhUVFAgMzAyIEZvdW5kIC9hcHAvYXV0aC8obG9naW58cmVnaXN0cmF0aW9uKT9yZXF1ZXN0PWFiY2RlXG4gIE8tLT4-LUI6IEhUVFAgMzAyIEZvdW5kIC9hcHAvYXV0aC8obG9naW58cmVnaXN0cmF0aW9uKT9yZXF1ZXN0PWFiY2RlXG5cbiAgQi0-PitPOiBHRVQgL2FwcC9hdXRoLyhsb2dpbnxyZWdpc3RyYXRpb24pP3JlcXVlc3Q9YWJjZGVcbiAgTy0-PitBOiBHRVQgL2F1dGgvKGxvZ2lufHJlZ2lzdHJhdGlvbik_cmVxdWVzdD1hYmNkZVxuICBBLT4-K0tBOiBHRVQgL2F1dGgvYnJvd3Nlci9yZXF1ZXN0cy8obG9naW58cmVnaXN0cmF0aW9uKT9yZXF1ZXN0PWFiY2RlXG4gIEtBLT4-LUE6IFNlbmRzIExvZ2luL1JlZ2lzdHJhdGlvbiBSZXF1ZXN0IEpTT04gUGF5bG9hZFxuICBOb3RlIG92ZXIgQSxLQTogIHtcIm1ldGhvZHNcIjp7XCJwYXNzd29yZFwiOi4uLixcIm9pZGNcIjouLn19XG4gIEEtLT4-QTogR2VuZXJhdGUgYW5kIHJlbmRlciBIVE1MXG4gIEEtLT4-LU86IFJldHVybiBIVE1MIChGb3JtLCAuLi4pXG4gIE8tLT4-LUI6IFJldHVybiBIVE1MIChGb3JtLCAuLi4pXG5cbiAgQi0tPj5COiBGaWxsIG91dCBIVE1MXG5cbiAgQi0-PitPOiBQT1NUIEhUTUwgRm9ybVxuICBPLT4-K0tQOiBQT1NUIEhUTUwgRm9ybVxuICBLUC0tPj5LUDogQ2hlY2tzIGxvZ2luIC8gcmVnaXN0cmF0aW9uIGRhdGFcblxuXG4gIGFsdCBMb2dpbiBkYXRhIGlzIHZhbGlkXG4gICAgS1AtLT4-LUtQOiBFeGVjdXRlIEpvYnMgZGVmaW5lZCBpbiBcIkFmdGVyIExvZ2luIFdvcmtmbG93KHMpXCJcbiAgICBLUC0tPj5POiBIVFRQIDMwMiBGb3VuZCAvYXBwL2Rhc2hib2FyZFxuICAgIE5vdGUgb3ZlciBLUCxCOiBTZXQtQ29va2llOiBhdXRoX3Nlc3Npb249Li4uXG4gICAgTy0tPj4tQjogSFRUUCAzMDIgRm91bmQgL2FwcC9kYXNoYm9hcmRcbiAgICBCLT4-TzogR0VUIC9hcHAvZGFzaGJvYXJkXG4gICAgTy0tPktBOiBWYWxpZGF0ZXMgU2Vzc2lvbiBDb29raWVcbiAgICBPLT4-QTogR0VUIC9kYXNoYm9hcmRcbiAgZWxzZSBMb2dpbiBkYXRhIGlzIGludmFsaWRcbiAgICBOb3RlIG92ZXIgS1AsQjogVXNlciByZXRyaWVzIGxvZ2luIC8gcmVnaXN0cmF0aW9uXG4gICAgS1AtLT4-TzogSFRUUCAzMDIgRm91bmQgL2FwcC9hdXRoLyhsb2dpbnxyZWdpc3RyYXRpb24pP3JlcXVlc3Q9YWJjZGVcbiAgICBPLS0-PkI6IEhUVFAgMzAyIEZvdW5kIC9hcHAvYXV0aC8obG9naW58cmVnaXN0cmF0aW9uKT9yZXF1ZXN0PWFiY2RlXG4gIGVuZFxuICAiLCJtZXJtYWlkIjp7InRoZW1lIjoibmV1dHJhbCIsInNlcXVlbmNlRGlhZ3JhbSI6eyJkaWFncmFtTWFyZ2luWCI6MTUsImRpYWdyYW1NYXJnaW5ZIjoxNSwiYm94VGV4dE1hcmdpbiI6MSwibm90ZU1hcmdpbiI6MTAsIm1lc3NhZ2VNYXJnaW4iOjU1LCJtaXJyb3JBY3RvcnMiOnRydWV9fSwidXBkYXRlRWRpdG9yIjpmYWxzZX0)

### Client-Side Browser Applications

Because Client-Side Browser Applications do not have access to ORY Kratos' Admin
API, they must use the ORY Kratos Public API instead. The flow for a Client-Side
Browser Application is almost the exact same as the one for Server-Side
Applications, with the small difference that
`https://example.org/.ory/kratos/public/auth/browser/requests/login?request=abcde`
would be called via AJAX instead of making a request to
`https://ory-kratos-admin.example-org.vpc/auth/browser/requests/login?request=abcde`.

> To prevent brute force, guessing, session injection, and other attacks, it is
> required that cookies are working for this endpoint. The cookie set in the
> initial HTTP request made to
> `https://example.org/.ory/kratos/public/auth/browser/login` MUST be set and
> available when calling this endpoint!

## Self-Service User Login and User Registration for API Clients

Will be addressed in a future release.

## Executing Jobs before User Login

ORY Kratos allows you to configure jobs that run before the Login Request is
generated. This may be helpful if you'd like to restrict logins to IPs coming
from your internal network or other logic.

You can find available `before` jobs in
[Self-Service Before Login Jobs](../workflows/jobs/before.md#user-login) and
configure them using the ORY Kratos configuration file:

```yaml
selfservice:
  login:
    before:
      - run: <job-name>
        config:
          # <job-config>
```

## Executing Jobs after User Login

ORY Kratos allows you to configure jobs that run before the Login Request is
generated. This may be helpful if you'd like to restrict logins to IPs coming
from your internal network or other logic.

You can find available `after` jobs in
[Self-Service After Login Jobs](../workflows/jobs/after.md#user-login) and
configure them using the ORY Kratos configuration file:

```yaml
selfservice:
  after:
    <strategy>:
      - run: <job-name>
        config:
          # <job-config>
```

It's possible to define jobs running after login for each individual User Login
Flow Strategy (e.g. `password`, `oidc`).

## Executing Jobs before User Registration

ORY Kratos allows you to configure jobs that run before the Registration Request
is generated. This may be helpful if you'd like to restrict registrations to IPs
coming from your internal network or other logic.

You can find available `before` jobs in
[Self-Service Before Registration Jobs](../workflows/jobs/before.md#user-registration)
and configure them using the ORY Kratos configuration file:

```yaml
selfservice:
  registration:
    before:
      - run: <job-name>
        config:
          # <job-config>
```

## Executing Jobs after User Registration

ORY Kratos allows you to configure jobs that run before the Registration Request
is generated. This may be helpful if you'd like to restrict registrations to IPs
coming from your internal network or other logic.

You can find available `after` jobs in
[Self-Service After Registration Jobs](../workflows/jobs/after.md#user-registration)
and configure them using the ORY Kratos configuration file:

```yaml
selfservice:
  after:
    <strategy>:
      - run: <job-name>
        config:
          # <job-config>
```

It's possible to define jobs running after registration for each individual User
Registration Flow Strategy (e.g. `password`, `oidc`).
