---
id: user-logout
title: User Logout
---

Ory Kratos supports two logout flows:

- Browser-based (easy): This flow works for all applications running on top of a
  browser. Websites, single-page apps, Cordova/Ionic, and so on.
- API-based (advanced): This flow works for native applications like iOS
  (Swift), Android (Java), Microsoft (.NET), React Native, Electron, and others.

## Self-Service User Logout for Server-Side Browser Applications

After successful logout, the browser will be redirected either to the `return_to` query parameter
from the initial request URL, or fall back to the `default_browser_return_url`
value set in Ory Kratos' configuration file:

```yaml
# kratos.yaml
selfservice:
  flows:
    logout:
      after:
        default_browser_return_url: http://test.kratos.ory.sh:4000/
```



## Self-Service User Logout for API Clients

This will be addressed in a future release of Ory Kratos.
