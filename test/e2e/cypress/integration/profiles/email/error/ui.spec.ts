import { APP_URL } from '../../../../helpers'

context('Email Profile', () => {
  describe('Self-Service Error UI', () => {
    before(() => {
      cy.useConfigProfile('email')
    })

    it('should show the error', () => {
      cy.visit(`${APP_URL}/error?id=stub:500`, {
        failOnStatusCode: false
      })

      cy.get('code').should('contain.text', 'This is a stub error.')
    })
  })
})
