---
id: configuration
title: Configuration
---

<!-- THIS FILE IS BEING AUTO-GENERATED. DO NOT MODIFY IT AS ALL CHANGES WILL BE OVERWRITTEN.
OPEN AN ISSUE IF YOU WOULD LIKE TO MAKE ADJUSTMENTS HERE AND MAINTAINERS WILL HELP YOU LOCATE THE RIGHT
FILE -->

If file `$HOME/.kratos.yaml` exists, it will be used as a configuration file which supports all
configuration settings listed below.

You can load the config file from another source using the `-c path/to/config.yaml` or `--config path/to/config.yaml`
flag: `kratos --config path/to/config.yaml`.

Config files can be formatted as JSON, YAML and TOML. Some configuration values support reloading without server restart.
All configuration values can be set using environment variables, as documented below.

To find out more about edge cases like setting string array values through environmental variables head to the
[Configuring ORY services](https://www.ory.sh/docs/ecosystem/configuring) section.

```yaml
## ORY Kratos Configuration
#


## Data Source Name ##
#
# DSN is used to specify the database credentials as a connection URI.
#
# Examples:
# - postgres://user:password@localhost:5432/database
# 
# Set this value using environment variables on
# - Linux/macOS:
#    $ export DSN=<value>
# - Windows Command Line (CMD):
#    > set DSN=<value>
#
dsn: postgres://user:password@localhost:5432/database

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
    default_schema_url: http://DtjjDkKuieRldUloBoJUH.pzd5TVXzmxCKYP,Xjaoykz-OpUOxIIgqMSCDy4d4PW.X0wak8rRBCPJWfUG4

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
      - occaecat
      - false

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
    redirect_to: https://CWHESVGZUaKjSwIUWZ.ruQTscEqPJ8vmBFeLV-VqOSmnv09BIOIk0NUdq.5vi5q9Akt+LP

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
      enabled: false

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
          - id: occaecat velit
            provider: github
            client_id: Excepteur reprehenderit irure aliqua
            client_secret: Ut do
            schema_url: https://HOKSAviGwOLGpwhZyKhtEaFJDNApaAh.jiwbXiOt2m5UFza0vfw,pQuyO7MmtB2Doo,eYwsWhqrbtB2l0LQHnd-QTq5sZe
            issuer_url: http://XixfGcTIez.jnHS-kaBKiuWE
            auth_url: https://oCDRaiUpwPCMIJoigltgbgOvN.tfyoVWWiHe3ExNY4-YJDSrK,o9CE+zRi7lvQu,Ly5cETsydyq2cOhvHyLx+Htcn
            token_url: https://XWJqt.wyjywG-PnQRl.DLX8rPGFy2GWgskty6mVEaViUHkg3y0.QetxWr2pbx3IyXPTpEkub4llAX5t6cUSjW
            scope:
              - exercitation aute in dolore elit
              - ipsum veniam mollit
              - qui occaecat
              - cupidatat

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
    request_lifespan: 53ns

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
    privileged_session_max_age: 467227398us

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
        - job: redirect
          config:
            default_redirect_url: http://iqM.knNoTfoDcL-cvBrSbEQDv3bKr-nmRJEI+
            allow_user_defined_redirect: false

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
            default_redirect_url: https://iervFKfgunZwAKXOqyn.gafqFvCmRSK6TeLOAHKrXUiXJ4Bw
            allow_user_defined_redirect: true
        - job: redirect
          config:
            default_redirect_url: http://yrazqvImTyVFAG.ukvxOlH-n6IRVHao+LOelMQsweXiu5lfO8ArF0Q,vGASTr+
            allow_user_defined_redirect: false
        - job: redirect
          config:
            default_redirect_url: https://tbEYSsWBDi.conqJATGyLu28i6jSvrVvjg4DYe3pbPgxe
            allow_user_defined_redirect: false
        - job: redirect
          config:
            default_redirect_url: https://DlbIzoOqhuchDdXEVprcav.oxocLDpfnxrfHnbGqMiITQ+AwzHldeeZW4wqUBr1HvG6lCMJF0KUVLUFlMimw8Zo4jusqbfDcgDV
            allow_user_defined_redirect: false
        - job: redirect
          config:
            default_redirect_url: https://zYpCBNzmaK.xhwfJy34JI8seWF56CNXIBK,iGB.u3bikEOcYXvwAXWbzqCJ9FoxdysDfd5Na,ygHjUk
            allow_user_defined_redirect: false

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
    request_lifespan: 1161796018ms

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
    link_lifespan: 4938m

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
    request_lifespan: 27131m

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
          default_redirect_url: http://YREurHiLzp.ttwfmsYFQA9aGYT3Qu8htMz-wHAW0AvQ
          allow_user_defined_redirect: true
      - job: redirect
        config:
          default_redirect_url: https://hxQRVEKrwjdzhzjLBHNGBC.ungocFe,.HUdc9ct.fXZLvUeNHhKcgmcEntz,krQFNqLAZbl36BnBHEPhQlHsvJABH
          allow_user_defined_redirect: false
      - job: redirect
        config:
          default_redirect_url: https://meTWpwfA.ttphLRWbjY-DiwVEdsTuhLEy-gp-k
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
            default_redirect_url: https://gpGbnLHFfCvFKWuhqWpyLlcbJYGD.laIWvzfRX8QDxRI3Dr4Ucc.SUHsH
            allow_user_defined_redirect: true
        - job: redirect
          config:
            default_redirect_url: https://puPhOLCtteVfJMCBgetkleH.rrknmg.rUAueqwSn6bmxDK8M7CDjdbcprssPvwtgCYcbtjIHpI-SF
            allow_user_defined_redirect: true
        - job: session
        - job: redirect
          config:
            default_redirect_url: http://UpoxON.ogqxXKs4vv5dEaesVNBOCjY44FLI1UmzKxhN5Z5S1Hff2ATifH6qh8vv93AyIbA8S7qZsZbC1ShT5
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
        - job: revoke_active_sessions
        - job: redirect
          config:
            default_redirect_url: https://SmJsGNNluRRKSgxOXRgiUVzJiEfLSPot.jwPunEJNPf6c6jB5jr2U0CqzFmI4J8aAJ
            allow_user_defined_redirect: true
        - job: session

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
    request_lifespan: 72565h

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
          default_redirect_url: https://SypafnuKOgLcckTwxntddnduEyNI.eowW7OL2xW23+TYUH,iiBG1JQ,ObEZ7lW67OFM
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
            default_redirect_url: http://hJKGYTVgVKkuztkNPTFeVrd.wzddjJn
            allow_user_defined_redirect: false
        - job: verify
        - job: session
        - job: redirect
          config:
            default_redirect_url: https://jXNiSAcUCnCA.psOQuczXJrk0rVCBl2QXiRV1i
            allow_user_defined_redirect: true
        - job: redirect
          config:
            default_redirect_url: http://QMpBCZwClyHilHGYbStxkxw.jkfSpIEWNSsjxeoRymJcpC
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
        - job: session
        - job: verify

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
    from_address: vAKZxglwVAyu5Cp@JojoWUObAffBJknCkaupbxfkkbqAsSOwr.mkq

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
  template_override_path: anim adipisicing labore ad

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
    host: aliquip Ut

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
    host: dolore amet dolor

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
  settings_ui: https://TEQPRfsOc.zeuD9WxdAqFein24B7RGFr1xbz9elO3Epo2pe

  ## mfa_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_MFA_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_MFA_UI=<value>
  #
  mfa_ui: http://p.xhkFW6mGC3m3oxwfROys8JCTIFltsVXLLlm+eVhKrtXo-5hK

  ## login_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_LOGIN_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_LOGIN_UI=<value>
  #
  login_ui: https://hYiS.ssamQbQlXOzWW6ZMi0RFITqWKbjQ09Qo1003mYgeryQrc7qcI

  ## registration_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_REGISTRATION_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_REGISTRATION_UI=<value>
  #
  registration_ui: https://UKkmnyVtTItDZUjPvvA.igk2KB,iPptbkfGq4vEXrjuMX8LAON9mQPISuxoE7rKzs7vUxBHz5wCal

  ## error_ui ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_ERROR_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_ERROR_UI=<value>
  #
  error_ui: https://PSFLXNBuuzLOINZPpwJbAHnIyhXlaNHp.tgdnUMxLi1ahwdXmTScR0ovpGMDYC7ITzg4MFNw1qPb

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
  verify_ui: http://VpsTSvbIoDGR.sqenIgZWamJhTygY3bZKreNJMJkZIcinLb1T2roKQ9CbChdcm12IAdfJNyRIc3NfjrH.AE,

  ## default_return_to ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_DEFAULT_RETURN_TO=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_DEFAULT_RETURN_TO=<value>
  #
  default_return_to: https://kvaMNtsGsytSgLZWJbMqJclCOlCx.joihqdxBskgZOihrJ

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
    public: http://NXtzTFgrZV.bvbppl7WvKv,shiH0MiMgabqAeldgmQn3InJhGmC0zDraJR8P0KrB8

    ## admin ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export URLS_SELF_ADMIN=<value>
    # - Windows Command Line (CMD):
    #    > set URLS_SELF_ADMIN=<value>
    #
    admin: https://aqw.hxbqLQgxMl8.fzOQhgh4-y

  ## whitelisted_return_to_domains ##
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_WHITELISTED_RETURN_TO_DOMAINS=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_WHITELISTED_RETURN_TO_DOMAINS=<value>
  #
  whitelisted_return_to_domains:
    - https://bLogannxoDTTyWOapwPx.kafZ47UAltiG
    - http://po.xlu-.ING4fGzXovGecl8mTqF5CNvEs-f3PEG8pUBW7.9kP7vC4iVJLM
    - https://CcljrjsSWDdiUyrPOFtjiN.lnZ,7+DVwJghPODsoP3ZPT67PHTbGXNxelOCqLdPNZuT33YVch
    - https://UOuNPcscNlARebZqSEXVVAzXmOpZog.tvjPb6osVZh9Ne4mJzgi-CyxSl1MLf7rqRQM6WtEhL5yft9bI.uWkkaXSp,+GX2Z
    - https://Ckk.qpaes86fTOaygFZ6vhEv5d

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
    - aliquipveniam aute
    - quis ut incididunt proident
    - fugiatanim irure proident
    - Ut ea proidentUt minim veniam nostrud irure
    - esse dolor tempor ipsum fugiat

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
    memory: 76609206

    ## iterations ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_ITERATIONS=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_ITERATIONS=<value>
    #
    iterations: 56167537

    ## parallelism ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_PARALLELISM=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_PARALLELISM=<value>
    #
    parallelism: 50445769

    ## salt_length ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_SALT_LENGTH=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_SALT_LENGTH=<value>
    #
    salt_length: 9907940

    ## key_length ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_KEY_LENGTH=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_KEY_LENGTH=<value>
    #
    key_length: 28822133

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
      same_site: None

```