import { APP_URL, assertRecoveryAddress, gen } from '../../../../helpers'

context('Recovery', () => {
  describe('successful flow', () => {
    let identity

    before(() => {
      cy.deleteMail()
    })

    beforeEach(() => {
      identity = gen.identity()
      cy.register(identity)
      cy.visit(APP_URL + '/recovery')
    })

    it('should contain the recovery address in the session', () => {
      cy.login(identity)
      cy.session().should(assertRecoveryAddress(identity))
    })

    it('should perform a recovery flow', () => {
      cy.get('input[name="email"]').type(identity.email)
      cy.get('button[value="link"]').click()

      cy.location('pathname').should('eq', '/recovery')
      cy.get('.messages .message').should(
        'have.text',
        'An email containing a recovery link has been sent to the email address you provided.'
      )
      cy.get('input[name="email"]').should(
        'have.value',
        identity.email
      )

      cy.recoverEmail({ expect: identity })

      cy.session()
      cy.location('pathname').should('eq', '/settings')

      const newPassword = gen.password()
      cy.get('input[name="password"]').clear().type(newPassword)
      cy.get('button[value="password"]').click()
      cy.get('.container').should(
        'contain.text',
        'Your changes have been saved!'
      )
      cy.get('input[name="password"]').should('be.empty')

      cy.logout()
      cy.login({ email: identity.email, password: newPassword })
    })
  })
})
