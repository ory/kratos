// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { gen, KRATOS_ADMIN, MOBILE_URL, website } from "../../../../helpers"

context("Mobile OIDC Settings", () => {
  before(() => {
    cy.useConfigProfile("mobile")
  })

  describe("link buttons", () => {
    const email = gen.email()
    const password = gen.password()

    before(() => {
      cy.registerApi({
        email,
        password,
        fields: { "traits.website": website },
      })
    })

    beforeEach(() => {
      cy.clearAllCookies()
      cy.loginMobile({ email, password })
      cy.visit(MOBILE_URL + "/Settings")
    })

    it("should show link buttons for unlinked providers", () => {
      cy.get('[data-testid="settings-oidc"]').should("exist")
      cy.get('[data-testid="settings-oidc"]').within(() => {
        cy.get('[data-testid="field/link/hydra"]').should("exist")
        cy.get('[data-testid="field/link/google"]').should("exist")
        cy.get('[data-testid="field/link/github"]').should("exist")
      })
    })
  })

  describe("unlink", () => {
    let email: string
    let password: string

    beforeEach(() => {
      email = gen.email()
      password = gen.password()

      // Create an identity with both password and OIDC credentials via Admin API
      cy.request({
        method: "POST",
        url: KRATOS_ADMIN + "/admin/identities",
        body: {
          schema_id: "default",
          traits: { email, website },
          credentials: {
            password: { config: { password } },
            oidc: {
              config: {
                providers: [
                  {
                    provider: "hydra",
                    subject: email,
                  },
                ],
              },
            },
          },
        },
      })

      cy.clearAllCookies()
      cy.loginMobile({ email, password })
      cy.visit(MOBILE_URL + "/Settings")
    })

    it("should show unlink button for linked provider", () => {
      cy.get('[data-testid="settings-oidc"]').should("exist")
      cy.get('[data-testid="settings-oidc"]').within(() => {
        cy.get('[data-testid="field/unlink/hydra"]').should("exist")
        cy.get('[data-testid="field/link/google"]').should("exist")
        cy.get('[data-testid="field/link/github"]').should("exist")
      })
    })

    it("should unlink a provider", () => {
      cy.get('[data-testid="settings-oidc"]').within(() => {
        cy.get('[data-testid="field/unlink/hydra"]').should("exist")
        cy.get('[data-testid="field/unlink/hydra"]').click()
      })

      cy.expectSettingsSaved()

      // After unlinking, the button should change to link
      cy.get('[data-testid="settings-oidc"]').within(() => {
        cy.get('[data-testid="field/link/hydra"]').should("exist")
        cy.get('[data-testid="field/unlink/hydra"]').should("not.exist")
      })
    })
  })
})
