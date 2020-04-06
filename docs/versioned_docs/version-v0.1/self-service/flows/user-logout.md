---
id: user-logout
title: User Logout
---

ORY Kratos supports two logout flows:

- Browser-based (easy): This flow works for all applications running on top of a
  browser. Websites, single-page apps, Cordova/Ionic, and so on.
- API-based (advanced): This flow works for native applications like iOS
  (Swift), Android (Java), Microsoft (.NET), React Native, Electron, and others.

## Self-Service User Logout for Browser Applications

> WARNING - This flow is currently vulnerable to CSRF attacks because anyone can
> direct your users to the logout endpoint. A future release of ORY Kratos will
> use POST Forms with Anti-CSRF Tokens to prevent this problem. This is tracked
> as [kratos#142](https://github.com/ory/kratos/issues/142).

To log a user out, all you have to do is to direct the browser to
`http://ory-kratos-public/auth/browser/logout`. After successful logout, the
browser will be redirected to the `redirect_to` value set in ORY Krato's
configuration file:

```
selfservice:
  logout:
    redirect_to: http://test.kratos.ory.sh:4000/
```

## Self-Service User Logout for API Clients

This will be addressed in a future release of ORY Kratos.
