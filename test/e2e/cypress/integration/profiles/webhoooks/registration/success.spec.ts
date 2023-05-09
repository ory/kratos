// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, APP_URL, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { testRegistrationWebhook } from "../../../../helpers/webhook"

context("Registration success with email profile with webhooks", () => {
  ;[
    {
      route: express.registration,
      app: "express" as "express",
      profile: "webhooks",
    },
  ].forEach(({ route, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.clearAllCookies()
        cy.visit(route)
      })

      it("should sign up and be logged in", () => {
        const email = gen.email()
        const password = gen.password()

        cy.get(appPrefix(app) + 'input[name="traits"]').should("not.exist")
        cy.get('input[name="traits.email"]').type(email)
        cy.get('input[name="password"]').type(password)

        cy.submitPasswordForm()
        if (app === "express") {
          cy.get("a[href*='sessions']").click()
        }
        cy.get("pre").should("contain.text", email)

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.schema_id).to.equal("default")
          expect(identity.schema_url).to.equal(`${APP_URL}/schemas/ZGVmYXVsdA`)
          expect(identity.traits.email).to.equal(email)
        })
      })

      it("should work without transient_payload", () => {
        const WEBHOOK_TARGET = "https://webhook-target-gsmwn5ab4a-uc.a.run.app"
        const documentUrl = (key: string) =>
          `${WEBHOOK_TARGET}/documents/${key}`
        const email = gen.email()
        const password = gen.password()
        const documentID = gen.password()
        const jsonnet = "function(ctx) ctx"
        cy.setupHooks("registration", "after", "password", [
          {
            hook: "web_hook",
            config: {
              body: `base64://${Buffer.from(jsonnet).toString("base64")}`,
              url: documentUrl(documentID),
              method: "PUT",
            },
          },
        ])

        cy.get('input[name="traits.email"]').type(email)
        cy.get('input[name="password"]').type(password)

        cy.submitPasswordForm()

        cy.request({
          url: documentUrl(documentID),
          method: "GET",
          headers: {
            Accept: "application/json",
          },
        }).then(({ body, status }) => {
          const b = JSON.parse(body)
          expect(status).to.equal(200)
          expect(b.identity).is.not.undefined
          expect(b.flow.transient_payload).is.empty
        })
      })

      it("should pass transient_payload to webhook", () => {
        testRegistrationWebhook(
          cy.setPostPasswordRegistrationHooks.bind(cy),
          () => {
            const email = gen.email()
            const password = gen.password()

            cy.get('input[name="traits.email"]').type(email)
            cy.get('input[name="password"]').type(password)

            cy.submitPasswordForm()
          },
        )
      })

      it("should sign up and modify the identity", () => {
        const email = gen.email()
        const password = gen.password()

        const updatedEmail = {
          identity: {
            traits: { email: "updated-" + email },
            verifiable_addresses: [
              { via: "email", value: "updated-" + email, verified: true },
              { via: "email", value: "this-email-should-be-ignored" },
              { via: "email", value: "" },
            ],
            recovery_addresses: [
              { via: "email", value: "updated-" + email },
              { via: "email", value: "this-email-should-be-ignored" },
              { via: "email", value: "" },
            ],
            metadata_public: { some: "public fields" },
          },
        }
        cy.setPostPasswordRegistrationHooks([
          { hook: "session" },
          {
            hook: "web_hook",
            config: {
              url:
                "http://127.0.0.1:4459/webhook/write?response=" +
                encodeURIComponent(JSON.stringify(updatedEmail)),
              method: "POST",
              body: "file://test/e2e/profiles/webhooks/webhook_body.jsonnet",
              response: { parse: true },
            },
          },
        ])

        cy.get(appPrefix(app) + 'input[name="traits"]').should("not.exist")
        cy.get('input[name="traits.email"]').type(email)
        cy.get('input[name="password"]').type(password)

        cy.submitPasswordForm()

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.schema_id).to.equal("default")
          expect(identity.schema_url).to.equal(`${APP_URL}/schemas/ZGVmYXVsdA`)
          expect(identity.traits.email).to.equal("updated-" + email)
          expect((identity as any).metadata_public.some).to.equal(
            "public fields",
          )
          expect(identity.verifiable_addresses[0].verified).to.equal(true)
          expect(identity.verifiable_addresses[0].verified_at).not.to.be.empty
          expect(identity.verifiable_addresses[0].via).to.eq("email")
          expect(identity.verifiable_addresses[0].value).to.eq(
            "updated-" + email,
          )
          expect(identity.verifiable_addresses).to.have.length(1)

          expect(identity.recovery_addresses).to.have.length(1)
          expect(identity.recovery_addresses[0].via).to.eq("email")
          expect(identity.recovery_addresses[0].value).to.eq("updated-" + email)
        })
      })
    })
  })
})
