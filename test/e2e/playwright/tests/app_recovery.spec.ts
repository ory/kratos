// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { expect } from "@playwright/test"
import { test } from "../fixtures"
import { search } from "../actions/mail"
import { extractCode } from "../lib/helper"

test.use({
  configOverride: {
    identity: {
      default_schema_id: "email",
      schemas: [
        {
          id: "email",
          url: "file://test/e2e/profiles/email/identity.traits.schema.json",
        },
      ],
    },
  },
})

test("recovery works", async ({ page, identity }) => {
  await page.goto("/Recovery")

  const emailInput = page.getByTestId("email")
  await emailInput.waitFor()

  await emailInput.fill(identity.traits.email)

  await page.getByTestId("submit-form").click()

  await page.getByTestId("ui/message/1060003").waitFor()

  const mails = await search(identity.traits.email, "to")
  expect(mails).toHaveLength(1)

  const code = extractCode(mails[0])

  const codeInput = page.getByTestId("code")
  await codeInput.fill(code)

  await page.getByTestId("field/method/code").getByTestId("submit-form").click()

  await page.getByTestId("ui/message/1060001").waitFor()
})

// TODO: add test for
// - recovery with a not registered email
// - recovery with a not verified email
// - recovery brute force
