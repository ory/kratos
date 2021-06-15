import {gen, MOBILE_URL, website} from '../../../../helpers'

context('Mobile Profile', () => {
  describe('Settings Flow Errors', () => {
    before(() => {
      cy.useConfigProfile('mobile')
    })

    let email, password

    before(() => {
      email = gen.email()
      password = gen.password()
      cy.registerApi({email, password, fields: {'traits.website': website}})
    })

    beforeEach(() => {
    cy.loginMobile({email, password})
    cy.visit(MOBILE_URL + "/Settings")
  })

  describe('profile', () => {
    it('fails with validation errors', () => {
      cy.get('*[data-testid="settings-profile"] input[data-testid="traits.website"]')
        .clear()
        .type('http://s')
      cy.get('*[data-testid="settings-profile"] div[data-testid="submit-form"]').click()
      cy.get('*[data-testid="field/traits.website"]').should(
        'contain.text',
        'length must be >= 10'
      )

      cy.get('*[data-testid="settings-password"]').should('exist')
    })
  })
  })
})
