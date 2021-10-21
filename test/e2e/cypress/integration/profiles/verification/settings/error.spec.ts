import {
  appPrefix,
  assertVerifiableAddress,
  gen,
  parseHtml,
  verifyHrefPattern
} from '../../../../helpers'

import { routes as react } from '../../../../helpers/react'
import { routes as express } from '../../../../helpers/express'

context('Account Verification Settings Error', () => {
  ;[
    {
      settings: react.settings,
      base: react.base,
      app: 'react' as 'react',
      profile: 'verification'
    },
    {
      settings: express.settings,
      base: express.base,
      app: 'express' as 'express',
      profile: 'verification'
    }
  ].forEach(({ profile, settings, app, base }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      describe('error flow', () => {
        let identity
        before(() => {
          cy.deleteMail()
        })

        beforeEach(() => {
          cy.longLinkLifespan()
          identity = gen.identityWithWebsite()
          cy.clearAllCookies()
          cy.registerApi(identity)
          cy.deleteMail({ atLeast: 1 }) // clean up registration email

          cy.login({ ...identity, cookieUrl: base })
          cy.visit(settings)
        })

        it('is unable to verify the email address if the code is no longer valid', () => {
          cy.shortLinkLifespan()
          cy.visit(settings)

          const email = `not-${identity.email}`
          cy.get(appPrefix(app) + 'input[name="traits.email"]')
            .clear()
            .type(email)
          cy.get('button[value="profile"]').click()

          cy.verifyEmailButExpired({
            expect: { email, password: identity.password }
          })
        })

        it('is unable to verify the email address if the code is incorrect', () => {
          const email = `not-${identity.email}`
          cy.get('input[name="traits.email"]').clear().type(email)
          cy.get('button[value="profile"]').click()

          cy.getMail().then((mail) => {
            const link = parseHtml(mail.body).querySelector('a')

            expect(verifyHrefPattern.test(link.href)).to.be.true

            cy.visit(link.href + '-not') // add random stuff to the confirm challenge
            cy.log(link.href)
            cy.getSession().then(
              assertVerifiableAddress({ isVerified: false, email })
            )
          })
        })

        xit('should not update the traits until the email has been verified and the old email has accepted the change', () => {
          // FIXME https://github.com/ory/kratos/issues/292
        })
      })
    })
  })
})
