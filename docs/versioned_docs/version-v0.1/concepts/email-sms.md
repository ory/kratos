---
id: email-sms
title: Out-of-band communication via E-Mail and SMS
---

ORY Kratos sends out-of-band messages via SMS or E-Mail. These messages are
required for The following exemplary use cases require these messages:

- Send an account activation email
- Verify an E-Mail address or mobile phone number using SMS
- Preventing Account Enumeration Attacks
- Sending a 2FA Codes
- ...

## Sending E-Mails via SMTP

To have E-Mail delivery running with ORY Kratos requires an SMTP server. This is
set up in the configuration file using an absolute URL with the `smtp` schema:

```yaml
courier:
  smtp:
    connection_uri: smtp://test:test@my-smtp-server:1025/
```

### Templates

A future version of ORY Kratos will feature proprietary E-Mail messages using
the Go template engine extended by
[sprig's template functions](http://masterminds.github.io/sprig/). It should
also be possible to internationalize these templates.

## Sending SMS

The Sending SMS feature is not supported at present. It will be available in a
future version of ORY Kratos.
