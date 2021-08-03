import { APP_URL, assertRecoveryAddress, gen } from '../../../../helpers'

context('Recovery Profile', () => {
  describe('Recovery', () => {
    before(() => {
      cy.useConfigProfile('recovery')
    })

    describe('settings flow', () => {
      let identity

      before(() => {
        cy.deleteMail()
      })

      beforeEach(() => {
        identity = gen.identity()
        cy.register(identity)
        cy.login(identity)
        cy.visit(APP_URL + '/settings')
      })

      const up = (id) => `next-${id}`

      it('should update the recovery address when updating the email', () => {
        const email = up(identity.email)
        cy.get('input[name="traits.email"]').clear().type(email)
        cy.get('button[value="profile"]').click()
        cy.get('.container').should(
          'contain.text',
          'Your changes have been saved!'
        )
        cy.get('input[name="traits.email"]').should('contain.value', email)

        cy.session().should(assertRecoveryAddress({ email }))
      })

      xit('should not show an immediate error when a recovery address already exists', () => {
        // account enumeration prevention, needs to be implemented.
      })
    })
  })
})
