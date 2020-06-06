import { APP_URL, gen, website } from '../../../../helpers'

context('Settings', () => {
  let email

  const hydraReauthFails = () => {
    cy.clearCookies()
    cy.visit(APP_URL + '/auth/login')
    cy.get('button[value="hydra"]').click()

    cy.get('#username').type(email)
    cy.get('#remember').click()
    cy.get('#accept').click()

    cy.get('input[name="traits.website"]').clear().type(website)
    cy.get('button[value="hydra"]').click()

    cy.get('.messages .message').should(
      'contain.text',
      'An account with the same identifier'
    )

    cy.noSession()
  }

  beforeEach(() => {
    cy.clearCookies()
    email = gen.email()

    cy.registerOidc({ email, expectSession: true, website })
    cy.visit(APP_URL + '/settings')
  })

  describe('oidc', () => {
    it('should show the correct options', () => {
      cy.get('#user-oidc button[value="hydra"]').should('not.exist')

      cy.get('#user-oidc button[value="google"]')
        .should('have.attr', 'name', 'link')
        .should('contain.text', 'Link google')

      cy.get('#user-oidc button[value="github"]')
        .should('have.attr', 'name', 'link')
        .should('contain.text', 'Link github')
    })

    it('should show the unlink once password is set', () => {
      cy.get('#user-oidc button[value="hydra"]').should('not.exist')

      cy.get('#user-password input[name="password"]').type(gen.password())
      cy.get('#user-password button[type="submit"]').click()
      cy.visit(APP_URL + '/settings')

      cy.get('#user-oidc button[value="hydra"]')
        .should('have.attr', 'name', 'unlink')
        .should('contain.text', 'Unlink hydra')
    })

    it('should link google', () => {
      cy.get('#user-oidc button[value="google"]').click()

      cy.get('input[name="scope"]').each(($el) => cy.wrap($el).click())
      cy.get('#remember').click()
      cy.get('#accept').click()

      cy.visit(APP_URL + '/settings')

      cy.get('#user-oidc button[value="google"]')
        .should('have.attr', 'name', 'unlink')
        .should('contain.text', 'Unlink google')

      cy.logout()

      cy.visit(APP_URL + '/auth/login')
      cy.get('button[value="google"]').click()
      cy.session()
    })

    it('should link google after re-auth', () => {
      cy.waitForPrivilegedSessionToExpire()
      cy.get('#user-oidc button[value="google"]').click()
      cy.location('pathname').should('equal', '/auth/login')
      cy.get('button[value="hydra"]').click()

      // prompt=login means that we need to re-auth!
      cy.get('#username').type(email)
      cy.get('#accept').click()

      // we re-authed, now we do the google oauth2 dance
      cy.get('#username').type(gen.email())
      cy.get('#accept').click()
      cy.get('input[name="scope"]').each(($el) => cy.wrap($el).click())
      cy.get('#accept').click()

      cy.get('.container').should(
        'contain.text',
        'Your changes have been saved!'
      )

      cy.get('#user-oidc button[value="google"]')
        .should('have.attr', 'name', 'unlink')
        .should('contain.text', 'Unlink google')

      cy.visit(APP_URL + '/settings')

      cy.get('#user-oidc button[value="google"]')
        .should('have.attr', 'name', 'unlink')
        .should('contain.text', 'Unlink google')
    })

    it('should unlink hydra and no longer be able to sign in', () => {
      cy.get('#user-oidc button[value="hydra"]').should('not.exist')

      cy.get('#user-password input[name="password"]').type(gen.password())
      cy.get('#user-password button[type="submit"]').click()
      cy.visit(APP_URL + '/settings')

      cy.get('#user-oidc button[value="hydra"]').click()

      // It will no longer be possible to sign up with this provider because of a UNIQUE key violation
      // because of the email verification table.
      //
      // Basically what needs to happen is for the user to use the LINK feature to link this account.
      hydraReauthFails()
    })

    it('should unlink hydra after reauth', () => {
      cy.get('#user-oidc button[value="hydra"]').should('not.exist')

      cy.get('#user-password input[name="password"]').type(gen.password())
      cy.get('#user-password button[type="submit"]').click()
      cy.visit(APP_URL + '/settings')

      cy.waitForPrivilegedSessionToExpire()
      cy.get('#user-oidc button[value="hydra"]').click()
      cy.location('pathname').should('equal', '/auth/login')
      cy.get('button[value="hydra"]').click()

      // prompt=login means that we need to re-auth!
      cy.get('#username').type(email)
      cy.get('#accept').click()

      hydraReauthFails()
    })
  })
})
