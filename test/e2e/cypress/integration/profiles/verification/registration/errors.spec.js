import {
  APP_URL,
  assertVerifiableAddress,
  gen,
  parseHtml,
  verifyHrefPattern
} from '../../../../helpers'

context('Verification Profile', () => {
  describe('Registration', () => {
    before(() => {
      cy.useConfigProfile('verification')
    })

    describe('error flow', () => {
      let identity
      beforeEach(() => {
        cy.shortLinkLifespan()
        cy.longVerificationLifespan()

        cy.visit(APP_URL + '/')
        cy.deleteMail()

        identity = gen.identity()
        cy.register(identity)
        cy.login(identity)
      })

      it('is unable to verify the email address if the code is no longer valid and resend the code', () => {
        cy.verifyEmailButExpired({ expect: { email: identity.email } })

        cy.longLinkLifespan()

        cy.get('input[name="email"]').should('be.empty')
        cy.get('input[name="email"]').type(identity.email)
        cy.get('button[value="link"]').click()
        cy.get('.messages .message').should(
          'contain.text',
          'An email containing a verification'
        )
        cy.verifyEmail({ expect: { email: identity.email } })
      })

      it('is unable to verify the email address if the code is incorrect', () => {
        cy.getMail().then((mail) => {
          const link = parseHtml(mail.body).querySelector('a')

          console.log(link.href)
          expect(verifyHrefPattern.test(link.href)).to.be.true

          cy.visit(link.href + '-not') // add random stuff to the confirm challenge
          cy.session().then(
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
