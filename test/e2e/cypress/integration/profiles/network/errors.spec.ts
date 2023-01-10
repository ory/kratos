// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { routes as express } from "../../../helpers/express"
import { gen } from "../../../helpers"

describe("Registration failures with email profile", () => {
  before(() => {
    cy.useConfigProfile("network")
    cy.proxy("express")
  })

  it("should not be able to register if we need a localhost schema", () => {
    cy.setDefaultIdentitySchema("localhost")
    cy.visit(express.registration, { failOnStatusCode: false })
    cy.get('[data-testid="code-box"]').should(
      "contain.text",
      "is not a public IP address", // could be ::1 or 127.0.0.1
    )
  })

  it("should not be able to register if we schema has a local ref", () => {
    cy.setDefaultIdentitySchema("ref")
    cy.visit(express.registration, { failOnStatusCode: false })
    cy.get('[data-testid="code-box"]').should(
      "contain.text",
      "192.168.178.1 is not a public IP address",
    )
  })

  it("should not be able to login because pre webhook uses local url", () => {
    cy.setDefaultIdentitySchema("working")
    cy.visit(express.login, { failOnStatusCode: false })
    cy.get('[data-testid="code-box"]').should(
      "contain.text",
      "192.168.178.2 is not a public IP address",
    )
  })

  it("should not be able to verify because post webhook uses local jsonnet", () => {
    cy.setDefaultIdentitySchema("working")
    cy.visit(express.registration, { failOnStatusCode: false })
    cy.get('input[name="traits.email"]').type(gen.email())
    cy.get('input[name="traits.website"]').type("https://google.com/")
    cy.get('input[name="password"]').type(gen.password())
    cy.get('[type="submit"]').click()
    cy.get('[data-testid="code-box"]').should(
      "contain.text",
      "192.168.178.3 is not a public IP address",
    )
  })
})
