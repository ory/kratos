import {
  APP_URL,
  assertVerifiableAddress,
  gen,
  parseHtml,
  verifyHrefPattern,
} from '../../../../helpers'

context('Verification Profile', () => {
  describe('Login', () => {
    before(() => {
      cy.useConfigProfile('verification')
    })

    describe('error flow', () => {
      let identity
      beforeEach(() => {
        cy.dontLoginUserAfterRegistration()
        cy.enableLoginForVerifiedAddressOnly()
        cy.deleteMail()
        cy.visit(APP_URL + '/')

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
})
