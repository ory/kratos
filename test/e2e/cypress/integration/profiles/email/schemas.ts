import { appPrefix, gen } from '../../../../helpers'
import { routes as express } from '../../../../helpers/express'
import { routes as react } from '../../../../helpers/react'

context('UI with email profile', () => {
  ;[
    {
      base: express.base,
      app: 'express' as 'express',
      profile: 'email'
    },
    {
      base: react.base,
      app: 'react' as 'react',
      profile: 'spa'
    }
  ].forEach(({ profile, app, base }) => {
    describe(`for app ${app}`, () => {
      const identity = gen.identity()

      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.login({ ...identity, cookieUrl: base })
        cy.visit(base)
      })

      describe('use select node', () => {
        it('should use the json schema titles', () => {
          cy.get(appPrefix(app) + 'a[href*="settings"]').click()
          cy.get('select[name="traits.color"]')
            .parent()
            .should('contain.text', 'Your color')
          cy.get('select[name="traits.color"]')
            .children('option')
            .then((options) => {
              const actual = [...options].map((o) => o.value)
              expect(actual).to.deep.eq(['blue', 'green', 'purple'])
            })
        })
      })
    })
  })
})
