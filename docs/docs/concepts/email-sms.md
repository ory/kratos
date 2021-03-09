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

## Mail courier

ORY Kratos processes email dispatch using a mail courier worker, which must run
as a singleton in order to process the mail queue correctly. It can be run as a
background worker on a single-instance Kratos setup or as a distinct singleton
foreground worker in multi-instance deployments.

### Single instance

To run the mail courier in the background on your single Kratos instance, add
the `--watch-courier` flag to your `kratos serve` command, as outlined in the
[CLI docs](../cli/kratos-serve.md)

### Multi-instance

If you're running multiple instances of Kratos (eg replicated Kubernetes
deployment), you need to run the mail courier as a separate singleton job. The
courier can be started with the `kratos courier watch` command
([CLI docs](../cli/kratos-courier.md)).

## Sending E-Mails via SMTP

To have E-Mail delivery running with ORY Kratos requires an SMTP server. This is
set up in the configuration file using an absolute URL with the `smtp` schema:

```yaml title="path/to/my/kratos/config.yml"
# $ kratos -c path/to/my/kratos/config.yml serve
courier:
  smtp:
    connection_uri: smtps://test:test@my-smtp-server:1025/
```

### Sender Address and Template Customization

You can customize the sender address and email templates.

```yaml title="path/to/my/kratos/config.yml"
# $ kratos -c path/to/my/kratos/config.yml serve
courier:
  ## SMTP Sender Address ##
  #
  # The recipient of an email will see this as the sender address.
  #
  # Default value: no-reply@ory.kratos.sh
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export COURIER_SMTP_FROM_ADDRESS=<value>
  # - Windows Command Line (CMD):
  #    > set COURIER_SMTP_FROM_ADDRESS=<value>
  #
  from_address: no-reply@ory.kratos.sh
  ## Override message templates ##
  #
  # You can override certain or all message templates by pointing this key to the path where the templates are located.
  #
  # Examples:
  # - /conf/courier-templates
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export COURIER_TEMPLATE_OVERRIDE_PATH=<value>
  # - Windows Command Line (CMD):
  #    > set COURIER_TEMPLATE_OVERRIDE_PATH=<value>
  #
  template_override_path: /conf/courier-templates
```

`email.subject.gotmpl` and `email.body.gotmpl` are common template file names
expected in remainder directories corresponding to respective methods for
filling E-mail subject and body.

> Templates use the engine golang text template for text/html email rendering:
> https://golang.org/pkg/text/template

- recovery: recovery email templates root directory
  - valid: sub directory containing templates with variables `To` and
    `VerificationURL` for validating a recovery
  - invalid: sub directory containing templates with variables `To` for
    invalidating a recovery
- verification: verification email templates root directory
  - valid: sub directory containing templates with variables `To` and
    `RecoveryURL` for validating a verification
  - invalid: sub directory containing templates with variables `To` for
    invalidating a verification

For example:
[`/courier/template/courier/builtin/templates/verification/valid/email.body.gotmpl`](https://github.com/ory/kratos/blob/master/courier/template/templates/verification/valid/email.body.gotmpl)

```gotmpl title="courier/template/templates/verification/valid/email.body.gotmpl"
Hi, please verify your account by clicking the following link:

<a href="{{ .VerificationURL }}">{{ .VerificationURL }}</a>
```

## Sending SMS

The Sending SMS feature is not supported at present. It will be available in a
future version of ORY Kratos.
