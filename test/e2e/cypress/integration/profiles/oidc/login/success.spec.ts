import {gen, website} from '../../../../helpers'
import {routes as react} from "../../../../helpers/react";
import {routes as express} from "../../../../helpers/express";

context('Social Sign In Successes', () => {
  [
    {
      login: react.login,
      registration: react.registration,
      app: 'react', profile: 'spa'
    },
    {
      login: express.login,
      registration: express.registration,
      app: 'express', profile: 'oidc'
    }
  ].forEach(({login, registration, profile, app}) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
      })

      beforeEach(() => {
        cy.clearAllCookies()
      })

      it('should be able to sign up, sign out, and then sign in', () => {
        const email = gen.email()
        cy.registerOidc({email, website, route: registration})
        cy.get('[data-testid=logout]').click()
        cy.noSession()
        cy.loginOidc({url: login})
      })
    })
  })
})
