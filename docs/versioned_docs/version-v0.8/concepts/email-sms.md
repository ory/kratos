---
id: email-sms
title: Out-of-band communication via E-Mail and SMS
---

Ory Kratos sends out-of-band messages via SMS or E-Mail. The following exemplary use cases require these messages:


- Send an account activation email
- Verify an E-Mail address or mobile phone number using SMS
- Preventing Account Enumeration Attacks
- Sending a 2FA Codes
- ...

## Mail courier

Ory Kratos processes email dispatch using a mail courier worker, which must run
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

To have E-Mail delivery running with Ory Kratos requires an SMTP server. This is
set up in the configuration file using an absolute URL with the `smtp` or
`smtps` scheme:

```yaml title="path/to/my/kratos/config.yml"
# $ kratos -c path/to/my/kratos/config.yml serve
courier:
  smtp:
    connection_uri: smtps://foo:bar@my-smtp-server:1234/
    # Examples:
    # - "smtp://foo:bar@my-mailserver:1234/?disable_starttls=true
    #   (NOT RECOMMENDED: Cleartext smtp for devel and legacy infrastructure
    #   only)"
    # - smtp://foo:bar@my-mailserver:1234/ (Explicit StartTLS with certificate
    #   trust verification)
    # - "smtp://foo:bar@my-mailserver:1234/?skip_ssl_verify=true (NOT
    #   RECOMMENDED: Explicit StartTLS without certificate trust verification)"
    # - smtps://foo:bar@my-mailserver:1234/ (Implicit TLS with certificate trust
    #   verification)
    # - "smtps://foo:bar@my-mailserver:1234/?skip_ssl_verify=true (NOT
    #   RECOMMENDED: Implicit TLS without certificate trust verification)"
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
  smtp:
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

Ory Kratos comes with built-in templates. If you wish to define your own, custom
templates, you should define `template_override_path`, as shown above, to
indicate where your custom templates are located. This will become the
`<template-root>` for your custom templates, as indicated below.

`email.subject.gotmpl`, `email.body.gotmpl` and `email.body.plaintext.gotmpl`
are common template file names expected in the sub directories of the root
directory, corresponding to the respective methods for filling e-mail subject
and body.

> Templates use the golang template engine in the `text/template` package for
> rendering the `email.subject.gotmpl` and `email.body.plaintext.gotmpl`
> templates, and the `html/template` package for rendering the
> `email.body.gotmpl` template: https://pkg.go.dev/text/template >
> https://pkg.go.dev/html/template
>
> Templates can use the [Sprig](https://github.com/Masterminds/sprig) library,
> which provides more than 100 commonly used template functions:
> http://masterminds.github.io/sprig/

- **recovery**: recovery email templates directory, expected to be located in
  `<root_directory>/recovery`
  - valid: sub directory, expected to be located in
    `<template-root>/recovery/valid`, containing templates with variables `To`,
    `RecoveryURL` and `Identity` for validating a recovery
  - invalid: sub directory, expected to be located in
    `<template-root>/recovery/invalid`, containing templates with variables `To`
    for invalidating a recovery
- **verification**: verification email templates directory, expected to be
  located in `<root_directory>/verification`
  - valid: sub directory, expected to be located in
    `<template-root>/verification/valid`, containing templates with variables
    `To`, `VerificationURL` and `Identity` for validating a verification
  - invalid: sub directory, expected to be located in
    `<template-root>/verification/invalid`, containing templates with variables
    `To` for invalidating a verification

For example:
[`https://github.com/ory/kratos/blob/master/courier/template/courier/builtin/templates/verification/valid/email.body.gotmpl`](https://github.com/ory/kratos/blob/master/courier/template/courier/builtin/templates/verification/valid/email.body.gotmpl)

```gotmpl title="courier/template/templates/verification/valid/email.body.gotmpl"
Hi, please verify your account by clicking the following link:

<a href="{{ .VerificationURL }}">{{ .VerificationURL }}</a>
```

```gotmp title="courier/template/templates/verification/valid/email.body.plaintext.gotmpl"
Hi, please verify your account by clicking the following link: {{ .VerificationURL }}
```

### The Identity attribute

To be able to customize the content of templates based on the identity of the
recipient of the e-mail, the identity has been made available as `Identity`.
This object is a map containing all the attributes of an identity, such as `id`,
`state`, `recovery_addresses`, `verifiable_addresses` and `traits`.

### Nested templates

You can use nested templates to render `email.subject.gotmpl`,
`email.body.gotmpl` and `email.body.plaintext.gotmpl` templates.

#### Example: i18n customization

Using nested templates, you can either use in-line template definitions, or as
in this example, use separate templates. In this example, we will define the
email body for recovery e-mails. Assuming that we have an attribute named `lang`
that contains the required language in the `traits` of the identity, we can
define our templates as indicated below.

```txt file="<template-root>/recovery/valid/email.body.gotmpl"

{{- if eq .Identity.traits.language "de" -}}
{{ template "email.body.de.gotmpl" . }}
{{- else -}}
{{ template "email.body.en.gotmpl" . }}
{{- end -}}
<a href="{{ .RecoveryURL }}">{{.RecoveryURL }}</a>
```

```txt file="<template-root>/recovery/valid/email.body.de.gotmpl"

Hallo {{ upper .Identity.traits.firstName }},

Um Ihr Konto wiederherzustellen, klicken Sie bitte auf den folgenden Link:
```

```txt file="<template-root>/recovery/valid/email.body.en.gotmpl"


Hello {{ upper .Identity.traits.firstName }},

to recover your account, please click on the link below:
```

As indicated by the example, we need a root template, which is the
`email.body.gotmpl` template, and then we define sub templates that conform to
the following pattern: `email.body*`. You can also see that the `Identity` of
the user is available in all templates, and that you can use Sprig functions
also in the nested templates.

### Custom Headers

You can configure custom SMTP headers. For example, if integrating with AWS SES
SMTP interface, the headers can be configured for cross-account sending:

```yaml title="path/to/my/kratos/config.yml"
# $ kratos -c path/to/my/kratos/config.yml serve
courier:
  smtp:
    headers:
      X-SES-SOURCE-ARN: arn:aws:ses:us-west-2:123456789012:identity/example.com
      X-SES-FROM-ARN: arn:aws:ses:us-west-2:123456789012:identity/example.com
      X-SES-RETURN-PATH-ARN: arn:aws:ses:us-west-2:123456789012:identity/example.com
```

## Sending SMS

The Sending SMS feature is not supported at present. It will be available in a
future version of Ory Kratos.
