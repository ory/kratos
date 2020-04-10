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

## dsn ##
#
# Set this value using environment variables on
# - Linux/macOS:
#    $ export DSN=<value>
# - Windows Command Line (CMD):
#    > set DSN=<value>
#
dsn: incididunt sed

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
    default_schema_url: http://UiZBEpglJFJasHwcnmGSNLqNxceqTszQ.yudf86xNGtsuzQwojxwBwmoZgPqybhHtFx8MVPKagGRv198efH7rZ+l0EfBn59SGKTcGTYegGwz

    ## schemas ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export IDENTITY_TRAITS_SCHEMAS=<value>
    # - Windows Command Line (CMD):
    #    > set IDENTITY_TRAITS_SCHEMAS=<value>
    #
    schemas:
      - -4426198.611738071
      - -20973061
      - 30539143.671887502
      - -80654892
      - -23320403

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
    redirect_to: http://rQcuiQiHEQtVbjJdjfAdAjZPCdrKB.ekpGZyP.X.cJOzZkIDVD

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
          - id: ipsum enim pariatur ex
            provider: generic
            client_id: elit pariatur esse voluptate
            client_secret: exercitation nostrud
            schema_url: http://PpxILCGFxIdvmdzSG.zcwXeqXRZ6z0nV.1PHlM-CLTQmf5H-
            issuer_url: http://UTJOJNnIFuKioDApBvYJRVQ.dipeLxSyff6NngZKKFoWIe3xdp6ZBehasuaekBqv+jC7ohG6lTCRqgA1wOVdlwKw3
            auth_url: http://lDicgdbXPdTMhGiXQuy.yotxYu8SdnadrRfbIHlpBmT-Wamk1YGKonJ3sJn.urUPs-PDVeqp
            token_url: https://yyChgalbSouVoVmuGKOmiXCTNTugR.lbfu8oK79mGfOV.qgZyI7ZElr2wKMiAaRr2jfv-pAd-fvH166UkH9
            scope:
              - sed
              - ad sunt fugiat non
              - ex consectetur mollit in
              - Excepteur non

  ## profile ##
  #
  profile:
    ## request_lifespan ##
    #
    # Default value: 1h
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_PROFILE_REQUEST_LIFESPAN=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_PROFILE_REQUEST_LIFESPAN=<value>
    #
    request_lifespan: 6631us

    ## privileged_session_max_age ##
    #
    # Default value: 1h
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_PROFILE_PRIVILEGED_SESSION_MAX_AGE=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_PROFILE_PRIVILEGED_SESSION_MAX_AGE=<value>
    #
    privileged_session_max_age: 97ms

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
    request_lifespan: 81s

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
    link_lifespan: 01443817h

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
    request_lifespan: 9801s

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
          default_redirect_url: https://YMiAxqdnyOXqOsOOATQsWtnO.obwaCoJh2VJjvhMmE3s8+vHoxt1
          allow_user_defined_redirect: false
      - job: redirect
        config:
          default_redirect_url: http://gNypoTJloTerpFuUTZ.vevYHEnBqnjty9tCK0zO1nG1mV3UNlTjVV0Yga
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
        - job: session

      ## oidc ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_LOGIN_AFTER_OIDC=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_LOGIN_AFTER_OIDC=<value>
      #
      oidc:
        - job: redirect
          config:
            default_redirect_url: https://GLmUoEaZe.ffGBPEFSxPGfHtXYdupILxsQcfRLT0xSya0RY0
            allow_user_defined_redirect: false

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
    request_lifespan: 97733237ns

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
          default_redirect_url: http://adpKHiYwrLHVMHxbehR.sletbrGNjT,ayog.gSCMhi
          allow_user_defined_redirect: false
      - job: redirect
        config:
          default_redirect_url: http://wlgRNYewWndobZdfGcsamnSvvkZfwbPeA.uqsrqOKPw9pUXh,wnA2PO3s7
          allow_user_defined_redirect: true
      - job: redirect
        config:
          default_redirect_url: https://nTKzsMZNNuca.xeoOAH657.F
          allow_user_defined_redirect: true

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
        - job: verify
        - job: redirect
          config:
            default_redirect_url: http://hbdMlaQwGMyWFPuzXwQxBRphtalNHwMxK.ipxruNeStlNsSe2UG5Hy,crqed.-X
            allow_user_defined_redirect: false
        - job: redirect
          config:
            default_redirect_url: http://k.qxnZsxSsJUHqgegZzQhTtLF4o14Uy4E
            allow_user_defined_redirect: false
        - job: redirect
          config:
            default_redirect_url: http://SlvUPhsysIqAKUrFPgEJVm.cdoKMCMhoHTzs.MfoMlr3sTeIPVRqLr+EmYNEBmdA02X7tl2B
            allow_user_defined_redirect: true
        - job: redirect
          config:
            default_redirect_url: http://e.hnzjfzdHCZviT,pnFxpUmThECVcsnk
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
        - job: verify
        - job: redirect
          config:
            default_redirect_url: https://hkjzLy.khlbvZi4LCf3-wdx1zGx4sNwDI1Ka
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
    from_address: fet@yzfcDtrPKTjPKtruIQ.xxfq

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
  template_override_path: exercitation dolor nisi

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
    host: velit Duis

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
    host: incididunt dolore ea commodo sit

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
  ## profile_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_PROFILE_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_PROFILE_UI=<value>
  #
  profile_ui: https://YUHf.ljyE6ce,izXD0+cX4h.UgwvzlzvTWwuPZA2I

  ## mfa_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_MFA_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_MFA_UI=<value>
  #
  mfa_ui: http://SDclIjhhLBHpxiybCEPvCcHusj.qqccejcQkCk5FB-tuFVsetfMquNfHOry6upMe

  ## login_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_LOGIN_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_LOGIN_UI=<value>
  #
  login_ui: http://dpnTUSElNHmXWRrrtbkY.tjqgokPIn21eBI3f-gL18A,sezrd1qG-XgE4QICLW,xqeqeYjXjIe

  ## registration_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_REGISTRATION_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_REGISTRATION_UI=<value>
  #
  registration_ui: https://GrlEuZIioiSNyFNPW.dnovnIm+TWTekKojZeHVgMnCSoA0Ws,Rze8pU-4bO

  ## error_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_ERROR_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_ERROR_UI=<value>
  #
  error_ui: https://sSPVIE.ugfHsG3LuzH

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
  verify_ui: http://mhwtRNkuyoCqYNMUv.uxeeSBuIGdz2vlD7KM3X2KTdvP7eWgXOW612qdJke2GuG8c,rIB142nad2x1qibe+d

  ## default_return_to ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_DEFAULT_RETURN_TO=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_DEFAULT_RETURN_TO=<value>
  #
  default_return_to: http://ljnYOAfFOjxducArzGPCspTrFvihXysoP.unbmoqo+Ll-t0+aWP6ivsIBzRBfqAbn9fXz9Ylf.gn,jm

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
    public: http://iGlUd.tqacGB2JGr-sjo38AP8Vs+iODE

    ## admin ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export URLS_SELF_ADMIN=<value>
    # - Windows Command Line (CMD):
    #    > set URLS_SELF_ADMIN=<value>
    #
    admin: http://jHBWBVdCpFDOIabFPH.ckcs7zgGtvdb5XzVuOfrAR02EdZ

  ## whitelisted_return_to_domains ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_WHITELISTED_RETURN_TO_DOMAINS=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_WHITELISTED_RETURN_TO_DOMAINS=<value>
  #
  whitelisted_return_to_domains:
    - http://gxKvKS.znwzk2i8HY-GUkMoLct3JU.pvrWFrhewWsktb7ciGA9Z0UEOLUXi2VkrM+WpH
    - https://GcnbHAQKJAeaPecnOKUNvrSsqLcNdq.toauHDa751a2KhtqtyNeciOtKUxLTE7cZgX1gc85G5TzIKuFyC2yU0wjLyUfR

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
  level: info

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
    - mollit dolore laboris
    - Ut etExcepteursit in culpa nisi
    - magna aliqua aute
    - dolor occaecatnulla
    - ea aliquasint irure sunt sit

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
    memory: 31292956

    ## iterations ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_ITERATIONS=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_ITERATIONS=<value>
    #
    iterations: 3228708

    ## parallelism ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_PARALLELISM=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_PARALLELISM=<value>
    #
    parallelism: 52425634

    ## salt_length ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_SALT_LENGTH=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_SALT_LENGTH=<value>
    #
    salt_length: 57006570

    ## key_length ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_KEY_LENGTH=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_KEY_LENGTH=<value>
    #
    key_length: 18810883

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
