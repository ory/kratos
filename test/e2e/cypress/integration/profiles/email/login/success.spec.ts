import {APP_URL, gen, website} from '../../../../helpers'
import {routes as express} from "../../../../helpers/express";
import {routes as react} from "../../../../helpers/react";

describe('Basic email profile with succeeding login flows', () => {
  const email = gen.email()
  const password = gen.password()

  before(() => {
    cy.registerApi({email, password, fields: {'traits.website': website}})
  });

  [{
    route: express.login,
    app: 'express' as 'express', profile: 'email'
  }, {
    route: react.login,
    app: 'react' as 'react', profile: 'spa'
  }].forEach(({route, profile, app}) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.useConfigProfile(profile)
        cy.clearCookies()
        cy.visit(route)
        cy.ensureCorrectApp(app)
      })

      it('should sign in and be logged in', () => {
        cy.get('input[name="password_identifier"]').type(email)
        cy.get('input[name="password"]').type(password)
        cy.submitPasswordForm()

        cy.getSession().should((session) => {
          const {identity} = session
          expect(identity.id).to.not.be.empty
          expect(identity.schema_id).to.equal('default')
          expect(identity.schema_url).to.equal(`${APP_URL}/schemas/default`)
          expect(identity.traits.website).to.equal(website)
          expect(identity.traits.email).to.equal(email)
        })
      })

      it('should sign in with case insensitive identifier', () => {
        cy.get('input[name="password_identifier"]').type(email.toUpperCase())
        cy.get('input[name="password"]').type(password)
        cy.submitPasswordForm()

        cy.getSession().should((session) => {
          const {identity} = session
          expect(identity.id).to.not.be.empty
          expect(identity.schema_id).to.equal('default')
          expect(identity.schema_url).to.equal(`${APP_URL}/schemas/default`)
          expect(identity.traits.website).to.equal(website)
          expect(identity.traits.email).to.equal(email)
        })
      })

      it('should sign in and be redirected', () => {
        cy.browserReturnUrlOry()
        cy.visit(
          route + '?return_to=https://www.ory.sh/'
        )

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

      cy.shortLoginLifespan()
      cy.browserReturnUrlOry()
      cy.clearAllCookies()

      cy.visit(
        express.login + "?return_to=https://www.ory.sh/"
      )
      cy.ensureCorrectApp('express')
    })

    it('should redirect to return_to when retrying expired flow', () => {
      cy.get('input[name="password_identifier"]').type(email.toUpperCase())
      cy.get('input[name="password"]').type(password)

      cy.longLoginLifespan()
      cy.submitPasswordForm()
      cy.get('.messages .message').should(
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
