// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { expect } from "@playwright/test"
import { test } from "../../../fixtures"
import { toConfig } from "../../../lib/helper"
import { RegistrationPage } from "../../../models/elements/registration"
import {
  OryKratosConfiguration,
  RegistrationFlowStyle,
  RegistrationNodeGroup,
} from "../../../../shared/config"

const selfservice: Partial<OryKratosConfiguration["selfservice"]> = {
  methods: {
    code: {
      passwordless_enabled: true,
    },
    password: {
      enabled: true,
    },
    webauthn: {
      enabled: true,
      config: {
        passwordless: true,
        rp: { id: "localhost", display_name: "Ory Kratos" },
      },
    },
    passkey: {
      enabled: true,
      config: {
        rp: { id: "localhost", display_name: "Ory Kratos" },
      },
    },
    totp: {
      enabled: true,
    },
    lookup_secret: {
      enabled: true,
    },
    oidc: {
      enabled: true,
      config: {
        providers: [
          {
            id: "github",
            provider: "github",
            label: "GitHub",
            client_id: "1",
            client_secret: "1",
            mapper_url: "base64://",
          },
          {
            id: "google",
            provider: "google",
            label: "Google",
            client_id: "1",
            client_secret: "1",
            mapper_url: "base64://e30=",
          },
        ],
      },
    },
  },
}

test.describe("profile_first strategy with all methods enabled", () => {
  ;["default", "password"].forEach((group: RegistrationNodeGroup) => {
    test.describe(`password group behavior is ${group}`, () => {
      ;["profile_first", "unified"].forEach((style: RegistrationFlowStyle) => {
        test.describe(`registration with ${style} enabled`, () => {
          ;[
            ["password"],
            ["password", "webauthn"],
            ["password", "code"],
            ["password", "code", "webauthn"],
            ["password", "code", "passkey"],
            ["password", "code", "passkey", "webauthn"],
          ].forEach((methods) => {
            test.describe(`methods ${methods.join(", ")} enabled`, () => {
              test.use({
                configOverride: {
                  ...toConfig({
                    style: "identifier_first",
                    mitigateEnumeration: false,
                    selfservice: {
                      ...selfservice,
                      methods: {
                        password: {
                          enabled: methods.includes("password"),
                        },
                        webauthn: {
                          enabled: methods.includes("webauthn"),
                          config: {
                            passwordless: true,
                            rp: { id: "localhost", display_name: "Ory Kratos" },
                          },
                        },
                        passkey: {
                          enabled: methods.includes("passkey"),
                          config: {
                            rp: { id: "localhost", display_name: "Ory Kratos" },
                          },
                        },
                        code: {
                          enabled: methods.includes("code"),
                          passwordless_enabled: methods.includes("code"),
                        },
                        totp: {
                          enabled: false,
                        },
                        lookup_secret: {
                          enabled: false,
                        },
                        oidc: {
                          enabled: false,
                        },
                      },
                      flows: {
                        registration: { style },
                      },
                    },
                  }),
                  feature_flags: {
                    password_profile_registration_node_group: group,
                  },
                },
              })
              test("registration does not have any duplicated fields when using profile first", async ({
                page,
                config,
              }) => {
                const registration = new RegistrationPage(page, config)
                await registration.open()

                await expect(
                  page.locator('[name="traits.email"]'),
                  "expect the profile form fields to not be duplicated",
                ).toHaveCount(style === "profile_first" ? 1 : methods.length)
              })
            })
          })
        })
      })
    })
  })
})
