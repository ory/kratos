import { APP_URL, gen } from '../../../../helpers'

context('Verification Profile', () => {
  describe('Login', () => {
    before(() => {
      cy.useConfigProfile('verification')
      cy.enableLoginForVerifiedAddressOnly()
    })

    describe('error flow', () => {
      it('Is unable to login as long as the email is not verified', () => {
        cy.deleteMail()

        const identity = gen.identity()
        cy.register(identity)

        cy.visit(APP_URL + '/')

        cy.get('input[name="password_identifier"]').type(identity.email)
        cy.get('input[name="password"]').type(identity.password)
        cy.get('button[value="password"]').click()

        cy.get('.error').contains('Account not active yet')

        cy.getCookie('ory_kratos_session').should('not.exist')
      })
    })
  })
})
