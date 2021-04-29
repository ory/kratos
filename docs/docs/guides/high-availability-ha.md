---
id: high-availability-ha
title: High Availability
---

This document explains how to set up Ory Kratos in for High Availability.

# Horizontal Scaling

Ory Kratos scales effortlessly to thousands of pods without any additional work!

# Mail courier

Ory Kratos processes emails by storing them in an email queue on your database
and running a mail courier worker to handle this queue. To avoid processing the
same email multiple times, only one instance of this mail courier should be run
at one time. For simple single-instance Kratos deployments, the courier can
simply be run as a background worker, but for multi-instance Kratos deployments,
it needs to be run a distinct singleton foreground worker. For setup details,
refer to the [Out of Band Communication guide](../concepts/email-sms.md).

Ory Kratos does not have any special requirements when it comes to High
Availability as it does not manage state itself but instead relies on the SQL
database for that.

It is therefore possible to use Ory Kratos with Auto-Scaling Groups (e.g. in
Kubernetes) without any additional configuration.
