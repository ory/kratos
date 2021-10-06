import {gen, website} from '../../../helpers'
import {routes as express} from "../../../helpers/express";
import {routes as react} from "../../../helpers/react";

context('2FA UI settings tests', () => {
  [
    {
      settings: react.settings,
      base: react.base,
      app: 'react', profile: 'spa'
    },
    {
      settings: express.settings,
      base: express.base,
      app: 'express', profile: 'mfa'
    }
  ].forEach(({settings, profile, base, app}) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
      })

      const email = gen.email()
      const password = gen.password()

      before(() => {
        cy.registerApi({email, password, fields: {'traits.website': website}})
      })

      beforeEach(() => {
        cy.clearCookies()
        cy.login({email, password, cookieUrl: base})
        cy.visit(settings)
      })

      it('shows all settings forms', () => {
        cy.get('h3').should('contain.text', 'Profile Settings')
        cy.get('h3').should('contain.text', 'Change Password')
        cy.get('h3').should('contain.text', 'Manage 2FA Backup Recovery Codes')
        cy.get('h3').should('contain.text', 'Manage 2FA TOTP Authenticator App')
        cy.get('h3').should('contain.text', 'Manage Hardware Tokens')
        cy.get('input[name="traits.email"]').should('contain.value', email)
        cy.get('input[name="traits.website"]').should('contain.value', website)

        cy.get('[data-testid="node/text/totp_secret_key/label"]').should(
          'contain.text',
          'This is your authenticator app secret'
        )
        cy.get('button').should(
          'contain.text',
          'Generate new backup recovery codes'
        )
        cy.get('button').should('contain.text', 'Add security key')
      })
    })
  })
})
