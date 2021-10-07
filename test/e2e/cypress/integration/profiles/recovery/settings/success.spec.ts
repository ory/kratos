import {assertRecoveryAddress, gen} from '../../../../helpers'
import {routes as react} from "../../../../helpers/react";
import {routes as express} from "../../../../helpers/express";

context('Account Recovery Success', () => {
  [
    {
      settings: react.settings,
      base: react.base,
      app: 'react', profile: 'spa'
    },
    {
      settings: express.settings,
      base: express.base,
      app: 'express', profile: 'recovery'
    }
  ].forEach(({settings, profile, base, app}) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
      })

      let identity

      beforeEach(() => {
        cy.deleteMail()
        cy.longRecoveryLifespan()
        cy.longLinkLifespan()
        cy.disableVerification()
        cy.enableRecovery()

        identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.login({...identity, cookieUrl: base})
      })

      it('should update the recovery address when updating the email', () => {
        cy.visit(settings)
        const email = gen.email()
        cy.get('input[name="traits.email"]').clear().type(email)
        cy.get('button[value="profile"]').click()
        cy.expectSettingsSaved()
        cy.get('input[name="traits.email"]').should('contain.value', email)

        cy.getSession().should(assertRecoveryAddress({email}))
      })

      xit('should not show an immediate error when a recovery address already exists', () => {
        // account enumeration prevention, needs to be implemented.
      })
    })
  })
})
