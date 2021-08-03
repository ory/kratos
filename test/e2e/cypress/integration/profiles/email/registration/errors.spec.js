import { APP_URL, gen } from '../../../../helpers'

context('Email Profile', () => {
  describe('Registration Flow Errors', () => {
    before(() => {
      cy.useConfigProfile('email')
    })

    beforeEach(() => {
      cy.visit(APP_URL + '/auth/registration')
    })

    const identity = gen.email()
    const password = gen.password()

    it('fails when CSRF cookies are missing', () => {
      cy.clearCookies()

      cy.get('input[name="traits.website"]').type('https://www.ory.sh')
      cy.get('input[name="traits.email"]')
        .type(identity)
        .should('have.value', identity)
      cy.get('input[name="password"]')
        .type('123456')
        .should('have.value', '123456')

      let initial
      cy.location().should((location) => {
        initial = location.search
      })
      cy.get('button[type="submit"]').click()

      // We end up at a new flow
      cy.location('search').should('not.eq', initial)
    })

    describe('show errors when invalid signup data is used', () => {
      it('should show an error when the password has leaked before', () => {
        cy.get('input[name="traits.website"]').type('https://www.ory.sh')
        cy.get('input[name="traits.email"]')
          .type(identity)
          .should('have.value', identity)
        cy.get('input[name="password"]')
          .type('123456')
          .should('have.value', '123456')

        cy.get('button[type="submit"]').click()
        cy.get('.messages .message').should('contain.text', 'data breaches')
      })

      it('should show an error when the password is too similar', () => {
        cy.get('input[name="traits.website"]').type('https://www.ory.sh')
        cy.get('input[name="traits.email"]').type(identity)
        cy.get('input[name="password"]').type(identity)

        cy.get('button[type="submit"]').click()
        cy.get('.messages .message').should('contain.text', 'too similar')
      })

      it('should show an error when the password is empty', () => {
        cy.get('input[name="traits.website"]').type('https://www.ory.sh')
        cy.get('input[name="traits.email"]').type(identity)

        cy.get('button[type="submit"]').click()
        cy.get('.messages .message').should('contain.text', 'length must be')
      })

      it('should show an error when the email is empty', () => {
        cy.get('input[name="traits.website"]').type('https://www.ory.sh')
        cy.get('input[name="password"]').type(password)

        cy.get('button[type="submit"]').click()
        cy.get('.messages .message').should('contain.text', 'valid "email"')
      })

      it('should show an error when the email is not an email', () => {
        cy.get('input[name="traits.website"]').type('https://www.ory.sh')
        cy.get('input[name="password"]').type('not-an-email')

        cy.get('button[type="submit"]').click()
        cy.get('.messages .message').should('contain.text', 'valid "email"')
      })

      it('should show a missing indicator if no fields are set', () => {
        cy.get('button[type="submit"]').click()
        cy.get('.messages .message').should('contain.text', 'but got 0')
      })

      it('should show an error when the website is not a valid URI', () => {
        cy.get('input[name="traits.website"]')
          .type('1234')
          .then(($input) => {
            expect($input[0].validationMessage).to.contain('URL')
          })
      })

      it('should show an error when the website is too short', () => {
        cy.get('input[name="traits.website"]').type('http://s')

        // fixme https://github.com/ory/kratos/issues/368
        cy.get('input[name="password"]').type(password)

        cy.get('button[type="submit"]').click()
        cy.get('.messages .message').should(
          'contain.text',
          'length must be >= 10'
        )
      })
    })
  })
})
