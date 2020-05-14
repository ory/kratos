import { APP_URL } from '../../../../helpers'

context('Login', () => {
  beforeEach(() => {
    cy.visit(APP_URL + '/auth/login')
  })

  describe('use ui elements', () => {
    it('clicks the log in link', () => {
      cy.get('a[href*="auth/registration"]').click()
      cy.location('pathname').should('include', 'auth/registration')
      cy.location('search').should('not.be.empty', 'request')
    })
  })
})
