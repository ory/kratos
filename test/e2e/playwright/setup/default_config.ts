// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { OryKratosConfiguration } from "../../cypress/support/config"

export const default_config: OryKratosConfiguration = {
  dsn: "",
  identity: {
    schemas: [
      {
        id: "default",
        url: "file://test/e2e/profiles/oidc/identity.traits.schema.json",
      },
    ],
  },
  serve: {
    public: {
      base_url: "http://localhost:4455/",
      cors: {
        enabled: true,
        allowed_origins: ["http://localhost:3000", "http://localhost:4457"],
        allowed_headers: ["Authorization", "Content-Type", "X-Session-Token"],
      },
    },
  },

  log: {
    level: "trace",
    leak_sensitive_values: true,
  },
  secrets: {
    cookie: ["PLEASE-CHANGE-ME-I-AM-VERY-INSECURE"],
    cipher: ["secret-thirty-two-character-long"],
  },
  selfservice: {
    default_browser_return_url: "http://localhost:4455/",
    allowed_return_urls: [
      "http://localhost:4455",
      "http://localhost:4457",
      "https://www.ory.sh/",
      "https://example.org/",
      "https://www.example.org/",
      "exp://example.com/my-app",
      "https://example.com/my-app",
    ],
    methods: {
      link: {
        config: {
          lifespan: "1h",
        },
      },
      code: {
        config: {
          lifespan: "1h",
        },
      },
      oidc: {
        enabled: true,
        config: {
          providers: [
            {
              id: "hydra",
              label: "Ory",
              provider: "generic",
              client_id: process.env["OIDC_HYDRA_CLIENT_ID"],
              client_secret: process.env["OIDC_HYDRA_CLIENT_SECRET"],
              issuer_url: "http://localhost:4444/",
              scope: ["offline"],
              mapper_url: "file://test/e2e/profiles/oidc/hydra.jsonnet",
            },
          ],
        },
      },
    },

    flows: {
      settings: {
        privileged_session_max_age: "5m",
        ui_url: "http://localhost:4455/settings",
      },
      logout: {
        after: {
          default_browser_return_url: "http://localhost:4455/login",
        },
      },
      registration: {
        ui_url: "http://localhost:4455/registration",
        after: {
          password: {
            hooks: [
              {
                hook: "session",
              },
            ],
          },
          oidc: {
            hooks: [
              {
                hook: "session",
              },
            ],
          },
        },
      },
      login: {
        ui_url: "http://localhost:4455/login",
      },
      error: {
        ui_url: "http://localhost:4455/error",
      },
      verification: {
        ui_url: "http://localhost:4455/verify",
      },
      recovery: {
        ui_url: "http://localhost:4455/recovery",
      },
    },
  },

  courier: {
    smtp: {
      connection_uri: "smtps://test:test@localhost:1025/?skip_ssl_verify=true",
    },
  },
}
