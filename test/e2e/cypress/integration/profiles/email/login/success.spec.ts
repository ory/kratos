import { APP_URL, appPrefix, gen, website } from '../../../../helpers'
import { routes as express } from '../../../../helpers/express'
import { routes as react } from '../../../../helpers/react'

describe('Basic email profile with succeeding login flows', () => {
  const email = gen.email()
  const password = gen.password()

  before(() => {
    cy.registerApi({ email, password, fields: { 'traits.website': website } })
  })
  ;[
    {
      route: express.login,
      app: 'express' as 'express',
      profile: 'email'
    },
    {
      route: react.login,
      app: 'react' as 'react',
      profile: 'spa'
    }
  ].forEach(({ route, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.useConfigProfile(profile)
        cy.clearAllCookies()
        cy.visit(route)
      })

      it('should sign in and be logged in', () => {
        cy.get(`${appPrefix(app)}input[name="password_identifier"]`).type(email)
        cy.get('input[name="password"]').type(password)
        cy.submitPasswordForm()
        cy.location('pathname').should('not.contain', '/login')

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.schema_id).to.equal('default')
          expect(identity.schema_url).to.equal(`${APP_URL}/schemas/default`)
          expect(identity.traits.website).to.equal(website)
          expect(identity.traits.email).to.equal(email)
        })
      })

      it('should sign in with case insensitive identifier surrounded by whitespace', () => {
        cy.get('input[name="password_identifier"]').type(
          '  ' + email.toUpperCase() + '  '
        )
        cy.get('input[name="password"]').type(password)
        cy.submitPasswordForm()
        cy.location('pathname').should('not.contain', '/login')

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.schema_id).to.equal('default')
          expect(identity.schema_url).to.equal(`${APP_URL}/schemas/default`)
          expect(identity.traits.website).to.equal(website)
          expect(identity.traits.email).to.equal(email)
        })
      })

      it('should sign in and be redirected', () => {
        cy.browserReturnUrlOry()
        cy.visit(route + '?return_to=https://www.ory.sh/')

        cy.get('input[name="password_identifier"]').type(email.toUpperCase())
        cy.get('input[name="password"]').type(password)
        cy.submitPasswordForm()

        cy.url().should('eq', 'https://www.ory.sh/')
      })
    })
  })

  describe('for app express handle return_to correctly for expired flows', () => {
    before(() => {
      cy.proxy('express')
      cy.useConfigProfile('email')

      cy.browserReturnUrlOry()
    })

    beforeEach(() => {
      cy.clearAllCookies()
    })

    it('should redirect to return_to when retrying expired flow', () => {
      cy.shortLoginLifespan()
      cy.wait(500)

      cy.visit(express.login + '?return_to=https://www.ory.sh/')

      cy.longLoginLifespan()

      cy.get(appPrefix('express') + 'input[name="password_identifier"]').type(
        email.toUpperCase()
      )
      cy.get('input[name="password"]').type(password)

      cy.submitPasswordForm()
      cy.get('[data-testid="ui/message/4010001"]').should(
        'contain.text',
        'The login flow expired'
      )

      // try again with long lifespan set
      cy.get('input[name="password_identifier"]').type(email.toUpperCase())
      cy.get('input[name="password"]').type(password)
      cy.submitPasswordForm()

      // check that redirection has happened
      cy.url().should('eq', 'https://www.ory.sh/')
    })
  })
})
