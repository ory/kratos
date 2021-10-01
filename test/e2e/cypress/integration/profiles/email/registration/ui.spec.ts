import {routes as express} from "../../../../helpers/express";
import {routes as react} from "../../../../helpers/react";

describe('Registration UI for email flows', () => {
  [{
    route: express.registration,
    app: 'express', profile: 'email'
  }, {
    route: react.registration,
    app: 'react', profile: 'spa'
  }].forEach(({route, profile, app}) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
      })

      beforeEach(() => {
        cy.visit(route)
      })

      describe('use ui elements', () => {
        it('should use the json schema titles', () => {
          cy.get('input[name="traits.email"]')
            .parent()
            .should('contain.text', 'Your E-Mail')
          cy.get('input[name="traits.website"]')
            .parent()
            .should('contain.text', 'Your website')
          cy.get('button[value="password"]').should('contain.text', 'Sign up')
        })

        it('clicks the log in link', () => {
          cy.get('*[data-testid="cta-link"]').click()
          cy.location('pathname').should('include', '/login')
          if (app === 'express') {
            cy.location('search').should('not.be.empty')
          }
        })
      })
    })
  })
})
