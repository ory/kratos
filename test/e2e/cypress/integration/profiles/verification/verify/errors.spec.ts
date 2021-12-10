import {
  APP_URL,
  assertVerifiableAddress,
  gen,
  parseHtml,
  verifyHrefPattern
} from '../../../../helpers'
import { routes as react } from '../../../../helpers/react'
import { routes as express } from '../../../../helpers/express'

context('Account Verification Error', () => {
  ;[
    {
      verification: react.verification,
      base: react.base,
      app: 'react' as 'react',
      profile: 'verification'
    },
    {
      verification: express.verification,
      base: express.base,
      app: 'express' as 'express',
      profile: 'verification'
    }
  ].forEach(({ profile, verification, app, base }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      let identity
      before(() => {
        cy.deleteMail()
      })

      beforeEach(() => {
        cy.clearAllCookies()
        cy.longVerificationLifespan()
        cy.longLinkLifespan()

        identity = gen.identity()
        cy.registerApi(identity)
        cy.deleteMail({ atLeast: 1 }) // clean up registration email
        cy.login(identity)
        cy.visit(verification)
      })

      it('is unable to verify the email address if the code is expired', () => {
        cy.shortLinkLifespan()

        cy.visit(verification)
        cy.get('input[name="email"]').type(identity.email)
        cy.get('button[value="link"]').click()

        cy.get('[data-testid="ui/message/1080001"]').should(
          'contain.text',
          'An email containing a verification'
        )

        cy.verifyEmailButExpired({ expect: { email: identity.email } })
      })

      it('is unable to verify the email address if the code is incorrect', () => {
        cy.get('input[name="email"]').type(identity.email)
        cy.get('button[value="link"]').click()

        cy.get('[data-testid="ui/message/1080001"]').should(
          'contain.text',
          'An email containing a verification'
        )

        cy.getMail().then((mail) => {
          const link = parseHtml(mail.body).querySelector('a')

          expect(verifyHrefPattern.test(link.href)).to.be.true

          cy.visit(link.href + '-not') // add random stuff to the confirm challenge
          cy.getSession().then(
            assertVerifiableAddress({
              isVerified: false,
              email: identity.email
            })
          )
        })
      })
    })
  })
})
