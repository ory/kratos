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
[Configuring ORY services](../../ecosystem/configuring) section.

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
dsn: adipisicing quis magna

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
    default_schema_url: https://fsQWhUJflUSCcPvHBIDYpjJOCmf.feFZPfpkba

    ## schemas ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export IDENTITY_TRAITS_SCHEMAS=<value>
    # - Windows Command Line (CMD):
    #    > set IDENTITY_TRAITS_SCHEMAS=<value>
    #
    schemas:
      - null
      - []
      - -16370057.675011605
      - -2944342
      - 82072525

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
    redirect_to: http://uMesjJKa.fdlESZbxhS5mhX,RCtdpVBKBzJYAm2

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
      enabled: false

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
          - id: dolore ullamco reprehenderit
            provider: github
            client_id: id et in cillum proident
            schema_url: http://zeRWBQejLfABIdzNEfrSBcLTuTTsWlK.ckZ-.U4p+R26OVv1LNHl4,kHDOTnrzHLNsHQgUUJQj6CsXnAxXP3cKpyjGvSHAQLtH1
            issuer_url: https://wIrFTwmovoKVoeZiPnoaMfpiNd.aqpSUH347WD5ZquJRFUgIW8VR8Jmi6.4ed4t.gGY.igAXg1TsGpijcsw7XCEKOEr+8k4jdvrLzkB+Iv0m+lFZuixAl.
            auth_url: https://xWlkOhIbvawWcUiyuiLyDtzCiu.ughS
            token_url: http://WnNeimhHxoFymJRIQDTwVSqpsTEszP.yiqoQPZE5rPQ9x1H8KCSWY1IDYIy7RkBMLL-Xu,t8u.
            scope:
              - amet
          - id: est veniam nostrud elit
            provider: github
            client_id: do
            schema_url: http://Cn.oolyMaWmMi4pTO-jqo++bkeqdldD
            issuer_url: https://TyaRfofiGhjUWVII.didlqAFXsH.caY-VEy-Z.KBc9r71AU2Rxf0wOpRK
            auth_url: https://QrWMZCtuycdGKwNFKfvAREL.kkqiK9NJavl
            token_url: http://igwccDHZToGCOUrVqujyGrteCdUgtfNLI.otrjCObMofSs75jungrzhMZ2SHvMxQoskV8XnlZspfqL-TomMP6sue
            scope:
              - qui anim
          - id: nulla es
            provider: generic
            client_id: aliqua qui exercitation Duis
            schema_url: https://FgrlUVxuVaMtict.erciZp+jEnu8,WvWVUcGiHe,BmCX8ni0z1j-NIvOMdzwQpoK6Uo4rVrj.QjmPmMVAX0uwcy
            issuer_url: https://wbXJIWddaeRZWRYuojuTAtreapAupo.nklaJ+12VzVO8Y
            auth_url: https://UMpByu.ncBkPk9h97UwFZ5xZjg-dZuw3LsDRfI,dZwZD5ZTLWbOTUNfu2OJ3mJljuFU6Cjl
            token_url: http://ivjnuiTlponKmKGkYHeLnfbmz.jpfwaFGcSoW4g23ELEl+uy
            scope:
              - ad dolor tempor commodo ipsum
              - magna officia proident Excepteur
              - voluptate
              - incididunt culpa tempor reprehenderit
              - in
          - id: eiusmod laboris non in
            provider: github
            client_id: mollit aliquip nostrud esse aute
            schema_url: http://nkVRLQFnhyWibFpgeeNozBAMQeUYUK.wircLy.rJ-Hu0JaYlH-BFYV
            issuer_url: https://WpBMOxrz.gxefFR0myyV-yE7QnrMBAhcaYo,C9DY4Jlhcoc
            auth_url: https://HqIifoBUfsxD.ehojuxzhGhXNjaa.CbqQu5s2rA2tP6PADF
            token_url: https://A.vrxcEVymjMeTB4OO1Ubu1MdhWxNH.bjnKGNUYsP51AmTCe+l5S
            scope:
              - esse
          - id: anim dolore deserunt
            provider: github
            client_id: consequat Lorem
            schema_url: https://EEsrJDUFuxGUhxdiXUZVpVXadUP.riimMGzuTjiX+Q.8mMeX7bdd.IlCjHIFUO0J7x1
            issuer_url: http://YL.rrtCVlqWYcy.
            auth_url: http://guYsEAWvAANMVFlrkto.ltTjqLauMnzvzrPf6yPuVpuUYXQ.t9.Yl0
            token_url: http://eCMApCJdeyOkNyGYfsfaT.efgyI3tNDNLuRWhXV.uSxyTk2PfEogwQKnR++BzTXav2dKdcInkGWnsVaMxGcnZYoAiyFIHV,GJ7-V6o
            scope:
              - pariatur est exercitation ex
              - reprehenderit in Ut
              - qui ullamco

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
    request_lifespan: 933ns

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
    request_lifespan: 217377539ms

    ## before ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_LOGIN_BEFORE=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_LOGIN_BEFORE=<value>
    #
    before:
      - run: redirect
        config:
          default_redirect_url: http://WKfggnUdmBmqSGXNBhcTmzxmTlwY.nvNiUVqGPr9hdxgvuew2qzizFpXaIMI6XjxGh1GtewlVslnrpWDKMI
          allow_user_defined_redirect: true
      - run: redirect
        config:
          default_redirect_url: https://Q.qqrQRwd9D2LpCIIsiJGmDtwUeFN7K2M,oEYmplOPm5GvyAEV.ZjgaiUlpd5VTl.YfIBYtk7x
          allow_user_defined_redirect: false
      - run: redirect
        config:
          default_redirect_url: https://djbWNLUWyfiiBTemHvfDzAgzdgdiPuC.the52ivnMoYIqFKwg.WW.GcLu0tLJaE6IqbYIeIwv3
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
        - run: redirect
          config:
            default_redirect_url: https://idbLfKjZFKs.sokzrcOcBmbMSMMOIsIO8AnPIB+WDwQOFCgfevDYboYcfgrtVefktvqcRyd1tqXV
            allow_user_defined_redirect: true
        - run: session
        - run: redirect
          config:
            default_redirect_url: http://lWQWVljOpQNkEdnWSOnBswGdNJ.qhzp6ESagmSF8kvUu-yr5VcpchTA4AlEM
            allow_user_defined_redirect: true
        - run: redirect
          config:
            default_redirect_url: https://bLavqxYeLbzzOyGQiyNYHXw.btwaM6mhfiLwmzr.YmStIXekF-4sMSQ2UYoy-H76gYKDt-f4fCYGTPXgxRTyUTGDYNyX.,Gz8
            allow_user_defined_redirect: false

      ## oidc ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_LOGIN_AFTER_OIDC=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_LOGIN_AFTER_OIDC=<value>
      #
      oidc:
        - run: revoke_active_sessions
        - run: redirect
          config:
            default_redirect_url: https://eqIDyCgC.helrz-0zQRBqWyoU31XsmMmFK7XNs6qL3P0RVCAndpJTWwawgri52Rl+CfkCF4Q.EbLa
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
    request_lifespan: 1116ns

    ## before ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_REGISTRATION_BEFORE=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_REGISTRATION_BEFORE=<value>
    #
    before:
      - run: redirect
        config:
          default_redirect_url: https://CiXnyysyrrPeKfMLdnBTleuBpBpgEDd.ukqsoZCJ,vK
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
        - run: session
        - run: redirect
          config:
            default_redirect_url: https://FbuCnyeKaYmo.nbivoAoTwc-7JvoBpq93aVeilzJaGIWFDZ0FlsvBbPkvzs
            allow_user_defined_redirect: true
        - run: redirect
          config:
            default_redirect_url: https://SUBHhklobrYuz.slt
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
        - run: redirect
          config:
            default_redirect_url: http://KMtPDYgYxNkRVUmBpsLMBZBZisPIgzFO.uolNIOExHuWuoEU5i.f+B
            allow_user_defined_redirect: true
        - run: redirect
          config:
            default_redirect_url: https://dcEuYUmtKmHmkqJ.sdNWbF59+cHhkVo9mh5wGFBEYdfsS+ZE
            allow_user_defined_redirect: false
        - run: redirect
          config:
            default_redirect_url: http://CoDsQYpesH.ogugUteCcFNFDteKbrGEr,v1B9xQ7
            allow_user_defined_redirect: true
        - run: redirect
          config:
            default_redirect_url: https://bmzOIiWDIxfBW.efwmWW6vNrE.n43ZPwuAGjRYbT9ZY3hVQONiO
            allow_user_defined_redirect: true
        - run: session

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
    from_address: G4ZWtG@Ox.lijk

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
  template_override_path: officia sunt

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
    host: aliqua eiusmod reprehenderit culpa

    ## port ##
    #
    # Default value: 4434
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_ADMIN_PORT=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_ADMIN_PORT=<value>
    #
    port: 47581

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
    host: anim occaecat laboris incididunt nulla

    ## port ##
    #
    # Default value: 4433
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_PUBLIC_PORT=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_PUBLIC_PORT=<value>
    #
    port: 38310

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
  profile_ui: http://yhCqzzjuno.ydwyNMX6Fzj3mJBvw9yfpLS7eP8+G7dlFHunmv5IVKaVoLiI1Hygw4R+5a,GSMnNNqhG1QQ5a4iJDb0ztC

  ## mfa_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_MFA_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_MFA_UI=<value>
  #
  mfa_ui: http://w.ozgPy9-ufRovl351Pb

  ## login_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_LOGIN_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_LOGIN_UI=<value>
  #
  login_ui: http://s.angSp4gCZ89mgfiu9q7jpnYsT

  ## registration_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_REGISTRATION_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_REGISTRATION_UI=<value>
  #
  registration_ui: https://H.fnhywbKRGZDekHGSg5WcE+njB+7wueD4e5dzGoTHKV4xtOQK4t5697zso

  ## error_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_ERROR_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_ERROR_UI=<value>
  #
  error_ui: http://wjBFVnqGcAFEPnK.tmflABUvZW,8IoR4ah4EgdL,TjSnmkQLzlE6FCOLvoK5jZO

  ## default_return_to ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_DEFAULT_RETURN_TO=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_DEFAULT_RETURN_TO=<value>
  #
  default_return_to: https://KTWwKDBgbuFYKYAVLpNTVGt.whdudJ5eReNhlasauqLkbZ1odnFaq9RlcyPyu3z

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
    public: https://CvuBlPvgtCGsmNFrooS.kyxK7eJfGYj.B02EI5LlzW-pZyGlvdzjo+TbzBX+2

    ## admin ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export URLS_SELF_ADMIN=<value>
    # - Windows Command Line (CMD):
    #    > set URLS_SELF_ADMIN=<value>
    #
    admin: https://k.uosuV9z4QJRr72Ptw6mMYYCiex

  ## whitelisted_return_to_domains ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_WHITELISTED_RETURN_TO_DOMAINS=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_WHITELISTED_RETURN_TO_DOMAINS=<value>
  #
  whitelisted_return_to_domains:
    - http://fUCoNcvtLcLRqdgfOrDzQEVDaWC.klofQ..pIhyCFJ.Dp32ak5eoL37p8ZrRX
    - http://IUCqrXgpYMZFGFnVKuXj.pwyMR5tUbVN+RyhfqWftM1ugWnl0rCwmtsKfBpJsOXwcO5RyiWzQwhV3s4HWp

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
  level: trace

  ## format ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export LOG_FORMAT=<value>
  # - Windows Command Line (CMD):
  #    > set LOG_FORMAT=<value>
  #
  format: text

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
    - suntproident deserunt ad

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
    memory: 21382549

    ## iterations ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_ITERATIONS=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_ITERATIONS=<value>
    #
    iterations: 37640235

    ## parallelism ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_PARALLELISM=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_PARALLELISM=<value>
    #
    parallelism: 39508903

    ## salt_length ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_SALT_LENGTH=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_SALT_LENGTH=<value>
    #
    salt_length: 14637443

    ## key_length ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_KEY_LENGTH=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_KEY_LENGTH=<value>
    #
    key_length: 48794105
```
