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


## identity ##
#
identity:
  
  ## JSON Schema URL for default identity traits ##
  #
  # Path to the JSON Schema which describes a default identity's traits.
  #
  # Examples:
  # - file://path/to/identity.traits.schema.json
  # - https://foo.bar.com/path/to/identity.traits.schema.json
  # 
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export IDENTITY_DEFAULT_SCHEMA_URL=<value>
  # - Windows Command Line (CMD):
  #    > set IDENTITY_DEFAULT_SCHEMA_URL=<value>
  #
  default_schema_url: https://foo.bar.com/path/to/identity.traits.schema.json

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
  #    $ export IDENTITY_SCHEMAS=<value>
  # - Windows Command Line (CMD):
  #    > set IDENTITY_SCHEMAS=<value>
  #
  schemas:
    - id: customer
      url: https://foo.bar.com/path/to/customer.traits.schema.json
    - id: employee
      url: https://foo.bar.com/path/to/employee.traits.schema.json
    - id: employee-v2
      url: https://foo.bar.com/path/to/employee.v2.traits.schema.json

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

## selfservice ##
#
selfservice:
  
  ## Redirect browsers to set URL per default ##
  #
  # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
  #
  # Examples:
  # - https://my-app.com/dashboard
  # 
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export SELFSERVICE_DEFAULT_BROWSER_RETURN_URL=<value>
  # - Windows Command Line (CMD):
  #    > set SELFSERVICE_DEFAULT_BROWSER_RETURN_URL=<value>
  #
  default_browser_return_url: https://my-app.com/dashboard

  ## Whitelisted Return To URLs ##
  #
  # List of URLs that are allowed to be redirected to. A redirection request is made by appending `?return_to=...` to Login, Registration, and other self-service flows.
  #
  # Examples:
  # - - https://app.my-app.com/dashboard
  #   - https://www.my-app.com/
  # 
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export SELFSERVICE_WHITELISTED_RETURN_URLS=<value>
  # - Windows Command Line (CMD):
  #    > set SELFSERVICE_WHITELISTED_RETURN_URLS=<value>
  #
  whitelisted_return_urls:
    - https://app.my-app.com/dashboard
    - https://www.my-app.com/

  ## flows ##
  #
  flows:
    
    ## settings ##
    #
    settings:
      
      ## URL of the Settings page. ##
      #
      # URL where the Settings UI is hosted. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
      #
      # Default value: https://www.ory.sh/kratos/docs/fallback/settings
      #
      # Examples:
      # - https://my-app.com/user/settings
      # 
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_FLOWS_SETTINGS_UI_URL=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_SETTINGS_UI_URL=<value>
      #
      ui_url: https://www.ory.sh/kratos/docs/fallback/settings

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
      #    $ export SELFSERVICE_FLOWS_SETTINGS_REQUEST_LIFESPAN=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_SETTINGS_REQUEST_LIFESPAN=<value>
      #
      request_lifespan: 1h

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
      #    $ export SELFSERVICE_FLOWS_SETTINGS_PRIVILEGED_SESSION_MAX_AGE=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_SETTINGS_PRIVILEGED_SESSION_MAX_AGE=<value>
      #
      privileged_session_max_age: 1h

      ## after ##
      #
      after:
        
        ## Redirect browsers to set URL per default ##
        #
        # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
        #
        # Examples:
        # - https://my-app.com/dashboard
        # 
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_FLOWS_SETTINGS_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_SETTINGS_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        #
        default_browser_return_url: https://my-app.com/dashboard

        ## password ##
        #
        password:
          
          ## Redirect browsers to set URL per default ##
          #
          # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
          #
          # Examples:
          # - https://my-app.com/dashboard
          # 
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_SETTINGS_AFTER_PASSWORD_DEFAULT_BROWSER_RETURN_URL=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_SETTINGS_AFTER_PASSWORD_DEFAULT_BROWSER_RETURN_URL=<value>
          #
          default_browser_return_url: https://my-app.com/dashboard

          ## hooks ##
          #
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_SETTINGS_AFTER_PASSWORD_HOOKS=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_SETTINGS_AFTER_PASSWORD_HOOKS=<value>
          #
          hooks:
            - hook: verify

        ## profile ##
        #
        profile:
          
          ## Redirect browsers to set URL per default ##
          #
          # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
          #
          # Examples:
          # - https://my-app.com/dashboard
          # 
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_SETTINGS_AFTER_PROFILE_DEFAULT_BROWSER_RETURN_URL=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_SETTINGS_AFTER_PROFILE_DEFAULT_BROWSER_RETURN_URL=<value>
          #
          default_browser_return_url: https://my-app.com/dashboard

          ## hooks ##
          #
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_SETTINGS_AFTER_PROFILE_HOOKS=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_SETTINGS_AFTER_PROFILE_HOOKS=<value>
          #
          hooks:
            - hook: verify

    ## logout ##
    #
    logout:
      
      ## after ##
      #
      after:
        
        ## Redirect browsers to set URL per default ##
        #
        # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
        #
        # Examples:
        # - https://my-app.com/dashboard
        # 
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_FLOWS_LOGOUT_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_LOGOUT_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        #
        default_browser_return_url: https://my-app.com/dashboard

    ## registration ##
    #
    registration:
      
      ## Registration UI URL ##
      #
      # URL where the Registration UI is hosted. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
      #
      # Default value: https://www.ory.sh/kratos/docs/fallback/registration
      #
      # Examples:
      # - https://my-app.com/signup
      # 
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_FLOWS_REGISTRATION_UI_URL=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_REGISTRATION_UI_URL=<value>
      #
      ui_url: https://www.ory.sh/kratos/docs/fallback/registration

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
      #    $ export SELFSERVICE_FLOWS_REGISTRATION_REQUEST_LIFESPAN=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_REGISTRATION_REQUEST_LIFESPAN=<value>
      #
      request_lifespan: 1m

      ## after ##
      #
      after:
        
        ## Redirect browsers to set URL per default ##
        #
        # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
        #
        # Examples:
        # - https://my-app.com/dashboard
        # 
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_FLOWS_REGISTRATION_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_REGISTRATION_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        #
        default_browser_return_url: https://my-app.com/dashboard

        ## password ##
        #
        password:
          
          ## Redirect browsers to set URL per default ##
          #
          # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
          #
          # Examples:
          # - https://my-app.com/dashboard
          # 
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_REGISTRATION_AFTER_PASSWORD_DEFAULT_BROWSER_RETURN_URL=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_REGISTRATION_AFTER_PASSWORD_DEFAULT_BROWSER_RETURN_URL=<value>
          #
          default_browser_return_url: https://my-app.com/dashboard

          ## hooks ##
          #
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_REGISTRATION_AFTER_PASSWORD_HOOKS=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_REGISTRATION_AFTER_PASSWORD_HOOKS=<value>
          #
          hooks:
            - hook: session

        ## oidc ##
        #
        oidc:
          
          ## Redirect browsers to set URL per default ##
          #
          # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
          #
          # Examples:
          # - https://my-app.com/dashboard
          # 
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_REGISTRATION_AFTER_OIDC_DEFAULT_BROWSER_RETURN_URL=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_REGISTRATION_AFTER_OIDC_DEFAULT_BROWSER_RETURN_URL=<value>
          #
          default_browser_return_url: https://my-app.com/dashboard

          ## hooks ##
          #
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_REGISTRATION_AFTER_OIDC_HOOKS=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_REGISTRATION_AFTER_OIDC_HOOKS=<value>
          #
          hooks:
            - hook: session

    ## login ##
    #
    login:
      
      ## Login UI URL ##
      #
      # URL where the Login UI is hosted. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
      #
      # Default value: https://www.ory.sh/kratos/docs/fallback/login
      #
      # Examples:
      # - https://my-app.com/login
      # 
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_FLOWS_LOGIN_UI_URL=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_LOGIN_UI_URL=<value>
      #
      ui_url: https://my-app.com/login

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
      #    $ export SELFSERVICE_FLOWS_LOGIN_REQUEST_LIFESPAN=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_LOGIN_REQUEST_LIFESPAN=<value>
      #
      request_lifespan: 1h

      ## after ##
      #
      after:
        
        ## Redirect browsers to set URL per default ##
        #
        # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
        #
        # Examples:
        # - https://my-app.com/dashboard
        # 
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_FLOWS_LOGIN_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_LOGIN_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        #
        default_browser_return_url: https://my-app.com/dashboard

        ## password ##
        #
        password:
          
          ## Redirect browsers to set URL per default ##
          #
          # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
          #
          # Examples:
          # - https://my-app.com/dashboard
          # 
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_LOGIN_AFTER_PASSWORD_DEFAULT_BROWSER_RETURN_URL=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_LOGIN_AFTER_PASSWORD_DEFAULT_BROWSER_RETURN_URL=<value>
          #
          default_browser_return_url: https://my-app.com/dashboard

          ## hooks ##
          #
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_LOGIN_AFTER_PASSWORD_HOOKS=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_LOGIN_AFTER_PASSWORD_HOOKS=<value>
          #
          hooks:
            - hook: revoke_active_sessions

        ## oidc ##
        #
        oidc:
          
          ## Redirect browsers to set URL per default ##
          #
          # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
          #
          # Examples:
          # - https://my-app.com/dashboard
          # 
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_LOGIN_AFTER_OIDC_DEFAULT_BROWSER_RETURN_URL=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_LOGIN_AFTER_OIDC_DEFAULT_BROWSER_RETURN_URL=<value>
          #
          default_browser_return_url: https://my-app.com/dashboard

          ## hooks ##
          #
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_LOGIN_AFTER_OIDC_HOOKS=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_LOGIN_AFTER_OIDC_HOOKS=<value>
          #
          hooks:
            - hook: revoke_active_sessions

    ## Email and Phone Verification and Account Activation Configuration ##
    #
    verification:
      
      ## Enable Email/Phone Verification ##
      #
      # If set to true will enable [Email and Phone Verification and Account Activation](https://www.ory.sh/kratos/docs/self-service/flows/verify-email-account-activation/).
      #
      # Default value: false
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_FLOWS_VERIFICATION_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_VERIFICATION_ENABLED=<value>
      #
      enabled: true

      ## Verify UI URL ##
      #
      # URL where the ORY Verify UI is hosted. This is the page where users activate and / or verify their email or telephone number. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
      #
      # Default value: https://www.ory.sh/kratos/docs/fallback/verification
      #
      # Examples:
      # - https://my-app.com/verify
      # 
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_FLOWS_VERIFICATION_UI_URL=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_VERIFICATION_UI_URL=<value>
      #
      ui_url: https://my-app.com/verify

      ## after ##
      #
      after:
        
        ## Redirect browsers to set URL per default ##
        #
        # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
        #
        # Examples:
        # - https://my-app.com/dashboard
        # 
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_FLOWS_VERIFICATION_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_VERIFICATION_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        #
        default_browser_return_url: https://my-app.com/dashboard

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
      #    $ export SELFSERVICE_FLOWS_VERIFICATION_REQUEST_LIFESPAN=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_VERIFICATION_REQUEST_LIFESPAN=<value>
      #
      request_lifespan: 1m

    ## Account Recovery Configuration ##
    #
    recovery:
      
      ## Enable Account Recovery ##
      #
      # If set to true will enable [Account Recovery](https://www.ory.sh/kratos/docs/self-service/flows/password-reset-account-recovery/).
      #
      # Default value: false
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_FLOWS_RECOVERY_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_RECOVERY_ENABLED=<value>
      #
      enabled: false

      ## Recovery UI URL ##
      #
      # URL where the ORY Recovery UI is hosted. This is the page where users request and complete account recovery. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
      #
      # Default value: https://www.ory.sh/kratos/docs/fallback/recovery
      #
      # Examples:
      # - https://my-app.com/verify
      # 
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_FLOWS_RECOVERY_UI_URL=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_RECOVERY_UI_URL=<value>
      #
      ui_url: https://www.ory.sh/kratos/docs/fallback/recovery

      ## after ##
      #
      after:
        
        ## Redirect browsers to set URL per default ##
        #
        # ORY Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
        #
        # Examples:
        # - https://my-app.com/dashboard
        # 
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_FLOWS_RECOVERY_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_RECOVERY_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        #
        default_browser_return_url: https://my-app.com/dashboard

      ## Self-Service Recovery Request Lifespan ##
      #
      # Sets how long the recovery request is valid. If expired, the user has to redo the flow.
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
      #    $ export SELFSERVICE_FLOWS_RECOVERY_REQUEST_LIFESPAN=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_RECOVERY_REQUEST_LIFESPAN=<value>
      #
      request_lifespan: 1s

    ## error ##
    #
    error:
      
      ## ORY Kratos Error UI URL ##
      #
      # URL where the ORY Kratos Error UI is hosted. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
      #
      # Default value: https://www.ory.sh/kratos/docs/fallback/error
      #
      # Examples:
      # - https://my-app.com/kratos-error
      # 
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_FLOWS_ERROR_UI_URL=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_ERROR_UI_URL=<value>
      #
      ui_url: https://my-app.com/kratos-error

  ## strategies ##
  #
  strategies:
    
    ## profile ##
    #
    profile:
      
      ## Enables Profile Management Strategy ##
      #
      # Default value: true
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_STRATEGIES_PROFILE_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_STRATEGIES_PROFILE_ENABLED=<value>
      #
      enabled: false

    ## recovery_token ##
    #
    recovery_token:
      
      ## Enables Token-based Account Recovery Strategy ##
      #
      # Default value: true
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_STRATEGIES_RECOVERY_TOKEN_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_STRATEGIES_RECOVERY_TOKEN_ENABLED=<value>
      #
      enabled: false

    ## password ##
    #
    password:
      
      ## Enables Password Strategy ##
      #
      # Default value: true
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
      
      ## Enables OpenID Connect Strategy ##
      #
      # Default value: false
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
            client_id: sit
            client_secret: adipisicing ut in sit
            mapper_url: base64://bG9jYWwgc3ViamVjdCA9I...
            issuer_url: https://accounts.google.com
            auth_url: https://accounts.google.com/o/oauth2/v2/auth
            token_url: https://www.googleapis.com/oauth2/v4/token
            scope:
              - profile
            tenant: contoso.onmicrosoft.com
          - id: google
            provider: google
            client_id: id velit sit
            client_secret: ad cupidatat laboris sunt
            mapper_url: https://foo.bar.com/path/to/oidc.jsonnet
            issuer_url: https://accounts.google.com
            auth_url: https://accounts.google.com/o/oauth2/v2/auth
            token_url: https://www.googleapis.com/oauth2/v4/token
            scope:
              - offline_access
            tenant: contoso.onmicrosoft.com

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
    from_address: V7cf2ygb0H@dgtERYRNqoqpVwWZSvoSvpnX.hn

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
    
    ## Admin Base URL ##
    #
    # The URL where the admin endpoint is exposed at.
    #
    # Examples:
    # - https://kratos.private-network:4434/
    # 
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_ADMIN_BASE_URL=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_ADMIN_BASE_URL=<value>
    #
    base_url: https://kratos.private-network:4434/

    ## Admin Host ##
    #
    # The host (interface) kratos' admin endpoint listens on.
    #
    # Default value: 0.0.0.0
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_ADMIN_HOST=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_ADMIN_HOST=<value>
    #
    host: dolore dolor consequat officia cupidatat

    ## Admin Port ##
    #
    # The port kratos' admin endpoint listens on.
    #
    # Default value: 4434
    #
    # Minimum value: 1
    #
    # Maximum value: 65535
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
    
    ## Public Base URL ##
    #
    # The URL where the public endpoint is exposed at.
    #
    # Examples:
    # - https://my-app.com/.ory/kratos/public
    # 
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_PUBLIC_BASE_URL=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_PUBLIC_BASE_URL=<value>
    #
    base_url: https://my-app.com/.ory/kratos/public

    ## Public Host ##
    #
    # The host (interface) kratos' public endpoint listens on.
    #
    # Default value: 0.0.0.0
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_PUBLIC_HOST=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_PUBLIC_HOST=<value>
    #
    host: et

    ## Public Port ##
    #
    # The port kratos' public endpoint listens on.
    #
    # Default value: 4433
    #
    # Minimum value: 1
    #
    # Maximum value: 65535
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

## log ##
#
log:
  
  ## level ##
  #
  # One of:
  # - trace
  # - debug
  # - info
  # - warning
  # - error
  # - fatal
  # - panic
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
  # One of:
  # - json
  # - text
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
  
  ## Default Encryption Signing Secrets ##
  #
  # The first secret in the array is used for singing and encrypting things while all other keys are used to verify and decrypt older things that were signed with that old secret.
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export SECRETS_DEFAULT=<value>
  # - Windows Command Line (CMD):
  #    > set SECRETS_DEFAULT=<value>
  #
  default:
    - mollit voluptate
    - magnaaute irure voluptate aliqua dolor
    - et voluptate qui ipsum
    - nulla voluptate non

  ## Singing Keys for Cookies ##
  #
  # The first secret in the array is used for encrypting cookies while all other keys are used to decrypt older cookies that were signed with that old secret.
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export SECRETS_COOKIE=<value>
  # - Windows Command Line (CMD):
  #    > set SECRETS_COOKIE=<value>
  #
  cookie:
    - dolor consequat in commodo ut
    - culpa enim incididunt
    - non eiusmod amet
    - Excepteur aliquip in nostrud
    - veniam cupidatat reprehenderit

## Hashing Algorithm Configuration ##
#
hashers:
  
  ## Configuration for the Argon2id hasher. ##
  #
  argon2:
    
    ## memory ##
    #
    # Minimum value: 16384
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_MEMORY=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_MEMORY=<value>
    #
    memory: 9923292

    ## iterations ##
    #
    # Minimum value: 1
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_ITERATIONS=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_ITERATIONS=<value>
    #
    iterations: 93455383

    ## parallelism ##
    #
    # Minimum value: 1
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_PARALLELISM=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_PARALLELISM=<value>
    #
    parallelism: 57113235

    ## salt_length ##
    #
    # Minimum value: 16
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_SALT_LENGTH=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_SALT_LENGTH=<value>
    #
    salt_length: 87632070

    ## key_length ##
    #
    # Minimum value: 16
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_KEY_LENGTH=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_KEY_LENGTH=<value>
    #
    key_length: 3894910

## session ##
#
session:
  
  ## Session Lifespan ##
  #
  # Defines how long a session is active. This value is ignored if the "remember me" feature is used.
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
  #    $ export SESSION_LIFESPAN=<value>
  # - Windows Command Line (CMD):
  #    > set SESSION_LIFESPAN=<value>
  #
  lifespan: 1h

  ## Cookie Same Site Configuration ##
  #
  # Default value: Lax
  #
  # One of:
  # - Strict
  # - Lax
  # - None
  # 
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export SESSION_COOKIE_SAME_SITE=<value>
  # - Windows Command Line (CMD):
  #    > set SESSION_COOKIE_SAME_SITE=<value>
  #
  cookie_same_site: Lax

```