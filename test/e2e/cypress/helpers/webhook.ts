// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { fail } from "assert"
import { gen } from "."

const WEBHOOK_TARGET = "https://webhook-target-gsmwn5ab4a-uc.a.run.app"
const documentUrl = (key: string) => `${WEBHOOK_TARGET}/documents/${key}`
const jsonnet = Buffer.from("function(ctx) ctx").toString("base64")

export const testRegistrationWebhook = (
  configSetup: (
    hooks: Array<{ hook: string; config?: any }>,
  ) => Cypress.Chainable<void>,
  act: () => Cypress.Chainable<void> | void,
) => {
  const documentID = gen.password()
  configSetup([
    {
      hook: "web_hook",
      config: {
        body: "base64://" + jsonnet,
        url: documentUrl(documentID),
        method: "PUT",
      },
    },
    { hook: "session" },
  ])

  const transient_payload = {
    stuff: {
      blub: [42, 3.14152],
      fu: "bar",
    },
    consent: true,
  }
  cy.intercept("POST", /.*\/self-service\/registration.*/, (req) => {
    switch (typeof req.body) {
      case "string":
        req.body =
          req.body +
          "&transient_payload=" +
          encodeURIComponent(JSON.stringify(transient_payload))
        break
      case "object":
        req.body = {
          ...req.body,
          transient_payload,
        }
        break

      default:
        fail()
        break
    }
    req.continue()
  })

  act()

  cy.request(documentUrl(documentID)).then(({ body, status }) => {
    const b = JSON.parse(body)
    expect(status).to.equal(200)
    expect(b.identity).is.not.undefined
    expect(b.flow.transient_payload).to.deep.equal(transient_payload)
  })
}
