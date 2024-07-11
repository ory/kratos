// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { expect } from "@playwright/test"
import { test } from "../../fixtures"
import { search } from "../../actions/mail"
import { extractCode } from "../../lib/helper"

const schemaConfig = {
  default_schema_id: "email",
  schemas: [
    {
      id: "email",
      url: "file://test/e2e/profiles/email/identity.traits.schema.json",
    },
  ],
}

test.describe("Recovery", () => {
  test.use({
    configOverride: {
      identity: {
        ...schemaConfig,
      },
      feature_flags: {
        use_continue_with_transitions: true,
      },
    },
  })

  test("succeeds with a valid email address", async ({ page, identity }) => {
    await page.goto("/Recovery")

    await page.getByTestId("email").fill(identity.email)
    await page.getByTestId("submit-form").click()
    await expect(page.getByTestId("ui/message/1060003")).toBeVisible()

    const mails = await search({ query: identity.email, kind: "to" })
    expect(mails).toHaveLength(1)

    const code = extractCode(mails[0])
    const wrongCode = "0" + code

    await test.step("enter wrong code", async () => {
      await page.getByTestId("code").fill(wrongCode)
      await page.getByText("Continue").click()
      await expect(page.getByTestId("ui/message/4060006")).toBeVisible()
    })

    await test.step("enter correct code", async () => {
      await page.getByTestId("code").fill(code)
      await page.getByText("Continue").click()
      await page.waitForURL(/Settings/)
      await expect(page.getByTestId("ui/message/1060001").first()).toBeVisible()
    })
  })

  test("wrong email address does not get sent", async ({ page, identity }) => {
    await page.goto("/Recovery")

    const wrongEmailAddress = "wrong-" + identity.email
    await page.getByTestId("email").fill(wrongEmailAddress)
    await page.getByTestId("submit-form").click()
    await expect(page.getByTestId("ui/message/1060003")).toBeVisible()

    try {
      await search({ query: identity.email, kind: "to" })
      expect(false).toBeTruthy()
    } catch (e) {
      // this is expected
    }
  })

  test("fails with an invalid code", async ({ page, identity }) => {
    await page.goto("/Recovery")

    await page.getByTestId("email").fill(identity.email)
    await page.getByTestId("submit-form").click()
    await page.getByTestId("ui/message/1060003").isVisible()

    const mails = await search({ query: identity.email, kind: "to" })
    expect(mails).toHaveLength(1)

    const code = extractCode(mails[0])
    const wrongCode = "0" + code

    await test.step("enter wrong repeatedly", async () => {
      for (let i = 0; i < 10; i++) {
        await page.getByTestId("code").fill(wrongCode)
        await page.getByText("Continue", { exact: true }).click()
        await expect(page.getByTestId("ui/message/4060006")).toBeVisible()
      }
    })

    await test.step("enter correct code fails", async () => {
      await page.getByTestId("code").fill(code)
      await page.getByText("Continue", { exact: true }).click()
      await expect(page.getByTestId("ui/message/4060006")).toBeVisible()
    })
  })

  test.describe("with short code expiration", () => {
    test.use({
      configOverride: {
        identity: {
          ...schemaConfig,
        },
        selfservice: {
          methods: {
            code: {
              config: {
                lifespan: "1ms",
              },
            },
          },
        },
        feature_flags: {
          use_continue_with_transitions: true,
        },
      },
    })

    test("fails with an expired code", async ({ page, identity }) => {
      await page.goto("/Recovery")

      await page.getByTestId("email").fill(identity.email)
      await page.getByTestId("submit-form").click()
      await page.getByTestId("ui/message/1060003").isVisible()

      const mails = await search({ query: identity.email, kind: "to" })
      expect(mails).toHaveLength(1)

      const code = extractCode(mails[0])

      await page.getByTestId("code").fill(code)
      await page.getByText("Continue", { exact: true }).click()
      await expect(page.getByTestId("email")).toBeVisible()
    })
  })
})
