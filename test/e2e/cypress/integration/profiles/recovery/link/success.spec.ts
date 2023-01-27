// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, assertRecoveryAddress, gen } from "../../../../helpers"
import { routes as react } from "../../../../helpers/react"
import { routes as express } from "../../../../helpers/express"

context("Account Recovery Success", () => {
  ;[
    {
      recovery: react.recovery,
      base: react.base,
      app: "react" as "react",
      profile: "spa",
    },
    {
      recovery: express.recovery,
      base: express.base,
      app: "express" as "express",
      profile: "recovery",
    },
  ].forEach(({ recovery, profile, base, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      let identity

      beforeEach(() => {
        cy.deleteMail()
        cy.longRecoveryLifespan()
        cy.longLinkLifespan()
        cy.disableVerification()
        cy.enableRecovery()
        cy.useRecoveryStrategy("link")

        identity = gen.identityWithWebsite()
        cy.registerApi(identity)
      })

      it("should contain the recovery address in the session", () => {
        cy.visit(recovery)
        cy.login({ ...identity, cookieUrl: base })
        cy.getSession().should(assertRecoveryAddress(identity))
      })

      it("should perform a recovery flow", () => {
        cy.recoverApi({ email: identity.email })

        cy.recoverEmail({ expect: identity })

        cy.getSession()
        cy.location("pathname").should("eq", "/settings")

        const newPassword = gen.password()
        cy.get(appPrefix(app) + 'input[name="password"]')
          .clear()
          .type(newPassword)
        cy.get('button[value="password"]').click()
        cy.expectSettingsSaved()
        cy.get('input[name="password"]').should("be.empty")

        cy.logout()
        cy.login({
          email: identity.email,
          password: newPassword,
          cookieUrl: base,
        })
      })
    })
  })

  it("should recover, set password and be redirected", () => {
    const app = "express" as "express"

    cy.deleteMail()
    cy.useConfigProfile("recovery")
    cy.proxy(app)

    cy.deleteMail()
    cy.longRecoveryLifespan()
    cy.longLinkLifespan()
    cy.disableVerification()
    cy.enableRecovery()

    const identity = gen.identityWithWebsite()
    cy.registerApi(identity)

    cy.recoverApi({ email: identity.email, returnTo: "https://www.ory.sh/" })

    cy.recoverEmail({ expect: identity })

    cy.getSession()
    cy.location("pathname").should("eq", "/settings")

    cy.get(appPrefix(app) + 'input[name="password"]')
      .clear()
      .type(gen.password())
    cy.get('button[value="password"]').click()
    cy.url().should("eq", "https://www.ory.sh/")
  })

  it("should recover even if already logged into another account", () => {
    const app = "express" as "express"

    cy.deleteMail()
    cy.useConfigProfile("recovery")
    cy.proxy(app)

    cy.deleteMail()
    cy.disableVerification()

    const identity1 = gen.identityWithWebsite()
    cy.registerApi(identity1)
    const identity2 = gen.identityWithWebsite()
    cy.registerApi(identity2)

    cy.recoverApi({ email: identity2.email })

    // first log in as identity1

    cy.visit(express.login)

    cy.get(appPrefix(app) + 'input[name="identifier"]').type(identity1.email)
    cy.get('input[name="password"]').type(identity1.password)
    cy.get('button[value="password"]').click()

    cy.location("pathname").should("not.contain", "/login")

    // then recover identity2, while still logged in as identity1

    cy.recoverEmail({ expect: identity2 })

    cy.getSession()
    cy.location("pathname").should("eq", "/settings")
    cy.get('input[name="traits.email"]').should("have.value", identity2.email)
  })
})
