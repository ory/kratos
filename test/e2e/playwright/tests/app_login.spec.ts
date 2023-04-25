// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { test, expect, Page } from "@playwright/test"

test.describe.configure({ mode: "parallel" })

async function performOidcLogin(popup: Page, username: string) {
  await popup.waitForLoadState()

  await popup.locator("#username").fill(username)
  await popup.getByRole("button", { name: "login" }).click()

  await popup.locator("#offline").click()
  await popup.locator("#openid").click()
  await popup.locator("#website").fill("https://example.com")
  await popup.getByRole("button", { name: "login" }).click()
}

async function rejectOidcLogin(popup: Page) {
  await popup.waitForLoadState()
  await popup.getByRole("button", { name: "reject" }).click()
  await popup.close()
}

async function testRegistrationOrLogin(page: Page, username: string) {
  const popupPromise = page.waitForEvent("popup")
  await page.getByText(/sign (up|in) with ory/i).click()
  const popup = await popupPromise

  await performOidcLogin(popup, username)

  await page.waitForURL("Home")
  expect(popup.isClosed()).toBeTruthy()
  await expect(page.getByText("Welcome back")).toBeVisible()
}

async function logout(page: Page) {
  return page.getByTestId("logout").click()
}

async function testRegistration(page: Page, username: string) {
  await page.goto("Registration")
  return testRegistrationOrLogin(page, username)
}

async function testLogin(page: Page, username: string) {
  await page.goto("Login")
  return testRegistrationOrLogin(page, username)
}

test.describe("Registration", () => {
  test("register twice", async ({ page }) => {
    await testRegistration(page, "registration@example.com")
    await logout(page)
    await testRegistration(page, "registration@example.com")
  })

  test("register, then login", async ({ page }) => {
    await testRegistration(page, "registration-login@example.com")
    await logout(page)
    await testLogin(page, "registration-login@example.com")
  })

  test("register, cancel, register", async ({ page }) => {
    await page.goto("Registration")

    let popupPromise = page.waitForEvent("popup")
    await page.getByText(/sign (up|in) with ory/i).click()
    let popup = await popupPromise

    await rejectOidcLogin(popup)

    await expect(page.getByText("login rejected request")).toBeVisible()
    popupPromise = page.waitForEvent("popup")
    await page.getByText(/continue/i).click()
    popup = await popupPromise
    await performOidcLogin(popup, "register-reject-then-accept@example.com")

    await page.waitForURL("Home")
    expect(popup.isClosed()).toBeTruthy()
    await expect(page.getByText("Welcome back")).toBeVisible()
  })
})

test.describe("Login", () => {
  test("login twice", async ({ page }) => {
    await testLogin(page, "login@example.com")
    await logout(page)
    await testLogin(page, "login@example.com")
  })

  test("login, then register", async ({ page }) => {
    await testLogin(page, "login-registration@example.com")
    await logout(page)
    await testRegistration(page, "login-registration@example.com")
  })

  test("login, cancel, login", async ({ page }) => {
    await page.goto("Login")

    let popupPromise = page.waitForEvent("popup")
    await page.getByText(/sign (up|in) with ory/i).click()
    let popup = await popupPromise

    await rejectOidcLogin(popup)

    await expect(page.getByText("login rejected request")).toBeVisible()
    popupPromise = page.waitForEvent("popup")
    await page.getByText(/sign in with ory/i).click()
    popup = await popupPromise
    await performOidcLogin(popup, "login-reject-then-accept@example.com")

    await page.waitForURL("Home")
    expect(popup.isClosed()).toBeTruthy()
    await expect(page.getByText("Welcome back")).toBeVisible()
  })
})
