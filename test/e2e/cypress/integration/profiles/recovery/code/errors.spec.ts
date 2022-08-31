import { extractRecoveryCode, appPrefix, gen } from '../../../../helpers'
import { routes as react } from '../../../../helpers/react'
import { routes as express } from '../../../../helpers/express'

context('Account Recovery Errors', () => {
  ;[
    {
      recovery: react.recovery,
      app: 'react' as 'react',
      profile: 'spa'
    },
    {
      recovery: express.recovery,
      app: 'express' as 'express',
      profile: 'recovery'
    }
  ].forEach(({ recovery, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.deleteMail()
        cy.longRecoveryLifespan()
        cy.longCodeLifespan()
        cy.disableVerification()
        cy.enableRecovery('code')
      })

      it('shows code expired message if expired code is submitted', () => {
        cy.visit(recovery)

        cy.shortCodeLifespan()

        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.get(appPrefix(app) + "input[name='email']").type(identity.email)
        cy.get("button[value='code']").click()
        cy.recoveryEmailWithCode({ expect: { email: identity.email } })
        cy.get("button[value='code']").click()

        cy.get('[data-testid="ui/message/4060005"]').should(
          'contain.text',
          'The recovery flow expired'
        )

        cy.noSession()
      })

      it('should receive a stub email when recovering a non-existent account', () => {
        cy.visit(recovery)

        const email = gen.email()
        cy.get(appPrefix(app) + 'input[name="email"]').type(email)
        cy.get('button[value="code"]').click()

        cy.location('pathname').should('eq', '/recovery')
        cy.get('[data-testid="ui/message/1060003"]').should(
          'have.text',
          'An email containing a recovery code has been sent to the email address you provided.'
        )
        cy.get('input[name="code"]').should('be.visible')

        cy.getMail().should((message) => {
          expect(message.subject).to.equal('Account access attempted')
          expect(message.fromAddress.trim()).to.equal('no-reply@ory.kratos.sh')
          expect(message.toAddresses).to.have.length(1)
          expect(message.toAddresses[0].trim()).to.equal(email)

          const code = extractRecoveryCode(message.body)
          expect(code).to.be.null
        })
      })

      it('should cause form errors', () => {
        cy.visit(recovery)

        cy.get('button[value="code"]').click()
        cy.get('[data-testid="ui/message/4000002"]').should(
          'contain.text',
          'Property email is missing.'
        )
        cy.get('[name="method"][value="code"]').should('exist')
      })

      it('should cause non-repeating form errors after submitting empty form twice. see: #2512', () => {
        cy.visit(recovery)
        cy.get('button[value="code"]').click()
        cy.location('pathname').should('eq', '/recovery')

        cy.get('button[value="code"]').click()
        cy.get('[data-testid="ui/message/4000002"]').should(
          'contain.text',
          'Property email is missing.'
        )
        cy.get('form')
          .find('[data-testid="ui/message/4000002"]')
          .should('have.length', 1)
        cy.get('[name="method"][value="code"]').should('exist')
      })

      it('is unable to recover the email address if the code is expired', () => {
        cy.shortLinkLifespan()
        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.recoverApi({ email: identity.email })
        cy.recoverEmailButExpired({ expect: { email: identity.email } })

        cy.get('[data-testid="ui/message/4060005"]').should(
          'contain.text',
          'The recovery flow expired'
        )

        cy.noSession()
      })

      it('is unable to recover the account if the code is incorrect', () => {
        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.visit(recovery)
        cy.get(appPrefix(app) + "input[name='email']").type(identity.email)
        cy.get("button[value='code']").click()
        cy.get('[data-testid="ui/message/1060003"]').should(
          'have.text',
          'An email containing a recovery code has been sent to the email address you provided.'
        )
        cy.get("input[name='code']").type('012345') // Invalid code
        cy.get("button[value='code']").click()
        cy.get('[data-testid="ui/message/4060006"]').should(
          'have.text',
          'The recovery code is invalid or has already been used. Please try again.'
        )
        cy.noSession()
      })

      it('invalid remote recovery email template', () => {
        cy.remoteCourierRecoveryTemplates()
        const identity = gen.identityWithWebsite()
        cy.recoverApi({ email: identity.email })

        cy.getMail().then((mail) => {
          expect(mail.body).to.include(
            'this is a remote invalid recovery template'
          )
        })
      })
    })
  })
})
