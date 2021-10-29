import { routes as express } from '../../../../helpers/express'
import { routes as react } from '../../../../helpers/react'
import { appPrefix } from '../../../../helpers'

describe('Handling self-service error flows', () => {
  ;[
    {
      route: express.base,
      app: 'express' as 'express',
      profile: 'email'
    },
    {
      route: react.base,
      app: 'react' as 'react',
      profile: 'spa'
    }
  ].forEach(({ route, app, profile }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      it('should show the error', () => {
        cy.visit(`${route}/error?id=stub:500`, {
          failOnStatusCode: false
        })

        cy.get(`${appPrefix(app)}code`).should(
          'contain.text',
          'This is a stub error.'
        )
      })
    })
  })
})
