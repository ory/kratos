// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { expect, Page } from "@playwright/test"
import { TOTP } from "otpauth"
import { OryKratosConfiguration } from "../../../shared/config"

export class SettingsPage {
  constructor(
    readonly page: Page,
    readonly config: OryKratosConfiguration,
  ) {}

  async isReady() {
    for (const csrfInput of await this.page
      .locator(`input[name="csrf_token"]`)
      .all()) {
      await expect(csrfInput).toHaveValue(/.+/)
    }
  }

  async open() {
    await this.page.goto(this.config.selfservice.flows.settings.ui_url)
    await this.isReady()
  }

  async setupTotp() {
    const totpSecret = await this.page
      .getByTestId("node/text/totp_secret_key/text")
      .locator("code")
      .textContent()
    await expect(totpSecret).toMatch(/^[A-Z2-7]{32}$/)
    const totpCode = () => {
      return new TOTP({ secret: totpSecret }).generate()
    }
    await this.page.fill("input[name=totp_code]", totpCode())
    await this.page.locator("button[name=method][value=totp]").click()

    return { totpCode }
  }
}
