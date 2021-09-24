import {gen} from '../../../../helpers'
import {routes as express} from "../../../../helpers/express";
import {routes as react} from "../../../../helpers/react";

describe('Basic email profile with failing login flows', () => {
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
      });

      beforeEach(() => {
        cy.visit(route)
      })

      it('fails when CSRF cookies are missing', () => {
        cy.clearCookies()

        cy.get('input[name="password_identifier"]').type('i-do-not-exist')
        cy.get('input[name="password"]').type('invalid-password')

        let initial
        cy.location().should((location) => {
          initial = location.search
        })
        cy.get('button[type="submit"]').click()

        // We end up at a new flow
        cy.location('search').should('not.eq', initial)
      })

      describe('shows validation errors when invalid signup data is used', () => {
        it('should show an error when the identifier is missing', () => {
          cy.get('button[type="submit"]').click()
          cy.get('*[data-testid="ui.node.message.4000001"]').should(
            'contain.text',
            'length must be >= 1, but got 0'
          )
        })

        it('should show an error when the password is missing', () => {
          const identity = gen.email()
          cy.get('input[name="password_identifier"]')
            .type(identity)
            .should('have.value', identity)

          cy.get('button[type="submit"]').click()

          cy.get('*[data-testid^="ui.node.message."]').invoke('text').then((text)=> {
            expect(text).to.be.oneOf(['length must be >= 1, but got 0', 'Property password is missing.'])
          })
        })

        it('should show fail to sign in', () => {
          cy.get('input[name="password_identifier"]').type('i-do-not-exist')
          cy.get('input[name="password"]').type('invalid-password')

          cy.get('button[type="submit"]').click()
          cy.get('*[data-testid="ui.node.message.4000006"]').should(
            'contain.text',
            'credentials are invalid'
          )
        })
      })
    })
  })
})
