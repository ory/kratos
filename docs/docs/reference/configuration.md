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
dsn: "postgres://user:
  password@postgresd:5432/database?sslmode=disable&max_conns=20&max_idle_conns=\
  4"

## identity ##
#
identity:
  
  ## traits ##
  #
  traits:
    
    ## JSON Schema URL for default identity traits ##
    #
    # Path to the JSON Schema which describes a default identity's traits.
    #
    # Examples:
    # - file://path/to/identity.traits.schema.json
    # - httpss://foo.bar.com/path/to/identity.traits.schema.json
    # 
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export IDENTITY_TRAITS_DEFAULT_SCHEMA_URL=<value>
    # - Windows Command Line (CMD):
    #    > set IDENTITY_TRAITS_DEFAULT_SCHEMA_URL=<value>
    #
    default_schema_url: httpss://foo.bar.com/path/to/identity.traits.schema.json

    ## Additional JSON Schemas for Identity Traits ##
    #
    # Examples:
    # - - id: customer
    #     url: https://foo.bar.com/path/to/customer.traits.schema.json
    #   - id: employee
    #     url: https://foo.bar.com/path/to/employee.traits.schema.json
    #   - id: employee-v2
    #     url: https://foo.bar.com/path/to/employee.v2.traits.schema.json
    # 
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export IDENTITY_TRAITS_SCHEMAS=<value>
    # - Windows Command Line (CMD):
    #    > set IDENTITY_TRAITS_SCHEMAS=<value>
    #
    schemas:
      - id: customer
        url: https://foo.bar.com/path/to/customer.traits.schema.json
      - id: employee
        url: https://foo.bar.com/path/to/employee.traits.schema.json
      - id: employee-v2
        url: https://foo.bar.com/path/to/employee.v2.traits.schema.json

## selfservice ##
#
selfservice:
  
  ## logout ##
  #
  logout:
    
    ## Redirect to URL after logout ##
    #
    # Examples:
    # - https://www.myapp.com/home
    # 
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_LOGOUT_REDIRECT_TO=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_LOGOUT_REDIRECT_TO=<value>
    #
    redirect_to: https://www.myapp.com/home

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
      enabled: false

      ## config ##
      #
      config:
        
        ## OpenID Connect and OAuth2 Providers ##
        #
        # A list and configuration of OAuth2 and OpenID Connect providers ORY Kratos should integrate with.
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_STRATEGIES_OIDC_CONFIG_PROVIDERS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_STRATEGIES_OIDC_CONFIG_PROVIDERS=<value>
        #
        providers:
          - id: google
            provider: google
            client_id: dolore adipisicing amet
            client_secret: esse ex ipsum qui commodo
            schema_url: https://foo.bar.com/path/to/oidc.schema.json
            issuer_url: https://accounts.google.com
            auth_url: https://accounts.google.com/o/oauth2/v2/auth
            token_url: https://www.googleapis.com/oauth2/v4/token
            scope:
              - profile
              - profile
              - profile
              - profile
              - profile
          - id: google
            provider: google
            client_id: est
            client_secret: sit Excepteur cillum
            schema_url: file://path/to/oidc.schema.json
            issuer_url: https://accounts.google.com
            auth_url: https://accounts.google.com/o/oauth2/v2/auth
            token_url: https://www.googleapis.com/oauth2/v4/token
            scope:
              - offline_access
              - offline_access

  ## settings ##
  #
  settings:
    
    ## request_lifespan ##
    #
    # Default value: 1h
    #
    # Examples:
    # - 1h
    # - 1m
    # - 1s
    # 
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_SETTINGS_REQUEST_LIFESPAN=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_SETTINGS_REQUEST_LIFESPAN=<value>
    #
    request_lifespan: 1m

    ## privileged_session_max_age ##
    #
    # Default value: 1h
    #
    # Examples:
    # - 1h
    # - 1m
    # - 1s
    # 
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_SETTINGS_PRIVILEGED_SESSION_MAX_AGE=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_SETTINGS_PRIVILEGED_SESSION_MAX_AGE=<value>
    #
    privileged_session_max_age: 1s

    ## after ##
    #
    after:
      
      ## Default Return To URL ##
      #
      # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. This value may be overridden by a `default_return_to` in a lower configuration level (`foo.bar.default_return_to` overrides `foo.default_return_to` overrides `default_return_to`) and by the `?return_to` query in certain cases.
      #
      # Examples:
      # - https://my-app.com/dashboard
      # 
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_SETTINGS_AFTER_DEFAULT_RETURN_TO=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_SETTINGS_AFTER_DEFAULT_RETURN_TO=<value>
      #
      default_return_to: https://my-app.com/dashboard

      ## password ##
      #
      password:
        
        ## Default Return To URL ##
        #
        # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. This value may be overridden by a `default_return_to` in a lower configuration level (`foo.bar.default_return_to` overrides `foo.default_return_to` overrides `default_return_to`) and by the `?return_to` query in certain cases.
        #
        # Examples:
        # - https://my-app.com/dashboard
        # 
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_SETTINGS_AFTER_PASSWORD_DEFAULT_RETURN_TO=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_SETTINGS_AFTER_PASSWORD_DEFAULT_RETURN_TO=<value>
        #
        default_return_to: https://my-app.com/dashboard

        ## hooks ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_SETTINGS_AFTER_PASSWORD_HOOKS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_SETTINGS_AFTER_PASSWORD_HOOKS=<value>
        #
        hooks:
          - hook: verify

      ## profile ##
      #
      profile:
        
        ## Default Return To URL ##
        #
        # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. This value may be overridden by a `default_return_to` in a lower configuration level (`foo.bar.default_return_to` overrides `foo.default_return_to` overrides `default_return_to`) and by the `?return_to` query in certain cases.
        #
        # Examples:
        # - https://my-app.com/dashboard
        # 
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_SETTINGS_AFTER_PROFILE_DEFAULT_RETURN_TO=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_SETTINGS_AFTER_PROFILE_DEFAULT_RETURN_TO=<value>
        #
        default_return_to: https://my-app.com/dashboard

        ## hooks ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_SETTINGS_AFTER_PROFILE_HOOKS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_SETTINGS_AFTER_PROFILE_HOOKS=<value>
        #
        hooks:
          - hook: verify

  ## verify ##
  #
  verify:
    
    ## Self-Service Verification Request Lifespan ##
    #
    # Sets how long the verification request (for the UI interaction) is valid.
    #
    # Default value: 1h
    #
    # Examples:
    # - 1h
    # - 1m
    # - 1s
    # 
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_VERIFY_REQUEST_LIFESPAN=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_VERIFY_REQUEST_LIFESPAN=<value>
    #
    request_lifespan: 1h

    ## Self-Service Verification Link Lifespan ##
    #
    # Sets how long the verification link (e.g. the one sent via email) is valid for.
    #
    # Default value: 24h
    #
    # Examples:
    # - 1h
    # - 1m
    # - 1s
    # 
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_VERIFY_LINK_LIFESPAN=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_VERIFY_LINK_LIFESPAN=<value>
    #
    link_lifespan: 1h

  ## login ##
  #
  login:
    
    ## request_lifespan ##
    #
    # Default value: 1h
    #
    # Examples:
    # - 1h
    # - 1m
    # - 1s
    # 
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_LOGIN_REQUEST_LIFESPAN=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_LOGIN_REQUEST_LIFESPAN=<value>
    #
    request_lifespan: 1h

    ## before ##
    #
    before:
      
      ## hooks ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_LOGIN_BEFORE_HOOKS=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_LOGIN_BEFORE_HOOKS=<value>
      #
      hooks:
        - hook: redirect
          config:
            default_redirect_url: https://RuRpDxGostkcVWIQOAHwbGxVnxJTbqrrX.upInhARM2qxBxkG4xDqxZAT9UK7SXwH+-jGMGgmGJthoi.58AZVyPZWA+Gzjn2HQo
            allow_user_defined_redirect: true
        - hook: redirect
          config:
            default_redirect_url: https://FXKFEFkRfRHYXySvc.bpAjtCUEIaQIdV2blx7-J.cgEx
            allow_user_defined_redirect: false
        - hook: redirect
          config:
            default_redirect_url: https://xFqgELpmuXFky.nqJ.F99qIRqLpD2t1nuhP0OabsI04jrXZLIDZjoGhG.4AITTj8IPGG,Hw4HbkhRR4nRAg3E0QVuLCZ6
            allow_user_defined_redirect: true

    ## after ##
    #
    after:
      
      ## Default Return To URL ##
      #
      # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. This value may be overridden by a `default_return_to` in a lower configuration level (`foo.bar.default_return_to` overrides `foo.default_return_to` overrides `default_return_to`) and by the `?return_to` query in certain cases.
      #
      # Examples:
      # - https://my-app.com/dashboard
      # 
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_LOGIN_AFTER_DEFAULT_RETURN_TO=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_LOGIN_AFTER_DEFAULT_RETURN_TO=<value>
      #
      default_return_to: https://my-app.com/dashboard

      ## password ##
      #
      password:
        
        ## Default Return To URL ##
        #
        # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. This value may be overridden by a `default_return_to` in a lower configuration level (`foo.bar.default_return_to` overrides `foo.default_return_to` overrides `default_return_to`) and by the `?return_to` query in certain cases.
        #
        # Examples:
        # - https://my-app.com/dashboard
        # 
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_LOGIN_AFTER_PASSWORD_DEFAULT_RETURN_TO=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_LOGIN_AFTER_PASSWORD_DEFAULT_RETURN_TO=<value>
        #
        default_return_to: https://my-app.com/dashboard

        ## hooks ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_LOGIN_AFTER_PASSWORD_HOOKS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_LOGIN_AFTER_PASSWORD_HOOKS=<value>
        #
        hooks:
          - hook: revoke_active_sessions

      ## oidc ##
      #
      oidc:
        
        ## Default Return To URL ##
        #
        # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. This value may be overridden by a `default_return_to` in a lower configuration level (`foo.bar.default_return_to` overrides `foo.default_return_to` overrides `default_return_to`) and by the `?return_to` query in certain cases.
        #
        # Examples:
        # - https://my-app.com/dashboard
        # 
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_LOGIN_AFTER_OIDC_DEFAULT_RETURN_TO=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_LOGIN_AFTER_OIDC_DEFAULT_RETURN_TO=<value>
        #
        default_return_to: https://my-app.com/dashboard

        ## hooks ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_LOGIN_AFTER_OIDC_HOOKS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_LOGIN_AFTER_OIDC_HOOKS=<value>
        #
        hooks:
          - hook: revoke_active_sessions

  ## registration ##
  #
  registration:
    
    ## request_lifespan ##
    #
    # Default value: 1h
    #
    # Examples:
    # - 1h
    # - 1m
    # - 1s
    # 
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_REGISTRATION_REQUEST_LIFESPAN=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_REGISTRATION_REQUEST_LIFESPAN=<value>
    #
    request_lifespan: 1m

    ## before ##
    #
    before:
      
      ## hooks ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_REGISTRATION_BEFORE_HOOKS=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_REGISTRATION_BEFORE_HOOKS=<value>
      #
      hooks:
        - hook: redirect
          config:
            default_redirect_url: http://yGtxVUQOkfbyeholrdjtbLMNX.zgxXQF3uwDMsNe18k-zgvnskS8sH3ftjZ9WefSJ.n2wpJsFEdrXi3h
            allow_user_defined_redirect: true
        - hook: redirect
          config:
            default_redirect_url: https://zrRHtzdFxpdGvEqCzYTzOkuBWHCw.hammheOKSzGxmwue.2c3PhAovYQ2YlXHvjOmoxdYp,5KN3pOj5PxRxIpED-IflJ,uzyFaNsMfEnz
            allow_user_defined_redirect: false
        - hook: redirect
          config:
            default_redirect_url: https://KijJVaKCzIYlMuJiu.xicVNxwVHdFKWmUsG+vba5Ip1XNez3mLG0E7PYGo6.uxT2jP7l9U
            allow_user_defined_redirect: false

    ## after ##
    #
    after:
      
      ## Default Return To URL ##
      #
      # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. This value may be overridden by a `default_return_to` in a lower configuration level (`foo.bar.default_return_to` overrides `foo.default_return_to` overrides `default_return_to`) and by the `?return_to` query in certain cases.
      #
      # Examples:
      # - https://my-app.com/dashboard
      # 
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_REGISTRATION_AFTER_DEFAULT_RETURN_TO=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_REGISTRATION_AFTER_DEFAULT_RETURN_TO=<value>
      #
      default_return_to: https://my-app.com/dashboard

      ## password ##
      #
      password:
        
        ## Default Return To URL ##
        #
        # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. This value may be overridden by a `default_return_to` in a lower configuration level (`foo.bar.default_return_to` overrides `foo.default_return_to` overrides `default_return_to`) and by the `?return_to` query in certain cases.
        #
        # Examples:
        # - https://my-app.com/dashboard
        # 
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_REGISTRATION_AFTER_PASSWORD_DEFAULT_RETURN_TO=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_REGISTRATION_AFTER_PASSWORD_DEFAULT_RETURN_TO=<value>
        #
        default_return_to: https://my-app.com/dashboard

        ## hooks ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_REGISTRATION_AFTER_PASSWORD_HOOKS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_REGISTRATION_AFTER_PASSWORD_HOOKS=<value>
        #
        hooks:
          - hook: redirect
            config:
              default_redirect_url: https://QVrbrLUUJzBZdLvIlgrFtXMwndPbTGTd.muebtY,..w+CZHjfG0pyIwqlzUdDLEHY.mqwdPgQLid9T5FlvLNfJjulpOcTXeG-qaX,g
              allow_user_defined_redirect: true
          - hook: verify
          - hook: session
          - hook: redirect
            config:
              default_redirect_url: http://KIz.nhlueozSA0ugo8ozxRc,TaQVv4E1Vc5fpn
              allow_user_defined_redirect: false

      ## oidc ##
      #
      oidc:
        
        ## Default Return To URL ##
        #
        # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. This value may be overridden by a `default_return_to` in a lower configuration level (`foo.bar.default_return_to` overrides `foo.default_return_to` overrides `default_return_to`) and by the `?return_to` query in certain cases.
        #
        # Examples:
        # - https://my-app.com/dashboard
        # 
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_REGISTRATION_AFTER_OIDC_DEFAULT_RETURN_TO=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_REGISTRATION_AFTER_OIDC_DEFAULT_RETURN_TO=<value>
        #
        default_return_to: https://my-app.com/dashboard

        ## hooks ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_REGISTRATION_AFTER_OIDC_HOOKS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_REGISTRATION_AFTER_OIDC_HOOKS=<value>
        #
        hooks:
          - hook: session
          - hook: verify
          - hook: redirect
            config:
              default_redirect_url: http://nVPqduVH.xvvRVqj+u0
              allow_user_defined_redirect: true
          - hook: redirect
            config:
              default_redirect_url: https://mYFiatdMMUvTEERlyAh.ucptMqdAbUTLcQrWZvwWZqriGzs
              allow_user_defined_redirect: false

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
    # This URI will be used to connect to the SMTP server. Use the query parameter to allow (`?skip_ssl_verify=true`) or disallow (`?skip_ssl_verify=false`) self-signed TLS certificates. Please keep in mind that any host other than localhost / 127.0.0.1 must use smtp over TLS (smtps) or the connection will not be possible.
    #
    # Examples:
    # - smtps://foo:bar@my-mailserver:1234/?skip_ssl_verify=false
    # 
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export COURIER_SMTP_CONNECTION_URI=<value>
    # - Windows Command Line (CMD):
    #    > set COURIER_SMTP_CONNECTION_URI=<value>
    #
    connection_uri: smtps://foo:bar@my-mailserver:1234/?skip_ssl_verify=false

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
    from_address: dZq@Z.xtul

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
    host: tempor nisi enim consectetur veniam

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
    host: id occaecat aliqua reprehenderit amet

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
  
  ## Settings UI URL ##
  #
  # URL where the Settings UI is hosted. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
  #
  # Examples:
  # - https://my-app.com/user/settings
  # 
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_SETTINGS_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_SETTINGS_UI=<value>
  #
  settings_ui: https://my-app.com/user/settings

  ## Login UI URL ##
  #
  # URL where the Login UI is hosted. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
  #
  # Examples:
  # - https://my-app.com/login
  # 
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_LOGIN_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_LOGIN_UI=<value>
  #
  login_ui: https://my-app.com/login

  ## Registration UI URL ##
  #
  # URL where the Registration UI is hosted. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
  #
  # Examples:
  # - https://my-app.com/signup
  # 
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_REGISTRATION_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_REGISTRATION_UI=<value>
  #
  registration_ui: https://my-app.com/signup

  ## ORY Kratos Error UI URL ##
  #
  # URL where the ORY Kratos Error UI is hosted. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
  #
  # Examples:
  # - https://my-app.com/kratos-error
  # 
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_ERROR_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_ERROR_UI=<value>
  #
  error_ui: https://my-app.com/kratos-error

  ## Verify UI URL ##
  #
  # URL where the ORY Verify UI is hosted. This is the page where users activate and / or verify their email or telephone number. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
  #
  # Examples:
  # - https://my-app.com/verify
  # 
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_VERIFY_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_VERIFY_UI=<value>
  #
  verify_ui: https://my-app.com/verify

  ## Default Return To URL ##
  #
  # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. This value may be overridden by a `default_return_to` in a lower configuration level (`foo.bar.default_return_to` overrides `foo.default_return_to` overrides `default_return_to`) and by the `?return_to` query in certain cases.
  #
  # Examples:
  # - https://my-app.com/dashboard
  # 
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_DEFAULT_RETURN_TO=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_DEFAULT_RETURN_TO=<value>
  #
  default_return_to: https://my-app.com/dashboard

  ## self ##
  #
  self:
    
    ## public ##
    #
    # Examples:
    # - https://my-app.com/.ory/kratos/public
    # 
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export URLS_SELF_PUBLIC=<value>
    # - Windows Command Line (CMD):
    #    > set URLS_SELF_PUBLIC=<value>
    #
    public: https://my-app.com/.ory/kratos/public

    ## admin ##
    #
    # Examples:
    # - https://kratos.private-network:4434/
    # 
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export URLS_SELF_ADMIN=<value>
    # - Windows Command Line (CMD):
    #    > set URLS_SELF_ADMIN=<value>
    #
    admin: https://kratos.private-network:4434/

  ## Multi-Factor UI URL ##
  #
  # URL where the Multi-Factor UI is hosted. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
  #
  # Examples:
  # - https://my-app.com/login/mfa
  # 
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_MFA_UI=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_MFA_UI=<value>
  #
  mfa_ui: https://my-app.com/login/mfa

  ## Whitelisted Return To URLs ##
  #
  # List of URLs that are allowed to be redirected to. A redirection request is made by appending `?return_to=...` to Login, Registration, and other self-service flows.
  #
  # Examples:
  # - https://app.my-app.com/dashboard
  # - https://www.my-app.com/
  # 
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export URLS_WHITELISTED_RETURN_TO_URLS=<value>
  # - Windows Command Line (CMD):
  #    > set URLS_WHITELISTED_RETURN_TO_URLS=<value>
  #
  whitelisted_return_to_urls: https://www.my-app.com/

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
  level: debug

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
    - voluptate ullamco proident
    - sed aute commodo cupidatat

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
    memory: 1421164

    ## iterations ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_ITERATIONS=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_ITERATIONS=<value>
    #
    iterations: 61189230

    ## parallelism ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_PARALLELISM=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_PARALLELISM=<value>
    #
    parallelism: 96467155

    ## salt_length ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_SALT_LENGTH=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_SALT_LENGTH=<value>
    #
    salt_length: 61865753

    ## key_length ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_KEY_LENGTH=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_KEY_LENGTH=<value>
    #
    key_length: 14300065

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