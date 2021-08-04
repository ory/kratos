import { APP_URL, gen } from '../../../../helpers'

context('Verification Profile', () => {
  describe('Login', () => {
    before(() => {
      cy.useConfigProfile('verification')
      cy.enableLoginForVerifiedAddressOnly()
    })

    describe('success flow', () => {
      it('Is able to login after successful email verification', () => {
        cy.deleteMail()

        const identity = gen.identity()
        cy.register(identity)
        cy.performEmailVerification({ expect: { email: identity.email } })

        cy.visit(APP_URL + '/')

        cy.get('input[name="password_identifier"]').type(identity.email)
        cy.get('input[name="password"]').type(identity.password)
        cy.get('button[value="password"]').click()

        cy.getCookie('ory_kratos_session').should('exist')
      })
    })
  })
})
