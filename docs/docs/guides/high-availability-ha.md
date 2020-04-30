---
id: high-availability-ha
title: High Availability
---

ORY Kratos does not have any special requirements when it comes to High Availability as it does not
manage state itself but instead relies on the SQL database for that.

It is therefore possible to use ORY Kratos with Auto-Scaling Groups (e.g. in Kubernetes) without
any additional configuration.
