import {routes as express} from "../../../../helpers/express";
import {routes as react} from "../../../../helpers/react";

context('UI tests using the email profile', () => {
  [{
    route: express.login,
    app: 'express', profile: 'email'
  }, {
    route: react.login,
    app: 'react', profile: 'spa'
  }].forEach(({route, profile, app}) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
      })

      beforeEach(() => {
        cy.visit(route)
      })

      it('should use the json schema titles', () => {
        cy.get('input[name="password_identifier"]')
          .parent()
          .should('contain.text', 'ID')
        cy.get('input[name="password"]')
          .parent()
          .should('contain.text', 'Password')
        cy.get('button[value="password"]').should('contain.text', 'Sign in')
      })

      it('clicks the log in link', () => {
        cy.get('a[href*="registration"]').click()
        cy.location('pathname').should('include', 'registration')

        if (app === 'express') {
          cy.location('search').should('not.be.empty')
        }
      })
    })
  })
})
