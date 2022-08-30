import {
  extractRecoveryCode,
  gen,
  MOBILE_URL,
  website
} from '../../../../helpers'

context('Mobile Profile', () => {
  describe('Recovery Flow Success', () => {
    before(() => {
      cy.enableRecovery('code')
      cy.useConfigProfile('mobile')
    })

    describe('code', () => {
      const email = gen.email()
      const password = gen.password()

      before(() => {
        cy.registerApi({
          email,
          password,
          fields: { 'traits.website': website }
        })
      })

      beforeEach(() => {
        cy.deleteMail()
        cy.visit(MOBILE_URL + '/Recovery')
      })

      it('recovers the account', () => {
        cy.get('*[data-testid="field/email"] input[data-testid="email"]')
          .clear()
          .type(email)
        cy.get(
          '*[data-testid="field/method/code"] div[data-testid="submit-form"]'
        ).click()

        cy.get('*[data-testid="ui/message/1060003"]').should(
          'contain.text',
          'An email containing a recovery code has been sent to the email address you provided.'
        )

        cy.getMail().should((message) => {
          expect(message.subject).to.equal('Recover access to your account')
          expect(message.toAddresses[0].trim()).to.equal(email)

          const code = extractRecoveryCode(message.body)
          expect(code).to.not.be.undefined
          expect(code.length).to.equal(8)

          cy.get('*[data-testid="field/code"] input[data-testid="code"]').type(
            code
          )
        })
        cy.get(
          '*[data-testid="field/method/code"] div[data-testid="submit-form"]'
        ).click()
        cy.get('[data-testid="session-token"]').should('not.be.empty')
        cy.get('[data-testid="ui/message/1060001"]').should(
          'contain.text',
          'You successfully recovered your account.'
        )
      })
    })
  })
})
