import {
  APP_URL,
  assertVerifiableAddress,
  gen,
  parseHtml,
  verifyHrefPattern,
} from '../../../../helpers'

context('Verify', () => {
  describe('success flow', () => {
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

    it('Is able to login after successful email verification', () => {
      cy.verifyEmail({ expect: { email: identity.email } })

      cy.visit(APP_URL)

      cy.get("form")
      cy.get('input[name="password_identifier"]').type(identity.email)
      cy.get('input[name="password"]').type(identity.password)
      cy.get('button[value="password"]').click()

      cy.getCookie('ory_kratos_session').should('exist')
    })
  })
})
