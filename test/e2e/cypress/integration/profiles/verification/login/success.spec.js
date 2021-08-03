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

    describe('success flow', () => {
      let identity
      beforeEach(() => {
        cy.dontLoginUserAfterRegistration()
        cy.enableLoginForVerifiedAddressOnly()
        cy.deleteMail()
        cy.visit(APP_URL + '/')

        identity = gen.identity()
        cy.register(identity)
      })

      it('Is able to login after successful email verification', () => {
        cy.performEmailVerification({ expect: { email: identity.email } })

        cy.visit(APP_URL + '/auth/login')

        cy.get('input[name="password_identifier"]').type(identity.email)
        cy.get('input[name="password"]').type(identity.password)
        cy.get('button[value="password"]').click()

        cy.getCookie('ory_kratos_session').should('exist')
      })
    })
  })
})
