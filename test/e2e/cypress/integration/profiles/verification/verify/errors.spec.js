import {
  APP_URL,
  assertVerifiableAddress,
  gen,
  parseHtml,
  verifyHrefPattern
} from '../../../../helpers'

context('Verification Profile', () => {
  describe('Verify', () => {
    before(() => {
      cy.useConfigProfile('verification')
    })

    describe('error flow', () => {
      let identity
      before(() => {
        cy.deleteMail()
      })

      beforeEach(() => {
        cy.clearCookies({ domain: null })
        cy.longVerificationLifespan()
        cy.longLinkLifespan()

        identity = gen.identity()
        cy.register(identity)
        cy.deleteMail({ atLeast: 1 }) // clean up registration email

        cy.clearCookies({ domain: null })
        cy.login(identity)
        cy.visit(APP_URL + '/verify')
      })

      it('is unable to verify the email address if the code is expired', () => {
        cy.shortLinkLifespan()

        cy.visit(APP_URL + '/verify')
        cy.get('input[name="email"]').type(identity.email)
        cy.get('button[value="link"]').click()

        cy.get('.messages .message').should(
          'contain.text',
          'An email containing a verification'
        )

        cy.verifyEmailButExpired({ expect: { email: identity.email } })
      })

      it('is unable to verify the email address if the code is incorrect', () => {
        cy.get('input[name="email"]').type(identity.email)
        cy.get('button[value="link"]').click()

        cy.get('.messages .message').should(
          'contain.text',
          'An email containing a verification'
        )

        cy.getMail().then((mail) => {
          const link = parseHtml(mail.body).querySelector('a')

          expect(verifyHrefPattern.test(link.href)).to.be.true

          cy.visit(link.href + '-not') // add random stuff to the confirm challenge
          cy.log(link.href)
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
