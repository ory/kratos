import {APP_URL, gen, website} from '../../../../helpers'
import {routes as react} from "../../../../helpers/react";
import {routes as express} from "../../../../helpers/express";

context('Social Sign In Settings Errors', () => {
  [
    {
      registration: react.registration,
      settings: react.settings,
      app: 'react', profile: 'spa'
    },
    {
      registration: express.registration,
      settings: express.settings,
      app: 'express', profile: 'oidc'
    }
  ].forEach(({registration, profile, app, settings}) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
      })
      let email

      beforeEach(() => {
        cy.clearAllCookies()
        email = gen.email()

        cy.registerOidc({email, expectSession: true, website, route: registration})
        cy.visit(settings)
      })

      describe('oidc', () => {
        it('should fail to link google because id token is missing', () => {
          cy.get('button[value="google"]').click()
          cy.get('#remember').click()
          cy.get('#accept').click()

          cy.get('[data-testid="ui/message/4000001"]').should(
            'contain.text',
            'Authentication failed because no id_token was returned. Please accept the "openid" permission and try again.'
          )
        })
      })
    })
  })
})
