import {APP_URL, assertAddress, gen, parseHtml, verifyHrefPattern,} from '../../../../helpers'

context('Registration', () => {
  describe('error flow', () => {
    let identity
    beforeEach(() => {
      cy.visit(APP_URL + '/auth/registration')
      cy.deleteMail()

      identity = gen.identity()
      cy.register(identity)
      cy.login(identity)
    })

    it('is unable to verify the email address if the code is no longer valid and resend the code', () => {
      cy.verifyEmailButExpired({expect: {email: identity.email}})

      cy.get('input[name="to_verify"]').should('be.empty')
      cy.get('input[name="to_verify"]').type(identity.email)
      cy.get('button[type="submit"]').click()

      cy.location('pathname').should('eq', '/')

      cy.verifyEmail({expect: {email: identity.email}})
    })

    it('is unable to verify the email address if the code is incorrect', () => {
      cy.getMail().then((mail) => {
        const link = parseHtml(mail.body).querySelector('a')

        expect(verifyHrefPattern.test(link.href)).to.be.true

        cy.visit(link.href + '-not') // add random stuff to the confirm challenge
        cy.log(link.href)
        cy.session().then(
          assertAddress({isVerified: false, email: identity.email})
        )
      })
    })
  })
})
