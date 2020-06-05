---
id: ui-user-interface
title: User Interface
---

ORY Kratos has no user interface included. Instead, it defines HTTP flows and
APIs that make it simple to write your own UI in a variety of languages and
frameworks.

The following two examples are typical UIs used in connection with ORY Kratos.

## Administrative User Interface (Admin UI)

The AUI might show all of the identities in the system and provide features to
administrators such as editing profiles, resetting passwords, and so on.

At present, there is no Open Source AUI for ORY Kratos.

## Self-service User Interface (SSUI)

The SSUI shows screens such as "login", "Registration", "Update your profile",
"Recover access to your account", and others. The following provides more reference for
SSUI at
[github.com/ory/kratos-selfservice-ui-node](https://github.com/ory/kratos-selfservice-ui-node).

The SSUI can be built in any programming language including Java, Node, or
Python and can be run both a server or a end-user device for example a browser,
or a mobile phone. Implementing a SSUI is simple and straight forward. There is
no complex authentication mechanism required and no need to worry about possible
attack vectors such as CSRF or Session Attacks since ORY Kratos provides the
preventive measures built in.

Chapter [Self-Service Flows](../self-service/flows/index) contains further
information on APIs and flows related to the SSUI, and build self service
applications.

## Messages

This section is a work-in-progress.
