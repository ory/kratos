import {gen} from '../../../../helpers'
import {routes as express} from "../../../../helpers/express";
import {routes as react} from "../../../../helpers/react";

describe('Basic email profile with failing login flows', () => {
  [
    {
      route: express.login,
      app: 'express', profile: 'email'
    },
    {
      route: react.login,
      app: 'react', profile: 'spa'
    }
  ].forEach(({route, profile, app}) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
      });

      beforeEach(() => {
        cy.clearCookies()
        cy.visit(route)
      })

      it('fails when CSRF cookies are missing', () => {
        cy.get('input[name="password_identifier"]').type('i-do-not-exist')
        cy.get('input[name="password"]').type('invalid-password')

        cy.shouldHaveCsrfError({app})
      })

      it('fails when a disallowed return_to url is requested', () => {
        cy.shouldErrorOnDisallowedReturnTo(route + '?return_to=https://not-allowed', {app})
      })

      describe('shows validation errors when invalid signup data is used', () => {
        beforeEach(() => {
          cy.clearCookies()
          cy.visit(route)
        })

        it('should show an error when the identifier is missing', () => {
          cy.submitPasswordForm()
          cy.get('*[data-testid="ui/message/4000002"]').should(
            'contain.text',
            'Property password_identifier is missing'
          )
          cy.get('*[data-testid="ui/message/4000002"]').should(
            'contain.text',
            'Property password is missing'
          )
        })

        it('should show an error when the password is missing', () => {
          const identity = gen.email()
          cy.get('input[name="password_identifier"]')
            .type(identity)
            .should('have.value', identity)

          cy.submitPasswordForm()
          cy.get('*[data-testid^="ui/message/"]').invoke('text').then((text) => {
            expect(text).to.be.oneOf(['length must be >= 1, but got 0', 'Property password is missing.'])
          })
        })

        it('should show fail to sign in', () => {
          cy.get('input[name="password_identifier"]').type('i-do-not-exist')
          cy.get('input[name="password"]').type('invalid-password')

          cy.submitPasswordForm()
          cy.get('*[data-testid="ui/message/4000006"]').should(
            'contain.text',
            'credentials are invalid'
          )
        })
      })
    })
  })
})
