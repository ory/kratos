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

      const email = gen.email()

      before(() => {
        cy.clearAllCookies()
        cy.registerOidc({ email, website, route: registration })
      })

      beforeEach(() => {
        cy.visit(base)
      })

      it('should sign out and be able to sign in again', () => {
        cy.get(appPrefix(app) + '[data-testid=logout]').click()
        cy.noSession()
        cy.url().should('include', '/login')
      })
    })
  })
})
