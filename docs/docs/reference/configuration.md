---
id: configuration
title: Configuration
---

<!-- THIS FILE IS BEING AUTO-GENERATED. DO NOT MODIFY IT AS ALL CHANGES WILL BE OVERWRITTEN.
OPEN AN ISSUE IF YOU WOULD LIKE TO MAKE ADJUSTMENTS HERE AND MAINTAINERS WILL HELP YOU LOCATE THE RIGHT
FILE -->

If file `$HOME/.kratos.yaml` exists, it will be used as a configuration file
which supports all configuration settings listed below.

You can load the config file from another source using the
`-c path/to/config.yaml` or `--config path/to/config.yaml` flag:
`kratos --config path/to/config.yaml`.

Config files can be formatted as JSON, YAML and TOML. Some configuration values
support reloading without server restart. All configuration values can be set
using environment variables, as documented below.

To find out more about edge cases like setting string array values through
environmental variables head to the
[Configuring ORY services](https://www.ory.sh/docs/ecosystem/configuring)
section.

```yaml
## ORY Kratos Configuration
#

## Data Source Name ##
#
# DSN is used to specify the database credentials as a connection URI.
#
# Examples:
# - "postgres://user:
#   password@postgresd:5432/database?sslmode=disable&max_conns=20&max_idle_conns=\
#   4"
# - mysql://user:secret@tcp(mysqld:3306)/database?max_conns=20&max_idle_conns=4
# - cockroach://user@cockroachdb:26257/database?sslmode=disable&max_conns=20&max_idle_conns=4
# - sqlite:///var/lib/sqlite/db.sqlite?_fk=true&mode=rwc
#
# Set this value using environment variables on
# - Linux/macOS:
#    $ export DSN=<value>
# - Windows Command Line (CMD):
#    > set DSN=<value>
#
dsn: sqlite:///var/lib/sqlite/db.sqlite?_fk=true&mode=rwc

## identity ##
#
identity:
  ## traits ##
  #
  traits:
    ## default_schema_url ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export IDENTITY_TRAITS_DEFAULT_SCHEMA_URL=<value>
    # - Windows Command Line (CMD):
    #    > set IDENTITY_TRAITS_DEFAULT_SCHEMA_URL=<value>
    #
    default_schema_url: https://cxE.zqrdEhnC2EWi

    ## schemas ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export IDENTITY_TRAITS_SCHEMAS=<value>
    # - Windows Command Line (CMD):
    #    > set IDENTITY_TRAITS_SCHEMAS=<value>
    #
    schemas:
      - 46430357.842930496
      - sunt
      - true
      - 4227808
      - true

## selfservice ##
#
selfservice:
  ## logout ##
  #
  logout:
    ## redirect_to ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_LOGOUT_REDIRECT_TO=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_LOGOUT_REDIRECT_TO=<value>
    #
    redirect_to: https://yXHrkOAUMgtjShUxOmmN.aovL2qLQma6zHRVKjTWhCZ7O2vyZWn5p+56Re

  ## strategies ##
  #
  strategies:
    ## password ##
    #
    password:
      ## enabled ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_STRATEGIES_PASSWORD_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_STRATEGIES_PASSWORD_ENABLED=<value>
      #
      enabled: true

    ## oidc ##
    #
    oidc:
      ## enabled ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_STRATEGIES_OIDC_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_STRATEGIES_OIDC_ENABLED=<value>
      #
      enabled: true

      ## config ##
      #
      config:
        ## providers ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_STRATEGIES_OIDC_CONFIG_PROVIDERS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_STRATEGIES_OIDC_CONFIG_PROVIDERS=<value>
        #
        providers:
          - id: nulla sit dolor qui
            provider: generic
            client_id: sed pariatur commodo proident
            client_secret: commodo ullamco
            schema_url: https://fRZBUeVznHWjiIPhgGuEYdZ.mpfuzgF+jDV6rMUswLn+Lxnab+KICB-pnABLpuiFRQEiAnlsBlk1vzaCA
            issuer_url: https://ZWkKOZXMJQSckOKzNuoSelxRVYUsfdNB.peyxBDnQ3xnsqKyQ05vRHlrrq4mGIt-
            auth_url: https://LiUzWpDPEqaXBGoIJwztEzERyPz.abjohV
            token_url: http://sJpuqVLMSYebYlfFJjLxs.snetmQs0l0pjS9LcYE8UvQZEoxPuuRf.Rp5spluMz,MmipV.i9-Is
            scope:
              - exercitation ea aute in
              - esse consectetur
              - consequat laboris
              - exercitation
              - Ut non id

  ## settings ##
  #
  settings:
    ## request_lifespan ##
    #
    # Default value: 1h
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_SETTINGS_REQUEST_LIFESPAN=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_SETTINGS_REQUEST_LIFESPAN=<value>
    #
    request_lifespan: 11986m

    ## privileged_session_max_age ##
    #
    # Default value: 1h
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_SETTINGS_PRIVILEGED_SESSION_MAX_AGE=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_SETTINGS_PRIVILEGED_SESSION_MAX_AGE=<value>
    #
    privileged_session_max_age: 1895827ns

    ## after ##
    #
    after:
      ## password ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_SETTINGS_AFTER_PASSWORD=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_SETTINGS_AFTER_PASSWORD=<value>
      #
      password:
        - job: verify
        - job: redirect
          config:
            default_redirect_url: https://bylYGWdXhPMOlgETHFLDYpAPXw.foqskCnwgBac9jTKPPUtyMQSntq3gVwffrbxTSeeYDo7
            allow_user_defined_redirect: true

      ## profile ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_SETTINGS_AFTER_PROFILE=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_SETTINGS_AFTER_PROFILE=<value>
      #
      profile:
        - job: redirect
          config:
            default_redirect_url: http://zVDjpjpHraGnnoJQvmethywVZpH.gjNOh6WlzFjO7U8l5Q6HdVeiQ5YbjWoP5
            allow_user_defined_redirect: false
        - job: redirect
          config:
            default_redirect_url: http://fjpOpoKfEyfQrv.fmjscUOm,ZHbqe2gfkwnp,i2LorPzQH0V0hNglGDL3E
            allow_user_defined_redirect: false
        - job: redirect
          config:
            default_redirect_url: https://cndbxNASYwjSzIzckckzGyqSwZKZWHlK.kyhaCm0ieSv,c
            allow_user_defined_redirect: true
        - job: verify

  ## verify ##
  #
  verify:
    ## Self-Service Verification Request Lifespan ##
    #
    # Sets how long the verification request (for the UI interaction) is valid.
    #
    # Default value: 1h
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_VERIFY_REQUEST_LIFESPAN=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_VERIFY_REQUEST_LIFESPAN=<value>
    #
    request_lifespan: 5us

    ## Self-Service Verification Link Lifespan ##
    #
    # Sets how long the verification link (e.g. the one sent via email) is valid for.
    #
    # Default value: 24h
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_VERIFY_LINK_LIFESPAN=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_VERIFY_LINK_LIFESPAN=<value>
    #
    link_lifespan: 08174629us

  ## login ##
  #
  login:
    ## request_lifespan ##
    #
    # Default value: 1h
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_LOGIN_REQUEST_LIFESPAN=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_LOGIN_REQUEST_LIFESPAN=<value>
    #
    request_lifespan: 430ns

    ## before ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_LOGIN_BEFORE=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_LOGIN_BEFORE=<value>
    #
    before:
      - job: redirect
        config:
          default_redirect_url: https://Onq.ekva66ChGJy8k7T,uVP5wLMBQ6tkiRs-KRjLf9t5Qv0Zk6jsZGzpuUGkpKybLwzZ,WrX
          allow_user_defined_redirect: false
      - job: redirect
        config:
          default_redirect_url: http://ErPLvISC.tujcyWXjGWrWaKKW1UU,g
          allow_user_defined_redirect: true

    ## after ##
    #
    after:
      ## password ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_LOGIN_AFTER_PASSWORD=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_LOGIN_AFTER_PASSWORD=<value>
      #
      password:
        - job: redirect
          config:
            default_redirect_url: https://znYBnrPWZTTbeT.sfgp0RbLJe9Eie0JzQgx,,yNL1A3.y7zNdzZeQhV1UmMh1SofCpoyRaPoFTcnmjurOUGFttPePL5.TS0
            allow_user_defined_redirect: false
        - job: revoke_active_sessions

      ## oidc ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_LOGIN_AFTER_OIDC=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_LOGIN_AFTER_OIDC=<value>
      #
      oidc:
        - job: revoke_active_sessions
        - job: redirect
          config:
            default_redirect_url: https://DiwBfLPpgautsYDHfKIseekWraYkhjok.kejpSr5uL5luHAvYXVLoK6M.N0cfJI0DzHKdrSLgAoZ8jk-HcrI-HsLvkH3wquf51gNJ8Q2G,IRERYGex5
            allow_user_defined_redirect: true

  ## registration ##
  #
  registration:
    ## request_lifespan ##
    #
    # Default value: 1h
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_REGISTRATION_REQUEST_LIFESPAN=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_REGISTRATION_REQUEST_LIFESPAN=<value>
    #
    request_lifespan: 3ns

    ## before ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_REGISTRATION_BEFORE=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_REGISTRATION_BEFORE=<value>
    #
    before:
      - job: redirect
        config:
          default_redirect_url: https://EIQPBaZyPKvfaNPn.wwzkGTlGJJe0h0nT-99U9oJE4Ii34vu9ocet,IOEiRXyMo1p6lfV8sb3MK4f4
          allow_user_defined_redirect: false
      - job: redirect
        config:
          default_redirect_url: https://cDEaBRfmdFlLaVdlCxBrqySNUjviz.enQLMHFJdAsphpap3
          allow_user_defined_redirect: false

    ## after ##
    #
    after:
      ## password ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_REGISTRATION_AFTER_PASSWORD=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_REGISTRATION_AFTER_PASSWORD=<value>
      #
      password:
        - job: redirect
          config:
            default_redirect_url: http://JhEXFvJnNbZKwIut.ywkiuYaJ06ItyYizPG,ap9JBtVqjKL1kRiOxx59oCGOIyX
            allow_user_defined_redirect: true
        - job: session
        - job: redirect
          config:
            default_redirect_url: https://olFF.qwtGKSnDp8aTp
            allow_user_defined_redirect: false

      ## oidc ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_REGISTRATION_AFTER_OIDC=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_REGISTRATION_AFTER_OIDC=<value>
      #
      oidc:
        - job: redirect
          config:
            default_redirect_url: http://qLMzanEJ.ykBDPsAnhk6
            allow_user_defined_redirect: true
        - job: redirect
          config:
            default_redirect_url: http://trNorZlje.xrHf8blytUB
            allow_user_defined_redirect: true

## Courier configuration ##
#
# The courier is responsible for sending and delivering messages over email, sms, and other means.
#
courier:
  ## SMTP Configuration ##
  #
  # Configures outgoing emails using the SMTP protocol.
  #
  smtp:
    ## SMTP connection string ##
    #
    # This URI will be used to connect to the SMTP server.
    #
    # Examples:
    # - smtps://foo:bar@my-mailserver:1234/
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export COURIER_SMTP_CONNECTION_URI=<value>
    # - Windows Command Line (CMD):
    #    > set COURIER_SMTP_CONNECTION_URI=<value>
    #
    connection_uri: smtps://foo:bar@my-mailserver:1234/

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
    from_address: Ou-d4YQ@CdlLYZDo.bt

  ## Override message templates ##
  #
  # You can override certain or all message templates by pointing this key to the path where the templates are located.
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export COURIER_TEMPLATE_OVERRIDE_PATH=<value>
  # - Windows Command Line (CMD):
  #    > set COURIER_TEMPLATE_OVERRIDE_PATH=<value>
  #
  template_override_path: mollit eiusmod

## serve ##
#
serve:
  ## admin ##
  #
  admin:
    ## host ##
    #
    # Default value: 0.0.0.0
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_ADMIN_HOST=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_ADMIN_HOST=<value>
    #
    host: pariatur Excepteur Lorem laboris aute

    ## port ##
    #
    # Default value: 4434
    #
    # Examples:
    # - 4434
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_ADMIN_PORT=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_ADMIN_PORT=<value>
    #
    port: 4434

  ## public ##
  #
  public:
    ## host ##
    #
    # Default value: 0.0.0.0
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_PUBLIC_HOST=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_PUBLIC_HOST=<value>
    #
    host: Duis enim ut cillum

    ## port ##
    #
    # Default value: 4433
    #
    # Examples:
    # - 4433
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_PUBLIC_PORT=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_PUBLIC_PORT=<value>
    #
    port: 4433

## urls ##
#
urls:
  ## settings_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_SETTINGS_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_SETTINGS_UI=<value>
  #
  settings_ui: https://bpEJKaB.azyDo86oKwN1wfbGx0JsEcK

  ## mfa_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_MFA_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_MFA_UI=<value>
  #
  mfa_ui: http://iCHtBWZjWFkl.ujwKx,UcsGzc.qDzBV4wOAlzeOMG

  ## login_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_LOGIN_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_LOGIN_UI=<value>
  #
  login_ui: http://VHkYFEOlmHxeBuKA.pjDGbjRNQElUV6,k6CgLuo0T-qDh+ZtQUAXixoPqQ.MWqFi.oerjTswpoqo5mpe

  ## registration_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_REGISTRATION_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_REGISTRATION_UI=<value>
  #
  registration_ui: http://HoErpcMluYnLMWmjqyjGgKWbJlPclt.uttdzVGlmc2VXb57Y2IX0wC-JeqTWNMyg-7YO073BK1RxMpHR1zs2b0+aNXEb68M2FxCAOIQtguXrLJerjT5exKj

  ## error_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_ERROR_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_ERROR_UI=<value>
  #
  error_ui: http://dEKpltztPgsJxLoIBnNLEFgnEs.wrmacY,vD7wlSEDTeoiHT2dbUcYcDYYQOKEkpVYeY

  ## Verify User Interface URL ##
  #
  # The URL of the Verify User Interface, the page where users can request activate and / or verify their email or telephone number.
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_VERIFY_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_VERIFY_UI=<value>
  #
  verify_ui: https://zxaFXYDA.cgbCNEpxg4VfUuoZbdn7xyBTy3SbKZgP

  ## default_return_to ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_DEFAULT_RETURN_TO=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_DEFAULT_RETURN_TO=<value>
  #
  default_return_to: http://nXykReCXwPFMMsLiHr.dxuWIIE+65ftLp2swqQndP0d

  ## self ##
  #
  self:
    ## public ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export URLS_SELF_PUBLIC=<value>
    # - Windows Command Line (CMD):
    #    > set URLS_SELF_PUBLIC=<value>
    #
    public: http://iZKbRSHZepNHxkfU.tqiPEhHqAdl9zphBpK0JLkt+kyoZWmfl1sn

    ## admin ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export URLS_SELF_ADMIN=<value>
    # - Windows Command Line (CMD):
    #    > set URLS_SELF_ADMIN=<value>
    #
    admin: http://PfnGJZXKZpWcmAW.wcmhK-Z

  ## whitelisted_return_to_urls ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_whitelisted_return_to_urls=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_whitelisted_return_to_urls=<value>
  #
  whitelisted_return_to_urls:
    - https://mLXbuCzlJbAaYVJaVQ.namM9tRqlZAoKO+uRYpH
    - http://WhIslWEQlfsYCGFmZrbPBCAX.qqlJ4L
    - http://JDOLAPSVNPPoVJlrCKjfXcNRPwlhn.qbcYGB2ssMW5wJ6gALSyIC9Xcb
    - http://AvZhqhALhEiVGwKHYv.hejlPo1f
    - http://srNUFuVdN.wtvgraSNguY0z+m29TUKewiZro,XuP8yHw.-6Pdl+dStj1JDZe

## log ##
#
log:
  ## level ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export LOG_LEVEL=<value>
  # - Windows Command Line (CMD):
  #    > set LOG_LEVEL=<value>
  #
  level: warning

  ## format ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export LOG_FORMAT=<value>
  # - Windows Command Line (CMD):
  #    > set LOG_FORMAT=<value>
  #
  format: json

## secrets ##
#
secrets:
  ## session ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export SECRETS_SESSION=<value>
  # - Windows Command Line (CMD):
  #    > set SECRETS_SESSION=<value>
  #
  session:
    - eiusmodeu animdolore aliqua
    - dolore officiaaliqua pariatur in adipisicing
    - occaecat ad sit velit aute
    - sit sedofficia commodo magna

## hashers ##
#
hashers:
  ## argon2 ##
  #
  argon2:
    ## memory ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_MEMORY=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_MEMORY=<value>
    #
    memory: 59788135

    ## iterations ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_ITERATIONS=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_ITERATIONS=<value>
    #
    iterations: 89048388

    ## parallelism ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_PARALLELISM=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_PARALLELISM=<value>
    #
    parallelism: 6070036

    ## salt_length ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_SALT_LENGTH=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_SALT_LENGTH=<value>
    #
    salt_length: 94287601

    ## key_length ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_KEY_LENGTH=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_KEY_LENGTH=<value>
    #
    key_length: 97471191

## security ##
#
security:
  ## session ##
  #
  session:
    ## cookie ##
    #
    cookie:
      ## same_site ##
      #
      # Default value: Lax
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SECURITY_SESSION_COOKIE_SAME_SITE=<value>
      # - Windows Command Line (CMD):
      #    > set SECURITY_SESSION_COOKIE_SAME_SITE=<value>
      #
      same_site: Strict
```
