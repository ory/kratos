---
id: index
title: Overview
---

ORY Kratos allows end-users to sign up, log in, log out, update profile
information, recover accounts, and perform other important account management
tasks without third party involvement ("self-service").

The opposite to self-service management is someone using administrative
privileges to create, update, or delete accounts.

## Network Flows for Browsers

All Self-Service Flows such as [User Login and User Registration](user-login-user-registration.mdx),
[Profile Management](user-settings-profile-management.mdx) use the same template:

1. The Browser makes an HTTP request to the flow's initialization endpoint (e.g.
   `/auth/browser/login`);
2. The initialization endpoint processes data and associates it with a request
   ID and redirects the browser to the flow's configured UI URL (e.g.
   `urls.login_ui`), appending the request ID as the `request` URL Query
   Parameter;
3. The endpoint responsible for the UI URL uses the `request` URL Query
   Parameter (e.g. `http://my-app/auth/login?request=abcde`) to fetch the data
   previously associated with the Request ID from either ORY Kratos's Public or
   Admin API.
4. The UI endpoint renders the fetched data in any way it sees it fit. The flow
   is typically completed by the browser making another request to one of ORY
   Kratos' endpoints, which is usually described in the fetched request data.
