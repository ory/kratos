import { APP_URL, gen, website } from '../../../../helpers'

context('Email Profile', () => {
  describe('Logout Flow Success', () => {
    before(() => {
      cy.useConfigProfile('email')
    })

    const email = gen.email()
    const password = gen.password()

    before(() => {
      cy.register({ email, password, fields: { 'traits.website': website } })
      cy.login({ email, password })
    })

    it('should sign out and be able to sign in again', () => {
      cy.visit(APP_URL + '/')
      cy.session()
      cy.get('a[href*="logout"]').click()
      cy.noSession()
      cy.url().should('include', '/auth/login')
    })
  })
})
