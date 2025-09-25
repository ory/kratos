// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { expect, Page } from "@playwright/test"
import { createInputLocator, InputLocator } from "../../selectors/input"
import { OryKratosConfiguration } from "../../../shared/config"

export class RegistrationPage {
  public identifier: InputLocator

  constructor(
    readonly page: Page,
    readonly config: OryKratosConfiguration,
  ) {
    this.identifier = createInputLocator(page, "identifier")
  }

  async open() {
    await Promise.all([
      this.page.goto(this.config.selfservice.flows.registration.ui_url),
      this.isReady(),
      this.page.waitForURL((url) =>
        url
          .toString()
          .includes(this.config.selfservice.flows.registration.ui_url),
      ),
    ])
    await this.isReady()
  }

  inputField(name: string) {
    return this.page.locator(`input[name="${name}"]`)
  }

  submitField(name: string) {
    return this.page.locator(`[type="submit"][name="method"][value="${name}"]`)
  }

  async isReady() {
    await expect(this.inputField("csrf_token").nth(0)).toBeHidden()
  }

  async triggerRegistrationWithCode(identifier: string) {
    await this.inputField("traits.phone").fill(identifier)
    await this.submitField("profile").click()
    await this.submitField("code").click()
  }
}
