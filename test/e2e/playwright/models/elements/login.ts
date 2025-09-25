// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { expect, Locator, Page } from "@playwright/test"
import { createInputLocator, InputLocator } from "../../selectors/input"
import { URLSearchParams } from "node:url"
import { OryKratosConfiguration } from "../../../shared/config"

enum LoginStyle {
  IdentifierFirst = "identifier_first",
  Unified = "unified",
}

type SubmitOptions = {
  submitWithKeyboard?: boolean
  waitForURL?: string | RegExp
}

export class LoginPage {
  public submitPassword: Locator
  public github: Locator
  public google: Locator
  public signup: Locator

  public identifier: InputLocator
  public password: InputLocator
  public totpInput: InputLocator
  public totpSubmit: Locator
  public lookupInput: InputLocator
  public lookupSubmit: Locator
  public codeSubmit = this.page.locator('button[type="submit"][value="code"]')
  public codeInput = createInputLocator(this.page, "code")

  public alert: Locator

  constructor(
    readonly page: Page,
    readonly config: OryKratosConfiguration,
  ) {
    this.identifier = createInputLocator(page, "identifier")
    this.password = createInputLocator(page, "password")
    this.totpInput = createInputLocator(page, "totp_code")
    this.lookupInput = createInputLocator(page, "lookup_secret")

    this.submitPassword = page.locator(
      '[type="submit"][name="method"][value="password"]',
    )

    this.github = page.locator('[name="provider"][value="github"]')
    this.google = page.locator('[name="provider"][value="google"]')

    this.totpSubmit = page.locator('[name="method"][value="totp"]')
    this.lookupSubmit = page.locator('[name="method"][value="lookup_secret"]')

    this.signup = page.locator('[data-testid="signup-link"]')

    // this.submitHydra = page.locator('[name="provider"][value="hydra"]')
    // this.forgotPasswordLink = page.locator(
    //   "[data-testid='forgot-password-link']",
    // )
    // this.logoutLink = page.locator("[data-testid='logout-link']")
  }

  async submitIdentifierFirst(identifier: string) {
    await this.inputField("identifier").fill(identifier)
    await this.submit("identifier_first", {
      waitForURL: new RegExp(this.config.selfservice.flows.login.ui_url),
    })
  }

  async loginWithPassword(
    identifier: string,
    password: string,
    opts?: SubmitOptions,
  ) {
    switch (this.config.selfservice.flows.login.style) {
      case LoginStyle.IdentifierFirst:
        await this.submitIdentifierFirst(identifier)
        break
      case LoginStyle.Unified:
        await this.inputField("identifier").fill(identifier)
        break
    }

    await this.inputField("password").fill(password)
    await this.submit("password", opts)
  }

  async triggerLoginWithCode(identifier: string, opts?: SubmitOptions) {
    switch (this.config.selfservice.flows.login.style) {
      case LoginStyle.IdentifierFirst:
        await this.submitIdentifierFirst(identifier)
        break
      case LoginStyle.Unified:
        await this.inputField("identifier").fill(identifier)
        break
    }

    await this.codeSubmit.click()
  }

  async open({
    aal,
    refresh,
  }: {
    aal?: string
    refresh?: boolean
  } = {}) {
    const p = new URLSearchParams()
    if (refresh) {
      p.append("refresh", "true")
    }

    if (aal) {
      p.append("aal", aal)
    }

    await Promise.all([
      this.page.goto(
        this.config.selfservice.flows.login.ui_url + "?" + p.toString(),
      ),
      this.isReady(),
      this.page.waitForURL((url) =>
        url.toString().includes(this.config.selfservice.flows.login.ui_url),
      ),
    ])
    await this.isReady()
  }

  async isReady() {
    await expect(this.inputField("csrf_token").nth(0)).toBeHidden()
  }

  submitMethod(method: string) {
    switch (method) {
      case "google":
      case "github":
      case "hydra":
        return this.page.locator(`[name="provider"][value="${method}"]`)
    }
    return this.page.locator(`[name="method"][value="${method}"]`)
  }

  inputField(name: string) {
    return this.page.locator(`input[name=${name}]`)
  }

  async submit(method: string, opts?: SubmitOptions) {
    const waitFor = [
      opts?.waitForURL
        ? this.page.waitForURL(opts.waitForURL)
        : Promise.resolve(),
    ]

    if (opts?.submitWithKeyboard) {
      waitFor.push(this.page.keyboard.press("Enter"))
    } else {
      waitFor.push(this.submitMethod(method).click())
    }

    await Promise.all(waitFor)
  }

  //
  // async submitPasswordForm(
  //   id: string,
  //   password: string,
  //   expectURL: string | RegExp,
  //   options: {
  //     submitWithKeyboard?: boolean
  //     style?: LoginStyle
  //   } = {
  //     submitWithKeyboard: false,
  //     style: LoginStyle.OneStep,
  //   },
  // ) {
  //   await this.isReady()
  //   await this.inputField("identifier").fill(id)
  //
  //   if (options.style === LoginStyle.IdentifierFirst) {
  //     await this.submitMethod("identifier_first").click()
  //     await this.inputField("password").fill(password)
  //   } else {
  //     await this.inputField("password").fill(password)
  //   }
  //
  //   const nav = this.page.waitForURL(expectURL)
  //
  //   if (submitWithKeyboard) {
  //     await this.page.keyboard.press("Enter")
  //   } else {
  //     await this.submitPassword.click()
  //   }
  //
  //   await nav
  // }
  //
  // readonly baseURL: string
  // readonly submitHydra: Locator
  // readonly forgotPasswordLink: Locator
  // readonly logoutLink: Locator
  //
  // async goto(returnTo?: string, refresh?: boolean) {
  //   const u = new URL(routes.hosted.login(this.baseURL))
  //   if (returnTo) {
  //     u.searchParams.append("return_to", returnTo)
  //   }
  //   if (refresh) {
  //     u.searchParams.append("refresh", refresh.toString())
  //   }
  //   await this.page.goto(u.toString())
  //   await this.isReady()
  // }
  //
  // async loginWithHydra(email: string, password: string) {
  //   await this.submitHydra.click()
  //   await this.page.waitForURL(new RegExp(OIDC_PROVIDER))
  //
  //   await this.page.locator("input[name=email]").fill(email)
  //   await this.page.locator("input[name=password]").fill(password)
  //
  //   await this.page.locator("input[name=submit][id=accept]").click()
  // }
  //
  // async loginWithOIDC(email = generateEmail(), password = generatePassword()) {
  //   await this.page.fill('[name="email"]', email)
  //   await this.page.fill('[name="password"]', password)
  //   await this.page.click("#accept")
  // }
  //
  // async loginAndAcceptConsent(
  //   email = generateEmail(),
  //   password = generatePassword(),
  //   {rememberConsent = true, rememberLogin = false} = {},
  // ) {
  //   await this.page.fill('[name="email"]', email)
  //   await this.page.fill('[name="password"]', password)
  //   rememberLogin && (await this.page.check('[name="remember"]'))
  //   await this.page.click("#accept")
  //
  //   await this.page.click("#offline")
  //   await this.page.click("#openid")
  //   rememberConsent && (await this.page.check("[name=remember]"))
  //   await this.page.click("#accept")
  //
  //   return email
  // }
  //
  // async expectAlert(id: string) {
  //   await this.page.getByTestId(`ui/message/${id}`).waitFor()
  // }
}
