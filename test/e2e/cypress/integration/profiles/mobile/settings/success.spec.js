import { gen, MOBILE_URL, website } from '../../../../helpers'

context('Mobile Profile', () => {
  describe('Login Flow Success', () => {
    before(() => {
      cy.useConfigProfile('mobile')
    })

    const up = (value) => `not-${value}`

    describe('password', () => {
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
        cy.loginMobile({ email, password })
        cy.visit(MOBILE_URL + '/Settings')
      })

      it('modifies the password', () => {
        const newPassword = up(password)
        cy.get(
          '*[data-testid="settings-password"] input[data-testid="password"]'
        )
          .clear()
          .type(newPassword)
        cy.get(
          '*[data-testid="settings-password"] div[data-testid="submit-form"]'
        ).click()

        cy.get(
          '*[data-testid="settings-password"] div[data-testid="submit-form"]'
        ).should('have.attr', 'data-focusable', 'true')
        cy.get('*[data-testid="logout"]').click()

        cy.visit(MOBILE_URL + '/Home')
        cy.loginMobile({ email, password })
        cy.get('[data-testid="session-token"]').should('not.exist')
        cy.loginMobile({ email, password: newPassword })
        cy.get('[data-testid="session-token"]').should('not.be.empty')
      })
    })

    describe('profile', () => {
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
        cy.loginMobile({ email, password })
        cy.visit(MOBILE_URL + '/Settings')
      })

      it('modifies an unprotected trait', () => {
        cy.get(
          '*[data-testid="settings-profile"] input[data-testid="traits.website"]'
        )
          .clear()
          .type('https://github.com/ory')
        cy.get(
          '*[data-testid="settings-profile"] div[data-testid="submit-form"]'
        ).click()
        cy.get(
          '*[data-testid="settings-profile"] div[data-testid="submit-form"]'
        ).should('have.attr', 'data-focusable', 'true')

        cy.visit(MOBILE_URL + '/Home')
        cy.get('[data-testid="session-content"]').should(
          'contain',
          'https://github.com/ory'
        )
      })

      it('modifies a protected trait', () => {
        const newEmail = up(email)
        cy.get(
          '*[data-testid="settings-profile"] input[data-testid="traits.email"]'
        )
          .clear()
          .type(newEmail)
        cy.get(
          '*[data-testid="settings-profile"] div[data-testid="submit-form"]'
        ).click()
        cy.get(
          '*[data-testid="settings-profile"] div[data-testid="submit-form"]'
        ).should('have.attr', 'data-focusable', 'true')

        cy.visit(MOBILE_URL + '/Home')
        cy.get('[data-testid="session-content"]').should('contain', newEmail)
      })
    })
  })
})
