import { APP_URL } from '../../../../helpers'

context('Email Profile', () => {
  describe('Login Flow UI', () => {
    before(() => {
      cy.useConfigProfile('email')
    })

    beforeEach(() => {
      cy.visit(APP_URL + '/auth/login')
    })

    describe('use ui elements', () => {
      it('should use the json schema titles', () => {
        cy.get('input[name="password_identifier"]')
          .siblings('span')
          .should('contain.text', 'ID')
        cy.get('input[name="password"]')
          .siblings('span')
          .should('contain.text', 'Password')
        cy.get('button[value="password"]').should('contain.text', 'Sign in')
      })

      it('clicks the log in link', () => {
        cy.get('a[href*="auth/registration"]').click()
        cy.location('pathname').should('include', 'auth/registration')
        cy.location('search').should('not.be.empty', 'request')
      })
    })
  })
})
