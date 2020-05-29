import {
  APP_URL,
  assertVerifiableAddress,
  gen,
  parseHtml,
  verifyHrefPattern,
} from '../../../../helpers'

context('Verify', () => {
  describe('error flow', () => {
    let identity
    before(() => {
      cy.deleteMail()
    })

    beforeEach(() => {
      identity = gen.identity()
      cy.register(identity)
      cy.deleteMail({ atLeast: 1 }) // clean up registration email

      cy.login(identity)
      cy.visit(APP_URL + '/verify')
    })

    it('is unable to verify the email address if the code is expired', () => {
      cy.get('input[name="to_verify"]').type(identity.email)
      cy.get('button[type="submit"]').click()

      cy.location('pathname').should('eq', '/')

      cy.verifyEmailButExpired({ expect: { email: identity.email } })
    })

    it('is unable to verify the email address if the code is incorrect', () => {
      cy.get('input[name="to_verify"]').type(identity.email)
      cy.get('button[type="submit"]').click()

      cy.location('pathname').should('eq', '/')

      cy.getMail().then((mail) => {
        const link = parseHtml(mail.body).querySelector('a')

        expect(verifyHrefPattern.test(link.href)).to.be.true

        cy.visit(link.href + '-not') // add random stuff to the confirm challenge
        cy.log(link.href)
        cy.session().then(
          assertVerifiableAddress({ isVerified: false, email: identity.email })
        )
      })
    })
  })
})
