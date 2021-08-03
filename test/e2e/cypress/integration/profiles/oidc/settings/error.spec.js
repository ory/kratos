import { APP_URL, gen, website } from '../../../../helpers'

context('OIDC Profile', () => {
  describe('Settings', () => {
    before(() => {
      cy.useConfigProfile('oidc')
    })

    const up = (value) => `not-${value}`
    const down = (value) => value.replace(/not-/, '')
    let email

    beforeEach(() => {
      cy.clearCookies()
      email = gen.email()

      cy.registerOidc({ email, expectSession: true, website })
      cy.visit(APP_URL + '/settings')
    })

    describe('oidc', () => {
      it('should fail to link google because id token is missing', () => {
        cy.get('button[value="google"]').click()
        cy.get('#remember').click()
        cy.get('#accept').click()

        cy.get('.messages .message').should(
          'contain.text',
          'Authentication failed because no id_token was returned. Please accept the "openid" permission and try again.'
        )
      })
    })
  })
})
