import {gen, website} from '../../../../helpers'
import {routes as react} from "../../../../helpers/react";
import {routes as express} from "../../../../helpers/express";

context('Social Sign Out Successes', () => {
  [
    {
      base: react.base,
      registration: react.registration,
      app: 'react', profile: 'spa'
    },
    {
      base: express.base,
      registration: express.registration,
      app: 'express', profile: 'oidc'
    }
  ].forEach(({base, registration,profile, app}) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
      })

      const email = gen.email()

      before(() => {
        cy.clearAllCookies()
        cy.registerOidc({email, website, route: registration})
      })

      beforeEach(() => {
        cy.visit(base)
      })

      it('should sign out and be able to sign in again', () => {
        cy.get('[data-testid=logout]').click()
        cy.noSession()
        cy.url().should('include', '/login')
      })
    })
  })
})
