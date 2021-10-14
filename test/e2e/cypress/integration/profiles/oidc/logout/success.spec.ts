import { appPrefix, gen, website } from '../../../../helpers'
import { routes as react } from '../../../../helpers/react'
import { routes as express } from '../../../../helpers/express'

context('Social Sign Out Successes', () => {
  ;[
    {
      base: react.base,
      registration: react.registration,
      app: 'react' as 'react',
      profile: 'spa'
    },
    {
      base: express.base,
      registration: express.registration,
      app: 'express' as 'express',
      profile: 'oidc'
    }
  ].forEach(({ base, registration, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      before(() => {
        cy.clearAllCookies()
      })

      beforeEach(() => {
        cy.visit(base)
        const email = gen.email()
        cy.registerOidc({ email, website, route: registration })
      })

      it('should sign out and be able to sign in again', () => {
        cy.get(appPrefix(app) + '[data-testid=logout]').click()
        cy.noSession()
        cy.url().should('include', '/login')
      })
    })
  })
})
