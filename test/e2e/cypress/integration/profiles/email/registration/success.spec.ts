import {APP_URL, gen} from '../../../../helpers'
import {routes as express} from "../../../../helpers/express";
import {routes as react} from "../../../../helpers/react";

context('Registration success with email profile', () => {
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

      it('should sign up and be logged in', () => {
        const email = gen.email()
        const password = gen.password()
        const website = 'https://www.ory.sh/'

        cy.get('input[name="traits"]').should('not.exist')
        cy.get('input[name="traits.email"]').type(email)
        cy.get('input[name="traits.website').type(website)
        cy.get('input[name="password"]').type(password)

        cy.get('button[type="submit"]').click()
        cy.get('pre').should('contain.text', email)

        cy.getSession().should((session) => {
          const {identity} = session
          expect(identity.id).to.not.be.empty
          expect(identity.verifiable_addresses).to.be.undefined
          expect(identity.schema_id).to.equal('default')
          expect(identity.schema_url).to.equal(`${APP_URL}/schemas/default`)
          expect(identity.traits.website).to.equal(website)
          expect(identity.traits.email).to.equal(email)
        })
      })

      it('should sign up and be redirected', () => {
        cy.browserReturnUrlOry()
        cy.visit(
          route + '?return_to=https://www.ory.sh/'
        )

        const email = gen.email()
        const password = gen.password()
        const website = 'https://www.ory.sh/'

        cy.get('input[name="traits"]').should('not.exist')
        cy.get('input[name="traits.email"]').type(email)
        cy.get('input[name="traits.website').type(website)
        cy.get('input[name="password"]').type(password)

        cy.get('button[type="submit"]').click()
        cy.url().should('eq', 'https://www.ory.sh/')
      })
    })
  })

  describe('redirect for express app',() => {
    it('should redirect to return_to after flow expires', () => {
      // Wait for flow to expire
      cy.useConfigProfile('email')
      cy.shortRegisterLifespan()
      cy.browserReturnUrlOry()
      cy.visit(
        express.registration + '?return_to=https://www.ory.sh/'
      )
      cy.wait(105)

      const email = gen.email()
      const password = gen.password()
      const website = 'https://www.ory.sh/'

      cy.get('input[name="traits"]').should('not.exist')
      cy.get('input[name="traits.email"]').type(email)
      cy.get('input[name="traits.website').type(website)
      cy.get('input[name="password"]').type(password)

      cy.longRegisterLifespan()
      cy.get('button[type="submit"]').click()

        cy.get('*[data-testid^="ui/message/"]').should(
          'contain.text',
          'The registration flow expired'
        )

      // Try again with long lifespan set
      cy.get('input[name="traits"]').should('not.exist')
      cy.get('input[name="traits.email"]').type(email)
      cy.get('input[name="traits.website').type(website)
      cy.get('input[name="password"]').type(password)
      cy.get('button[type="submit"]').click()

      cy.url().should('eq', 'https://www.ory.sh/')
    })
  })
})
