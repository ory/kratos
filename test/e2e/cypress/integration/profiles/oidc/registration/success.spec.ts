import {APP_URL, appPrefix, gen, website} from '../../../../helpers'
import {routes as react} from "../../../../helpers/react";
import {routes as express} from "../../../../helpers/express";

context('Social Sign Up Successes', () => {
  [
    {
      login: react.login,
      registration: react.registration,
      app: 'react' as 'react', profile: 'spa'
    },
    {
      login: express.login,
      registration: express.registration,
      app: 'express' as 'express', profile: 'oidc'
    }
  ].forEach(({registration, login, profile, app}) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.clearAllCookies()
        cy.visit(registration)
      })

      const shouldSession = (email) => (session) => {
        const {identity} = session
        expect(identity.id).to.not.be.empty
        expect(identity.schema_id).to.equal('default')
        expect(identity.schema_url).to.equal(`${APP_URL}/schemas/default`)
        expect(identity.traits.website).to.equal(website)
        expect(identity.traits.email).to.equal(email)
      }

      it('should be able to sign up with incomplete data and finally be signed in', () => {
        const email = gen.email()

        cy.registerOidc({email, expectSession: false, route: registration})

        cy.get('#registration-password').should('not.exist')
        cy.get(appPrefix(app)+'[name="traits.email"]').should('have.value', email)
        cy.get('[data-testid="ui/message/4000002"]').should(
          'contain.text',
          'Property website is missing'
        )

        cy.get('[name="traits.consent"][type="checkbox"]').siblings('label').click()
        cy.get('[name="traits.newsletter"][type="checkbox"]').siblings('label').click()
        cy.get('[name="traits.website"]').type('http://s')

        cy.get('[name="provider"]')
          .should('have.length', 1)
          .should('have.value', 'hydra')
          .should('contain.text', 'Continue')
          .click()

        cy.get('#registration-password').should('not.exist')
        cy.get('[name="traits.email"]').should('have.value', email)
        cy.get('[name="traits.website"]').should('have.value', 'http://s')
        cy.get('[data-testid="ui/message/4000001"]').should(
          'contain.text',
          'length must be >= 10'
        )
        cy.get('[name="traits.website"]')
          .should('have.value', 'http://s')
          .clear()
          .type(website)

        cy.get('[name="traits.consent"]').should('be.checked')
        cy.get('[name="traits.newsletter"]').should('be.checked')

        cy.triggerOidc()

        cy.getSession().should((session) => {
          shouldSession(email)(session)
          expect(session.identity.traits.consent).to.equal(true)
        })
      })

      it('should be able to sign up with complete data', () => {
        const email = gen.email()

        cy.registerOidc({email, website, route: registration})
        cy.getSession().should(shouldSession(email))
      })

      it('should be able to convert a sign up flow to a sign in flow', () => {
        const email = gen.email()

        cy.registerOidc({email, website, route: registration})
        cy.logout()
        cy.noSession()
        cy.visit(registration)
        cy.triggerOidc()

        cy.getSession().should(shouldSession(email))
      })

      it('should be able to convert a sign in flow to a sign up flow', () => {
        const email = gen.email()
        cy.visit(login)
        cy.triggerOidc()

        cy.get('#username').clear().type(email)
        cy.get('#remember').click()
        cy.get('#accept').click()
        cy.get('[name="scope"]').each(($el) => cy.wrap($el).click())
        cy.get('#remember').click()
        cy.get('#accept').click()

        cy.get('[data-testid="ui/message/4000002"]').should(
          'contain.text',
          'Property website is missing'
        )
        cy.get('[name="traits.website"]').type('http://s')

        cy.triggerOidc()

        cy.get('[data-testid="ui/message/4000001"]').should(
          'contain.text',
          'length must be >= 10'
        )
        cy.get('[name="traits.website"]')
          .should('have.value', 'http://s')
          .clear()
          .type(website)
        cy.triggerOidc()

        cy.getSession().should(shouldSession(email))
      })

      it('should be able to sign up with redirects', () => {
        const email = gen.email()
        cy.registerOidc({email, website, route: registration + '?return_to=https://www.ory.sh/'})
        cy.location('href').should('eq','https://www.ory.sh/')
        cy.logout()
      })
    })
  })
})
