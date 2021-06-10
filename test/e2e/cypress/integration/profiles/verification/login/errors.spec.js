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
      cy.dontLoginUserAfterRegistration()
      cy.enableLoginForVerifiedAddressOnly()

      identity = gen.identity()
      cy.register(identity)
    })

    it('Is unable to login as long as the email is not verified', () => {
      cy.get("form")
      cy.get('input[name="password_identifier"]').type(identity.email)
      cy.get('input[name="password"]').type(identity.password)
      cy.get('button[value="password"]').click()

      cy.get('.error').contains("Account not active yet")

      cy.getCookie('ory_kratos_session').should('not.exist')
    })
  })
})
