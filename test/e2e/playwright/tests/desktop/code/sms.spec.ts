// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { test } from "../../../fixtures"
import { toConfig } from "../../../lib/helper"
import smsSchema from "../../../fixtures/schemas/sms"
import { LoginPage } from "../../../models/elements/login"
import { hasSession } from "../../../actions/session"
import { createIdentityWithPhoneNumber } from "../../../actions/identity"
import {
  deleteDocument,
  documentUrl,
  fetchDocument,
} from "../../../actions/webhook"
import { RegistrationPage } from "../../../models/elements/registration"
import { CountryNames, generatePhoneNumber } from "phone-number-generator-js"

const documentId = "doc-" + Math.random().toString(36).substring(7)

test.describe("account enumeration protection off", () => {
  test.use({
    configOverride: {
      security: {
        account_enumeration: {
          enabled: false,
        },
      },
      selfservice: {
        flows: {
          login: {
            style: "unified",
          },
          registration: {
            after: {
              code: {
                hooks: [
                  {
                    hook: "session",
                  },
                ],
              },
            },
          },
        },
        methods: {
          code: {
            passwordless_enabled: true,
          },
          password: {
            enabled: false,
          },
        },
      },
      courier: {
        channels: [
          {
            id: "sms",
            type: "http",
            request_config: {
              body: "base64://ZnVuY3Rpb24oY3R4KSB7DQpjdHg6IGN0eCwNCn0=",
              method: "PUT",
              url: documentUrl(documentId),
            },
          },
        ],
      },
      identity: {
        default_schema_id: "sms",
        schemas: [
          {
            id: "sms",
            url:
              "base64://" +
              Buffer.from(JSON.stringify(smsSchema), "ascii").toString(
                "base64",
              ),
          },
        ],
      },
    },
  })

  test.afterEach(async () => {
    await deleteDocument(documentId)
  })

  test("login succeeds", async ({ page, config, kratosPublicURL }) => {
    const identity = await createIdentityWithPhoneNumber(page.request)

    const login = new LoginPage(page, config)
    await login.open()
    await login.triggerLoginWithCode(identity.phone)

    const result = await fetchDocument(documentId)
    await login.codeInput.input.fill(result.ctx.template_data.login_code)
    await login.codeSubmit.getByText("Continue").click()
    await hasSession(page.request, kratosPublicURL)
  })

  test("registration succeeds", async ({ page, config, kratosPublicURL }) => {
    const phone = generatePhoneNumber({
      countryName: CountryNames.Germany,
      withoutCountryCode: false,
    })

    const registration = new RegistrationPage(page, config)
    await registration.open()
    await registration.triggerRegistrationWithCode(phone)

    const result = await fetchDocument(documentId)
    const code = result.ctx.template_data.registration_code
    await registration.inputField("code").fill(code)
    await registration.submitField("code").getByText("Continue").click()
    await hasSession(page.request, kratosPublicURL)
  })
})
