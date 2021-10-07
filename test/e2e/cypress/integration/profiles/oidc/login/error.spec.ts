import {gen, website} from '../../../../helpers'
import {routes as react} from "../../../../helpers/react";
import {routes as express} from "../../../../helpers/express";

context('Social Sign In Errors', () => {
  [
    {
      login: react.login,
      app: 'react', profile: 'spa'
    },
    {
      login: express.login,
      app: 'express', profile: 'oidc'
    }
  ].forEach(({login, profile, app}) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
      })

      beforeEach(() => {
        cy.clearAllCookies()
        cy.visit(login)
      })

      it('should fail when the login request is rejected', () => {
       cy.triggerOidc()
        cy.get('#reject').click()
        cy.location('pathname').should('equal', '/login')
        cy.get('[data-testid="ui/message/4000001"]').should(
          'contain.text',
          'login rejected request'
        )
        cy.noSession()
      })

      it('should fail when the consent request is rejected', () => {
        const email = gen.email()
       cy.triggerOidc()
        cy.get('#username').type(email)
        cy.get('#accept').click()
        cy.get('#reject').click()
        cy.location('pathname').should('equal', '/login')
        cy.get('[data-testid="ui/message/4000001"]').should(
          'contain.text',
          'consent rejected request'
        )
        cy.noSession()
      })

      it('should fail when the id_token is missing', () => {
        const email = gen.email()
       cy.triggerOidc()
        cy.get('#username').type(email)
        cy.get('#accept').click()
        cy.get('#website').type(website)
        cy.get('#accept').click()
        cy.location('pathname').should('equal', '/login')
        cy.get('[data-testid="ui/message/4000001"]').should('contain.text', 'no id_token')
      })
    })
  })
})
