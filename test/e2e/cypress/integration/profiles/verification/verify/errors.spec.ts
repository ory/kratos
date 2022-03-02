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

      it('responds with a HTML response on link click of an API flow if the flow is expired', () => {
        cy.updateConfigFile((config) => {
          config.selfservice.flows.verification.lifespan = '1s'
          return config
        })

        cy.verificationApi({
          email: identity.email
        })

        cy.wait(1000)
        cy.shortVerificationLifespan()

        cy.getMail().then((message) => {
          expect(message.subject.trim()).to.equal(
            'Please verify your email address'
          )
          expect(message.toAddresses[0].trim()).to.equal(identity.email)

          const link = parseHtml(message.body).querySelector('a')

          cy.longVerificationLifespan()
          cy.visit(link.href)
          cy.get('[data-testid="ui/message/4070005"]').should(
            'contain.text',
            'verification flow expired'
          )

          cy.getSession().should((session) => {
            assertVerifiableAddress({
              isVerified: false,
              email: identity.email
            })(session)
          })
        })
      })

      it('responds with a HTML response on link click of an API flow if the link is expired', () => {
        cy.shortLinkLifespan()

        // Init expired flow
        cy.verificationApi({
          email: identity.email
        })

        cy.verifyEmailButExpired({ expect: { email: identity.email } })
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
        cy.get('[name="method"][value="link"]').should('exist')
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

      it('unable to verify non-existent account', async () => {
        cy.get('input[name="email"]').type(gen.identity().email)
        cy.get('button[value="link"]').click()
        cy.getMail().then((mail) => {
          expect(mail.subject).eq(
            'Someone tried to verify this email address (remote)'
          )
        })
      })
    })
  })
})
