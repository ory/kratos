import {gen} from '../../../../helpers'
import {routes as express} from "../../../../helpers/express";
import {routes as react} from "../../../../helpers/react";

describe('Registration failures with email profile', () => {
  [
    {
      route: express.registration,
      app: 'express', profile: 'email'
    },
    {
      route: react.registration,
      app: 'react', profile: 'spa'
    }
  ].forEach(({route, profile, app}) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
      })

      beforeEach(() => {
        cy.visit(route)
      })

      const identity = gen.email()
      const password = gen.password()

      it('fails when CSRF cookies are missing', () => {
        cy.clearCookies()

        cy.get('input[name="traits.website"]').type('https://www.ory.sh')
        cy.get('input[name="traits.email"]')
          .type(identity)
          .should('have.value', identity)
        cy.get('input[name="password"]')
          .type('123456')
          .should('have.value', '123456')

        let initial
        cy.location().should((location) => {
          initial = location.search
        })
        cy.get('button[type="submit"]').click()

        // We end up at a new flow
        cy.location('search').should('not.eq', initial)
        if (app === 'express') {
          cy.location('pathname').should('include', '/error')
          cy.get('code').should('contain.text', 'csrf_token')
        } else {
          cy.location('pathname').should('include', '/registration')
          cy.get('.Toastify').should('contain.text', 'A security violation was detected, please fill out the form again.')
        }
      })

      it('fails when a disallowed return_to url is requested', () => {
        cy.visit(route + '?return_to=https://not-allowed', {failOnStatusCode: false})
        if (app === 'react') {
          cy.location('pathname').should('include', '/registration')
          cy.get('.Toastify').should('contain.text', 'The return_to address is not allowed.')
        } else {
          cy.location('pathname').should('contain', 'error')
          cy.get('code').should('contain.text', 'Requested return_to URL \\"https://not-allowed\\" is not whitelisted.')
        }
      })

      describe('show errors when invalid signup data is used', () => {
        it('should show an error when the password has leaked before', () => {
          cy.get('input[name="traits.website"]').type('https://www.ory.sh')
          cy.get('input[name="traits.email"]')
            .type(identity)
            .should('have.value', identity)
          cy.get('input[name="password"]')
            .type('123456')
            .should('have.value', '123456')

          cy.get('button[type="submit"]').click()
          cy.get('*[data-testid^="ui.node.message"]').should('contain.text', 'data breaches')
        })

        it('should show an error when the password is too similar', () => {
          cy.get('input[name="traits.website"]').type('https://www.ory.sh')
          cy.get('input[name="traits.email"]').type(identity)
          cy.get('input[name="password"]').type(identity)

          cy.get('button[type="submit"]').click()
          cy.get('*[data-testid^="ui.node.message"]').should('contain.text', 'too similar')
        })

        it('should show an error when the password is empty', () => {
          cy.get('input[name="traits.website"]').type('https://www.ory.sh')
          cy.get('input[name="traits.email"]').type(identity)

          cy.get('button[type="submit"]').click()

          cy.get('*[data-testid^="ui.node.message."]').invoke('text').then((text) => {
            expect(text).to.be.oneOf(['length must be >= 1, but got 0', 'Property password is missing.'])
          })
        })

        it('should show an error when the email is empty', () => {
          cy.get('input[name="traits.website"]').type('https://www.ory.sh')
          cy.get('input[name="password"]').type(password)

          cy.get('button[type="submit"]').click()
          cy.get('*[data-testid^="ui.node.message."]').invoke('text').then((text) => {
            expect(text).to.be.oneOf(['"" is not valid "email"length must be >= 3, but got 0', 'Property email is missing.'])
          })
        })

        it('should show an error when the email is not an email', () => {
          cy.get('input[name="traits.website"]').type('https://www.ory.sh')
          cy.get('input[name="password"]').type('not-an-email')

          cy.get('button[type="submit"]').click()
          cy.get('*[data-testid="ui.node.message.4000001"], *[data-testid="ui.node.message.4000002"]').should('exist')
        })

        it('should show a missing indicator if no fields are set', () => {
          cy.get('button[type="submit"]').click()
          cy.get('*[data-testid="ui.node.message.4000001"], *[data-testid="ui.node.message.4000002"]').should('exist')
        })

        it('should show an error when the website is not a valid URI', () => {
          cy.get('input[name="traits.website"]')
            .type('1234')
            .then(($input) => {
              expect(($input[0] as HTMLInputElement).validationMessage).to.contain('URL')
            })
        })

        it('should show an error when the website is too short', () => {
          cy.get('input[name="traits.website"]').type('http://s')

          // fixme https://github.com/ory/kratos/issues/368
          cy.get('input[name="password"]').type(password)

          cy.get('button[type="submit"]').click()
          cy.get('*[data-testid^="ui.node.message"]').should(
            'contain.text',
            'length must be >= 10'
          )
        })
      })
    })
  })
})
