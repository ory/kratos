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

        let initial
        cy.location().should((location) => {
          initial = location.search
        })

        cy.clearAllCookies()
        cy.submitPasswordForm()

        // We end up at a new flow
        if (app === 'express') {
          cy.location().should((location) => {
            expect(initial).to.not.be.empty
            expect(location.search).to.not.eq(initial)
          })

          cy.location('pathname').should('include', '/error')
          cy.get('code').should('contain.text', 'csrf_token')
        } else {
          cy.location('pathname').should('include', '/login')
          cy.get('.Toastify').should('contain.text', 'A security violation was detected, please fill out the form again.')
        }
      })

      it('fails when a disallowed return_to url is requested', () => {
        cy.visit(route + '?return_to=https://not-allowed', {failOnStatusCode: false})
        if (app === 'react') {
          cy.location('pathname').should('include', '/login')
          cy.get('.Toastify').should('contain.text', 'The return_to address is not allowed.')
        } else {
          cy.location('pathname').should('contain', 'error')
          cy.get('code').should('contain.text', 'Requested return_to URL \\"https://not-allowed\\" is not whitelisted.')
        }
      })

      describe('shows validation errors when invalid signup data is used', () => {
        beforeEach(() => {
          cy.clearCookies()
          cy.visit(route)
        })

        it('should show an error when the identifier is missing', () => {
          cy.submitPasswordForm()
          cy.get('*[data-testid="ui/message/4000001"]').should(
            'contain.text',
            'length must be >= 1, but got 0'
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
