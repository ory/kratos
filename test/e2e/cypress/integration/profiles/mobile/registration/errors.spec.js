import {gen, MOBILE_URL, website} from '../../../../helpers'

context('Mobile Profile', () => {
  describe('Registration Flow Errors', () => {
    before(() => {
      cy.useConfigProfile('mobile')
    })

    beforeEach(() => {
      cy.visit(MOBILE_URL + "/Registration")
    })

    const email = gen.email()
    const password = gen.password()

    describe('show errors when invalid signup data is used', () => {
      it('should show an error when the password has leaked before', () => {
      cy.get('input[data-testid="traits.email"]').type(email)
      cy.get('input[data-testid="password"]').type('123456')
      cy.get('input[data-testid="traits.website"]').type(website)
      cy.get('div[data-testid="submit-form"]').click()

      cy.get('*[data-testid="field/password"]')
        .should('contain.text', 'data breaches')
    })

    it('should show an error when the password is too similar', () => {
      cy.get('input[data-testid="traits.email"]').type(email)
      cy.get('input[data-testid="password"]').type(email)
      cy.get('input[data-testid="traits.website"]').type(website)
      cy.get('div[data-testid="submit-form"]').click()

      cy.get('*[data-testid="field/password"]')
        .should('contain.text', 'too similar')
    })

    it('should show an error when the password is empty', () => {
      cy.get('input[data-testid="traits.website"]').type(website)
      cy.get('input[data-testid="traits.email"]').type(email)

      cy.get('div[data-testid="submit-form"]').click()
      cy.get('*[data-testid="field/password"]')
        .should('contain.text', 'length must be')
    })

    it('should show an error when the email is empty', () => {
      cy.get('input[data-testid="traits.website"]').type('https://www.ory.sh')
      cy.get('input[data-testid="password"]').type(password)

      cy.get('div[data-testid="submit-form"]').click()
      cy.get('*[data-testid="field/traits.email"]')
        .should('contain.text', 'valid "email"')
    })

    it('should show an error when the email is not an email', () => {
      cy.get('input[data-testid="traits.website"]').type('https://www.ory.sh')
      cy.get('input[data-testid="traits.email"]').type('not-an-email')
      cy.get('input[data-testid="password"]').type(password)

      cy.get('div[data-testid="submit-form"]').click()
      cy.get('*[data-testid="field/traits.email"]')
        .should('contain.text', 'valid "email"')
    })

    it('should show a missing indicator if no fields are set', () => {
      cy.get('div[data-testid="submit-form"]').click()
      cy.get('*[data-testid="field/password"]').should('contain.text', 'length must be')
    })


    it('should show an error when the website is too short', () => {
      cy.get('input[data-testid="traits.website"]').type('http://s')
      cy.get('input[data-testid="traits.email"]').type(email)

      // fixme https://github.com/ory/kratos/issues/368
      cy.get('input[data-testid="password"]').type(password)

      cy.get('div[data-testid="submit-form"]').click()
      cy.get('*[data-testid="field/traits.website"]').should(
        'contain.text',
        'length must be >= 10'
      )
    })
    })
  })
})
