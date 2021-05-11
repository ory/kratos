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

This reference configuration documents all keys, also deprecated ones! It is a
reference for all possible configuration values.

If you are looking for an example configuration, it is better to try out the
quickstart.

To find out more about edge cases like setting string array values through
environmental variables head to the
[Configuring ORY services](https://www.ory.sh/docs/ecosystem/configuring)
section.

```yaml
## Ory Kratos Configuration
#

## identity ##
#
identity:
  ## JSON Schema URL for default identity traits ##
  #
  # URL for JSON Schema which describes a default identity's traits. Can be a file path, a https URL, or a base64 encoded string. Will have ID: "default"
  #
  # Examples:
  # - file://path/to/identity.traits.schema.json
  # - https://foo.bar.com/path/to/identity.traits.schema.json
  # - base64://ewogICIkc2NoZW1hIjogImh0dHA6Ly9qc29uLXNjaGVtYS5vcmcvZHJhZnQtMDcvc2NoZW1hIyIsCiAgInR5cGUiOiAib2JqZWN0IiwKICAicHJvcGVydGllcyI6IHsKICAgICJiYXIiOiB7CiAgICAgICJ0eXBlIjogInN0cmluZyIKICAgIH0KICB9LAogICJyZXF1aXJlZCI6IFsKICAgICJiYXIiCiAgXQp9
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export IDENTITY_DEFAULT_SCHEMA_URL=<value>
  # - Windows Command Line (CMD):
  #    > set IDENTITY_DEFAULT_SCHEMA_URL=<value>
  #
  default_schema_url: file://path/to/identity.traits.schema.json

  ## Additional JSON Schemas for Identity Traits ##
  #
  # Examples:
  # - - id: customer
  #     url: base64://ewogICIkc2NoZW1hIjogImh0dHA6Ly9qc29uLXNjaGVtYS5vcmcvZHJhZnQtMDcvc2NoZW1hIyIsCiAgInR5cGUiOiAib2JqZWN0IiwKICAicHJvcGVydGllcyI6IHsKICAgICJiYXIiOiB7CiAgICAgICJ0eXBlIjogInN0cmluZyIKICAgIH0KICB9LAogICJyZXF1aXJlZCI6IFsKICAgICJiYXIiCiAgXQp9
  #   - id: employee
  #     url: https://foo.bar.com/path/to/employee.traits.schema.json
  #   - id: employee-v2
  #     url: file://path/to/employee.v2.traits.schema.json
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export IDENTITY_SCHEMAS=<value>
  # - Windows Command Line (CMD):
  #    > set IDENTITY_SCHEMAS=<value>
  #
  schemas:
    - id: customer
      url: base64://ewogICIkc2NoZW1hIjogImh0dHA6Ly9qc29uLXNjaGVtYS5vcmcvZHJhZnQtMDcvc2NoZW1hIyIsCiAgInR5cGUiOiAib2JqZWN0IiwKICAicHJvcGVydGllcyI6IHsKICAgICJiYXIiOiB7CiAgICAgICJ0eXBlIjogInN0cmluZyIKICAgIH0KICB9LAogICJyZXF1aXJlZCI6IFsKICAgICJiYXIiCiAgXQp9
    - id: employee
      url: https://foo.bar.com/path/to/employee.traits.schema.json
    - id: employee-v2
      url: file://path/to/employee.v2.traits.schema.json

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
  # Ory Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
  #
  # Examples:
  # - https://my-app.com/dashboard
  # - /dashboard
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export SELFSERVICE_DEFAULT_BROWSER_RETURN_URL=<value>
  # - Windows Command Line (CMD):
  #    > set SELFSERVICE_DEFAULT_BROWSER_RETURN_URL=<value>
  #
  default_browser_return_url: https://my-app.com/dashboard

  ## flows ##
  #
  flows:
    ## logout ##
    #
    logout:
      ## after ##
      #
      after:
        ## Redirect browsers to set URL per default ##
        #
        # Ory Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
        #
        # Examples:
        # - https://my-app.com/dashboard
        # - /dashboard
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
      ## lifespan ##
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
      #    $ export SELFSERVICE_FLOWS_REGISTRATION_LIFESPAN=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_REGISTRATION_LIFESPAN=<value>
      #
      lifespan: 1h

      ## before ##
      #
      before:
        ## hooks ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_FLOWS_REGISTRATION_BEFORE_HOOKS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_REGISTRATION_BEFORE_HOOKS=<value>
        #
        hooks:
          - hook: web_hook
            config:
              url: http://a.aaa
              method: ''
              auth:
                type: api_key
                config:
                  name: ''
                  value: ''
                  in: header
              body: ''

      ## after ##
      #
      after:
        ## password ##
        #
        password:
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

          ## Redirect browsers to set URL per default ##
          #
          # Ory Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
          #
          # Examples:
          # - https://my-app.com/dashboard
          # - /dashboard
          #
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_REGISTRATION_AFTER_PASSWORD_DEFAULT_BROWSER_RETURN_URL=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_REGISTRATION_AFTER_PASSWORD_DEFAULT_BROWSER_RETURN_URL=<value>
          #
          default_browser_return_url: https://my-app.com/dashboard

        ## oidc ##
        #
        oidc:
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

          ## Redirect browsers to set URL per default ##
          #
          # Ory Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
          #
          # Examples:
          # - https://my-app.com/dashboard
          # - /dashboard
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
        #    $ export SELFSERVICE_FLOWS_REGISTRATION_AFTER_HOOKS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_REGISTRATION_AFTER_HOOKS=<value>
        #
        hooks:
          - hook: web_hook
            config:
              url: http://a.aaa
              method: ''
              auth:
                type: api_key
                config:
                  name: ''
                  value: ''
                  in: header
              body: ''

        ## Redirect browsers to set URL per default ##
        #
        # Ory Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
        #
        # Examples:
        # - https://my-app.com/dashboard
        # - /dashboard
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_FLOWS_REGISTRATION_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_REGISTRATION_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        #
        default_browser_return_url: https://my-app.com/dashboard

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
      ui_url: https://my-app.com/signup

    ## login ##
    #
    login:
      ## lifespan ##
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
      #    $ export SELFSERVICE_FLOWS_LOGIN_LIFESPAN=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_LOGIN_LIFESPAN=<value>
      #
      lifespan: 1h

      ## before ##
      #
      before:
        ## hooks ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_FLOWS_LOGIN_BEFORE_HOOKS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_LOGIN_BEFORE_HOOKS=<value>
        #
        hooks:
          - hook: web_hook
            config:
              url: http://a.aaa
              method: ''
              auth:
                type: api_key
                config:
                  name: ''
                  value: ''
                  in: header
              body: ''

      ## after ##
      #
      after:
        ## password ##
        #
        password:
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

          ## Redirect browsers to set URL per default ##
          #
          # Ory Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
          #
          # Examples:
          # - https://my-app.com/dashboard
          # - /dashboard
          #
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_LOGIN_AFTER_PASSWORD_DEFAULT_BROWSER_RETURN_URL=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_LOGIN_AFTER_PASSWORD_DEFAULT_BROWSER_RETURN_URL=<value>
          #
          default_browser_return_url: https://my-app.com/dashboard

        ## oidc ##
        #
        oidc:
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

          ## Redirect browsers to set URL per default ##
          #
          # Ory Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
          #
          # Examples:
          # - https://my-app.com/dashboard
          # - /dashboard
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
        #    $ export SELFSERVICE_FLOWS_LOGIN_AFTER_HOOKS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_LOGIN_AFTER_HOOKS=<value>
        #
        hooks:
          - hook: web_hook
            config:
              url: http://a.aaa
              method: ''
              auth:
                type: api_key
                config:
                  name: ''
                  value: ''
                  in: header
              body: ''

        ## Redirect browsers to set URL per default ##
        #
        # Ory Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
        #
        # Examples:
        # - https://my-app.com/dashboard
        # - /dashboard
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_FLOWS_LOGIN_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_LOGIN_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        #
        default_browser_return_url: https://my-app.com/dashboard

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

    ## Email and Phone Verification and Account Activation Configuration ##
    #
    verification:
      ## Verify UI URL ##
      #
      # URL where the Ory Verify UI is hosted. This is the page where users activate and / or verify their email or telephone number. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
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
        ## hooks ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_FLOWS_VERIFICATION_AFTER_HOOKS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_VERIFICATION_AFTER_HOOKS=<value>
        #
        hooks:
          - hook: web_hook
            config:
              url: http://a.aaa
              method: ''
              auth:
                type: api_key
                config:
                  name: ''
                  value: ''
                  in: header
              body: ''

        ## Redirect browsers to set URL per default ##
        #
        # Ory Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
        #
        # Examples:
        # - https://my-app.com/dashboard
        # - /dashboard
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
      #    $ export SELFSERVICE_FLOWS_VERIFICATION_LIFESPAN=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_VERIFICATION_LIFESPAN=<value>
      #
      lifespan: 1h

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
      enabled: false

    ## Account Recovery Configuration ##
    #
    recovery:
      ## Recovery UI URL ##
      #
      # URL where the Ory Recovery UI is hosted. This is the page where users request and complete account recovery. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
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
      ui_url: https://my-app.com/verify

      ## after ##
      #
      after:
        ## hooks ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_FLOWS_RECOVERY_AFTER_HOOKS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_RECOVERY_AFTER_HOOKS=<value>
        #
        hooks:
          - hook: web_hook
            config:
              url: http://a.aaa
              method: ''
              auth:
                type: api_key
                config:
                  name: ''
                  value: ''
                  in: header
              body: ''

        ## Redirect browsers to set URL per default ##
        #
        # Ory Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
        #
        # Examples:
        # - https://my-app.com/dashboard
        # - /dashboard
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
      #    $ export SELFSERVICE_FLOWS_RECOVERY_LIFESPAN=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_RECOVERY_LIFESPAN=<value>
      #
      lifespan: 1h

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

    ## error ##
    #
    error:
      ## Ory Kratos Error UI URL ##
      #
      # URL where the Ory Kratos Error UI is hosted. Check the [reference implementation](https://github.com/ory/kratos-selfservice-ui-node).
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

    ## settings ##
    #
    settings:
      ## lifespan ##
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
      #    $ export SELFSERVICE_FLOWS_SETTINGS_LIFESPAN=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_FLOWS_SETTINGS_LIFESPAN=<value>
      #
      lifespan: 1h

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
        ## password ##
        #
        password:
          ## hooks ##
          #
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_SETTINGS_AFTER_PASSWORD_HOOKS=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_SETTINGS_AFTER_PASSWORD_HOOKS=<value>
          #
          hooks:
            - hook: web_hook
              config:
                url: http://a.aaa
                method: ''
                auth:
                  type: api_key
                  config:
                    name: ''
                    value: ''
                    in: header
                body: ''

          ## Redirect browsers to set URL per default ##
          #
          # Ory Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
          #
          # Examples:
          # - https://my-app.com/dashboard
          # - /dashboard
          #
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_SETTINGS_AFTER_PASSWORD_DEFAULT_BROWSER_RETURN_URL=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_SETTINGS_AFTER_PASSWORD_DEFAULT_BROWSER_RETURN_URL=<value>
          #
          default_browser_return_url: https://my-app.com/dashboard

        ## profile ##
        #
        profile:
          ## hooks ##
          #
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export SELFSERVICE_FLOWS_SETTINGS_AFTER_PROFILE_HOOKS=<value>
          # - Windows Command Line (CMD):
          #    > set SELFSERVICE_FLOWS_SETTINGS_AFTER_PROFILE_HOOKS=<value>
          #
          hooks:
            - hook: web_hook
              config:
                url: http://a.aaa
                method: ''
                auth:
                  type: api_key
                  config:
                    name: ''
                    value: ''
                    in: header
                body: ''

          ## Redirect browsers to set URL per default ##
          #
          # Ory Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
          #
          # Examples:
          # - https://my-app.com/dashboard
          # - /dashboard
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
        #    $ export SELFSERVICE_FLOWS_SETTINGS_AFTER_HOOKS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_SETTINGS_AFTER_HOOKS=<value>
        #
        hooks:
          - hook: web_hook
            config:
              url: http://a.aaa
              method: ''
              auth:
                type: api_key
                config:
                  name: ''
                  value: ''
                  in: header
              body: ''

        ## Redirect browsers to set URL per default ##
        #
        # Ory Kratos redirects to this URL per default on completion of self-service flows and other browser interaction. Read this [article for more information on browser redirects](https://www.ory.sh/kratos/docs/concepts/browser-redirect-flow-completion).
        #
        # Examples:
        # - https://my-app.com/dashboard
        # - /dashboard
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_FLOWS_SETTINGS_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_FLOWS_SETTINGS_AFTER_DEFAULT_BROWSER_RETURN_URL=<value>
        #
        default_browser_return_url: https://my-app.com/dashboard

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
      ui_url: https://my-app.com/user/settings

  ## methods ##
  #
  methods:
    ## link ##
    #
    link:
      ## Enables Link Method ##
      #
      # Default value: true
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_METHODS_LINK_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_METHODS_LINK_ENABLED=<value>
      #
      enabled: false

    ## password ##
    #
    password:
      ## Password Configuration ##
      #
      # Define how passwords are validated.
      #
      config:
        ## Allow Password Breaches ##
        #
        # Defines how often a password may have been breached before it is rejected.
        #
        # Minimum value: 0
        #
        # Maximum value: 100
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_METHODS_PASSWORD_CONFIG_MAX_BREACHES=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_METHODS_PASSWORD_CONFIG_MAX_BREACHES=<value>
        #
        max_breaches: 0

        ## Ignore Lookup Network Errors ##
        #
        # If set to false the password validation fails when the network or the Have I Been Pwnd API is down.
        #
        # Default value: true
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_METHODS_PASSWORD_CONFIG_IGNORE_NETWORK_ERRORS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_METHODS_PASSWORD_CONFIG_IGNORE_NETWORK_ERRORS=<value>
        #
        ignore_network_errors: false

        ## Custom haveibeenpwned host ##
        #
        # Allows changing the default HIBP host to a self hosted version.
        #
        # Default value: api.pwnedpasswords.com
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_METHODS_PASSWORD_CONFIG_HAVEIBEENPWNED_HOST=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_METHODS_PASSWORD_CONFIG_HAVEIBEENPWNED_HOST=<value>
        #
        haveibeenpwned_host: ''

      ## Enables Username/Email and Password Method ##
      #
      # Default value: true
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_METHODS_PASSWORD_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_METHODS_PASSWORD_ENABLED=<value>
      #
      enabled: false

    ## Specify OpenID Connect and OAuth2 Configuration ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SELFSERVICE_METHODS_OIDC=<value>
    # - Windows Command Line (CMD):
    #    > set SELFSERVICE_METHODS_OIDC=<value>
    #
    # This can be set as an environment variable by supplying it as a JSON object.
    #
    oidc:
      ## config ##
      #
      config:
        ## OpenID Connect and OAuth2 Providers ##
        #
        # A list and configuration of OAuth2 and OpenID Connect providers Ory Kratos should integrate with.
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SELFSERVICE_METHODS_OIDC_CONFIG_PROVIDERS=<value>
        # - Windows Command Line (CMD):
        #    > set SELFSERVICE_METHODS_OIDC_CONFIG_PROVIDERS=<value>
        #
        providers:
          - id: google
            provider: google
            client_id: ''
            client_secret: ''
            mapper_url: file://path/to/oidc.jsonnet
            issuer_url: https://accounts.google.com
            auth_url: https://accounts.google.com/o/oauth2/v2/auth
            token_url: https://www.googleapis.com/oauth2/v4/token
            scope:
              - offline_access
            tenant: common
            requested_claims:
              id_token:
                email:
                email_verified:
            label: ''

      ## Enables OpenID Connect Method ##
      #
      # Default value: false
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_METHODS_OIDC_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_METHODS_OIDC_ENABLED=<value>
      #
      enabled: false

    ## profile ##
    #
    profile:
      ## Enables Profile Management Method ##
      #
      # Default value: true
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SELFSERVICE_METHODS_PROFILE_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set SELFSERVICE_METHODS_PROFILE_ENABLED=<value>
      #
      enabled: false

  ## Whitelisted Return To URLs ##
  #
  # List of URLs that are allowed to be redirected to. A redirection request is made by appending `?return_to=...` to Login, Registration, and other self-service flows.
  #
  # Examples:
  # - - https://app.my-app.com/dashboard
  #   - /dashboard
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
    - /dashboard
    - https://www.my-app.com/

## serve ##
#
serve:
  ## public ##
  #
  public:
    ## Base URL ##
    #
    # The URL where the endpoint is exposed at. This domain is used to generate redirects, form URLs, and more.
    #
    # Examples:
    # - https://my-app.com/
    # - https://my-app.com/.ory/kratos/public
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_PUBLIC_BASE_URL=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_PUBLIC_BASE_URL=<value>
    #
    base_url: https://my-app.com/

    ## Domain Aliases ##
    #
    # Adds an alias domain. If a request with the hostname (FQDN) matching the hostname in the alias is found, that URL is used as the base URL.
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_PUBLIC_DOMAIN_ALIASES=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_PUBLIC_DOMAIN_ALIASES=<value>
    #
    domain_aliases:
      - match_domain: localhost
        base_path: /
        scheme: http

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
    host: ''

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

    ## socket ##
    #
    # Sets the permissions of the unix socket
    #
    socket:
      ## group ##
      #
      # Group of unix socket. If empty, the group will be the primary group of the user running Kratos.
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PUBLIC_SOCKET_GROUP=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PUBLIC_SOCKET_GROUP=<value>
      #
      group: ''

      ## mode ##
      #
      # Mode of unix socket in numeric form
      #
      # Default value: 493
      #
      # Minimum value: 0
      #
      # Maximum value: 511
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PUBLIC_SOCKET_MODE=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PUBLIC_SOCKET_MODE=<value>
      #
      mode: 0

      ## owner ##
      #
      # Owner of unix socket. If empty, the owner will be the user running Kratos.
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PUBLIC_SOCKET_OWNER=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PUBLIC_SOCKET_OWNER=<value>
      #
      owner: ''

    ## cors ##
    #
    # Configures Cross Origin Resource Sharing for public endpoints.
    #
    cors:
      ## allowed_origins ##
      #
      # A list of origins a cross-domain request can be executed from. If the special * value is present in the list, all origins will be allowed. An origin may contain a wildcard (*) to replace 0 or more characters (i.e.: http://*.domain.com). Only one wildcard can be used per origin.
      #
      # Default value: *
      #
      # Examples:
      # - - https://example.com
      #   - https://*.example.com
      #   - https://*.foo.example.com
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PUBLIC_CORS_ALLOWED_ORIGINS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PUBLIC_CORS_ALLOWED_ORIGINS=<value>
      #
      allowed_origins:
        - https://example.com
        - https://*.example.com
        - https://*.foo.example.com

      ## allowed_methods ##
      #
      # A list of HTTP methods the user agent is allowed to use with cross-domain requests.
      #
      # Default value: POST,GET,PUT,PATCH,DELETE
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PUBLIC_CORS_ALLOWED_METHODS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PUBLIC_CORS_ALLOWED_METHODS=<value>
      #
      allowed_methods:
        - POST

      ## allowed_headers ##
      #
      # A list of non simple headers the client is allowed to use with cross-domain requests.
      #
      # Default value: Authorization,Content-Type,X-Session-Token
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PUBLIC_CORS_ALLOWED_HEADERS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PUBLIC_CORS_ALLOWED_HEADERS=<value>
      #
      allowed_headers:
        - ''

      ## exposed_headers ##
      #
      # Sets which headers are safe to expose to the API of a CORS API specification.
      #
      # Default value: Content-Type
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PUBLIC_CORS_EXPOSED_HEADERS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PUBLIC_CORS_EXPOSED_HEADERS=<value>
      #
      exposed_headers:
        - ''

      ## allow_credentials ##
      #
      # Sets whether the request can include user credentials like cookies, HTTP authentication or client side SSL certificates.
      #
      # Default value: true
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PUBLIC_CORS_ALLOW_CREDENTIALS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PUBLIC_CORS_ALLOW_CREDENTIALS=<value>
      #
      allow_credentials: false

      ## options_passthrough ##
      #
      # TODO
      #
      # Default value: false
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PUBLIC_CORS_OPTIONS_PASSTHROUGH=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PUBLIC_CORS_OPTIONS_PASSTHROUGH=<value>
      #
      options_passthrough: false

      ## max_age ##
      #
      # Sets how long (in seconds) the results of a preflight request can be cached. If set to 0, every request is preceded by a preflight request.
      #
      # Minimum value: 0
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PUBLIC_CORS_MAX_AGE=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PUBLIC_CORS_MAX_AGE=<value>
      #
      max_age: 0

      ## debug ##
      #
      # Adds additional log output to debug server side CORS issues.
      #
      # Default value: false
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PUBLIC_CORS_DEBUG=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PUBLIC_CORS_DEBUG=<value>
      #
      debug: false

      ## enabled ##
      #
      # Sets whether CORS is enabled.
      #
      # Default value: false
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PUBLIC_CORS_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PUBLIC_CORS_ENABLED=<value>
      #
      enabled: false

  ## admin ##
  #
  admin:
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
    host: ''

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

    ## socket ##
    #
    # Sets the permissions of the unix socket
    #
    socket:
      ## group ##
      #
      # Group of unix socket. If empty, the group will be the primary group of the user running Kratos.
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_ADMIN_SOCKET_GROUP=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_ADMIN_SOCKET_GROUP=<value>
      #
      group: ''

      ## mode ##
      #
      # Mode of unix socket in numeric form
      #
      # Default value: 493
      #
      # Minimum value: 0
      #
      # Maximum value: 511
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_ADMIN_SOCKET_MODE=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_ADMIN_SOCKET_MODE=<value>
      #
      mode: 0

      ## owner ##
      #
      # Owner of unix socket. If empty, the owner will be the user running Kratos.
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_ADMIN_SOCKET_OWNER=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_ADMIN_SOCKET_OWNER=<value>
      #
      owner: ''

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

## tracing ##
#
# Ory Kratos supports distributed tracing.
#
tracing:
  ## service_name ##
  #
  # Specifies the service name to use on the tracer.
  #
  # Examples:
  # - Ory Kratos
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export TRACING_SERVICE_NAME=<value>
  # - Windows Command Line (CMD):
  #    > set TRACING_SERVICE_NAME=<value>
  #
  service_name: Ory Kratos

  ## providers ##
  #
  providers:
    ## zipkin ##
    #
    # Configures the zipkin tracing backend.
    #
    # Examples:
    # - server_url: http://localhost:9411/api/v2/spans
    #
    zipkin:
      ## server_url ##
      #
      # The address of Zipkin server where spans should be sent to.
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export TRACING_PROVIDERS_ZIPKIN_SERVER_URL=<value>
      # - Windows Command Line (CMD):
      #    > set TRACING_PROVIDERS_ZIPKIN_SERVER_URL=<value>
      #
      server_url: http://localhost:9411/api/v2/spans

    ## jaeger ##
    #
    # Configures the jaeger tracing backend.
    #
    jaeger:
      ## propagation ##
      #
      # The tracing header format
      #
      # Examples:
      # - jaeger
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export TRACING_PROVIDERS_JAEGER_PROPAGATION=<value>
      # - Windows Command Line (CMD):
      #    > set TRACING_PROVIDERS_JAEGER_PROPAGATION=<value>
      #
      propagation: jaeger

      ## sampling ##
      #
      # Examples:
      # - type: const
      #   value: 1
      #   server_url: http://localhost:5778/sampling
      #
      sampling:
        ## type ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export TRACING_PROVIDERS_JAEGER_SAMPLING_TYPE=<value>
        # - Windows Command Line (CMD):
        #    > set TRACING_PROVIDERS_JAEGER_SAMPLING_TYPE=<value>
        #
        type: const

        ## value ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export TRACING_PROVIDERS_JAEGER_SAMPLING_VALUE=<value>
        # - Windows Command Line (CMD):
        #    > set TRACING_PROVIDERS_JAEGER_SAMPLING_VALUE=<value>
        #
        value: 1

        ## server_url ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export TRACING_PROVIDERS_JAEGER_SAMPLING_SERVER_URL=<value>
        # - Windows Command Line (CMD):
        #    > set TRACING_PROVIDERS_JAEGER_SAMPLING_SERVER_URL=<value>
        #
        server_url: http://localhost:5778/sampling

      ## local_agent_address ##
      #
      # The address of the jaeger-agent where spans should be sent to.
      #
      # Examples:
      # - 127.0.0.1:6831
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export TRACING_PROVIDERS_JAEGER_LOCAL_AGENT_ADDRESS=<value>
      # - Windows Command Line (CMD):
      #    > set TRACING_PROVIDERS_JAEGER_LOCAL_AGENT_ADDRESS=<value>
      #
      local_agent_address: 127.0.0.1:6831

  ## provider ##
  #
  # Set this to the tracing backend you wish to use. Supports Jaeger, Zipkin and DataDog. If omitted or empty, tracing will be disabled. Use environment variables to configure DataDog (see https://docs.datadoghq.com/tracing/setup/go/#configuration).
  #
  # One of:
  # - jaeger
  # - zipkin
  # - datadog
  # - elastic-apm
  #
  # Examples:
  # - jaeger
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export TRACING_PROVIDER=<value>
  # - Windows Command Line (CMD):
  #    > set TRACING_PROVIDER=<value>
  #
  provider: jaeger

## Log ##
#
# Configure logging using the following options. Logging will always be sent to stdout and stderr.
#
log:
  ## Leak Sensitive Log Values ##
  #
  # If set will leak sensitive values (e.g. emails) in the logs.
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export LOG_LEAK_SENSITIVE_VALUES=<value>
  # - Windows Command Line (CMD):
  #    > set LOG_LEAK_SENSITIVE_VALUES=<value>
  #
  leak_sensitive_values: false

  ## format ##
  #
  # The log format can either be text or JSON.
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
  format: json

  ## level ##
  #
  # Debug enables stack traces on errors. Can also be set using environment variable LOG_LEVEL.
  #
  # Default value: info
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
  level: trace

## secrets ##
#
secrets:
  ## Signing Keys for Cookies ##
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
    - ipsumipsumipsumi

  ## Default Encryption Signing Secrets ##
  #
  # The first secret in the array is used for signing and encrypting things while all other keys are used to verify and decrypt older things that were signed with that old secret.
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export SECRETS_DEFAULT=<value>
  # - Windows Command Line (CMD):
  #    > set SECRETS_DEFAULT=<value>
  #
  default:
    - ipsumipsumipsumi

## Hashing Algorithm Configuration ##
#
hashers:
  ## Configuration for the Argon2id hasher. ##
  #
  argon2:
    ## iterations ##
    #
    # Default value: 1
    #
    # Minimum value: 1
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_ITERATIONS=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_ITERATIONS=<value>
    #
    iterations: 1

    ## parallelism ##
    #
    # Number of parallel workers, defaults to 2*runtime.NumCPU().
    #
    # Minimum value: 1
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_PARALLELISM=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_PARALLELISM=<value>
    #
    parallelism: 1

    ## salt_length ##
    #
    # Default value: 16
    #
    # Minimum value: 16
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_SALT_LENGTH=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_SALT_LENGTH=<value>
    #
    salt_length: 16

    ## key_length ##
    #
    # Default value: 32
    #
    # Minimum value: 16
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_KEY_LENGTH=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_KEY_LENGTH=<value>
    #
    key_length: 16

    ## expected_duration ##
    #
    # The time a hashing operation (~login latency) should take.
    #
    # Default value: 500ms
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_EXPECTED_DURATION=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_EXPECTED_DURATION=<value>
    #
    expected_duration: 0ns

    ## expected_deviation ##
    #
    # The standard deviation expected for hashing operations. If this value is exceeded you will be warned in the logs to adjust the parameters.
    #
    # Default value: 500ms
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_EXPECTED_DEVIATION=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_EXPECTED_DEVIATION=<value>
    #
    expected_deviation: 0ns

    ## dedicated_memory ##
    #
    # The memory dedicated for Kratos. As password hashing is very resource intense, Kratos will monitor the memory consumption and warn about high values.
    #
    # Default value: 1GB
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_DEDICATED_MEMORY=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_DEDICATED_MEMORY=<value>
    #
    dedicated_memory: 0B

    ## memory ##
    #
    # Default value: 128MB
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_ARGON2_MEMORY=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_ARGON2_MEMORY=<value>
    #
    memory: 0B

  ## Configuration for the Bcrypt hasher. Minimum is 4 when --dev flag is used and 12 otherwise. ##
  #
  bcrypt:
    ## cost ##
    #
    # Default value: 12
    #
    # Minimum value: 4
    #
    # Maximum value: 31
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export HASHERS_BCRYPT_COST=<value>
    # - Windows Command Line (CMD):
    #    > set HASHERS_BCRYPT_COST=<value>
    #
    cost: 4

  ## Password hashing algorithm ##
  #
  # One of the values: argon2, bcrypt
  #
  # Default value: bcrypt
  #
  # One of:
  # - argon2
  # - bcrypt
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export HASHERS_ALGORITHM=<value>
  # - Windows Command Line (CMD):
  #    > set HASHERS_ALGORITHM=<value>
  #
  algorithm: argon2

## session ##
#
session:
  ## cookie ##
  #
  cookie:
    ## Session Cookie Name ##
    #
    # Sets the session cookie name. Use with care!
    #
    # Default value: ory_kratos_session
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SESSION_COOKIE_NAME=<value>
    # - Windows Command Line (CMD):
    #    > set SESSION_COOKIE_NAME=<value>
    #
    name: ''

    ## Make Session Cookie Persistent ##
    #
    # If set to true will persist the cookie in the end-user's browser using the `max-age` parameter which is set to the `session.lifespan` value. Persistent cookies are not deleted when the browser is closed (e.g. on reboot or alt+f4).
    #
    # Default value: true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SESSION_COOKIE_PERSISTENT=<value>
    # - Windows Command Line (CMD):
    #    > set SESSION_COOKIE_PERSISTENT=<value>
    #
    persistent: false

    ## Session Cookie Path ##
    #
    # Sets the session cookie path. Use with care!
    #
    # Default value: /
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SESSION_COOKIE_PATH=<value>
    # - Windows Command Line (CMD):
    #    > set SESSION_COOKIE_PATH=<value>
    #
    path: ''

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
    same_site: Strict

    ## Session Cookie Domain ##
    #
    # Sets the session cookie domain. Useful when dealing with subdomains. Use with care!
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SESSION_COOKIE_DOMAIN=<value>
    # - Windows Command Line (CMD):
    #    > set SESSION_COOKIE_DOMAIN=<value>
    #
    domain: ''

  ## Session Lifespan ##
  #
  # Defines how long a session is active. Once that lifespan has been reached, the user needs to sign in again.
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
  #    $ export SESSION_LIFESPAN=<value>
  # - Windows Command Line (CMD):
  #    > set SESSION_LIFESPAN=<value>
  #
  lifespan: 1h

## The kratos version this config is written for. ##
#
# SemVer according to https://semver.org/ prefixed with `v` as in our releases.
#
# Examples:
# - v0.5.0-alpha.1
#
# Set this value using environment variables on
# - Linux/macOS:
#    $ export VERSION=<value>
# - Windows Command Line (CMD):
#    > set VERSION=<value>
#
version: v0.5.0-alpha.1

## dev ##
#
# Set this value using environment variables on
# - Linux/macOS:
#    $ export DEV=<value>
# - Windows Command Line (CMD):
#    > set DEV=<value>
#
dev: false

## help ##
#
# Set this value using environment variables on
# - Linux/macOS:
#    $ export HELP=<value>
# - Windows Command Line (CMD):
#    > set HELP=<value>
#
help: false

## sqa-opt-out ##
#
# Default value: false
#
# Set this value using environment variables on
# - Linux/macOS:
#    $ export SQA-OPT-OUT=<value>
# - Windows Command Line (CMD):
#    > set SQA-OPT-OUT=<value>
#
sqa-opt-out: false

## watch-courier ##
#
# Default value: false
#
# Set this value using environment variables on
# - Linux/macOS:
#    $ export WATCH-COURIER=<value>
# - Windows Command Line (CMD):
#    > set WATCH-COURIER=<value>
#
watch-courier: false

## Metrics port ##
#
# The port the courier's metrics endpoint listens on (0/disabled by default).
#
# Minimum value: 0
#
# Maximum value: 65535
#
# Examples:
# - 4434
#
# Set this value using environment variables on
# - Linux/macOS:
#    $ export EXPOSE-METRICS-PORT=<value>
# - Windows Command Line (CMD):
#    > set EXPOSE-METRICS-PORT=<value>
#
expose-metrics-port: 4434

## config ##
#
# Set this value using environment variables on
# - Linux/macOS:
#    $ export CONFIG=<value>
# - Windows Command Line (CMD):
#    > set CONFIG=<value>
#
config:
  - ''

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

    ## SMTP Sender Name ##
    #
    # The recipient of an email will see this as the sender name.
    #
    # Examples:
    # - Bob
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export COURIER_SMTP_FROM_NAME=<value>
    # - Windows Command Line (CMD):
    #    > set COURIER_SMTP_FROM_NAME=<value>
    #
    from_name: Bob

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
    from_address: aaa@a.aa

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
