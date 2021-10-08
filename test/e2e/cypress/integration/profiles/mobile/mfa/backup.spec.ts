import { gen, MOBILE_URL, website } from '../../../../helpers'

context('Mobile Profile', () => {
  describe('TOTP 2FA Flow', () => {
    before(() => {
      cy.useConfigProfile('mobile')
    })

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

      it('should be able to lifecycle through lookup_secret flows', () => {
        cy.get('*[data-testid="field/lookup_secret_codes"]').should('not.exist')
        cy.get('*[data-testid="field/lookup_secret_confirm/true"]').should(
          'not.exist'
        )
        cy.get('*[data-testid="field/lookup_secret_reveal/true"]').should(
          'not.exist'
        )
        cy.get('*[data-testid="field/lookup_secret_regenerate/true"]').click()
        cy.get('*[data-testid="field/lookup_secret_reveal/true"]').should(
          'not.exist'
        )
        cy.get('*[data-testid="field/lookup_secret_codes"]').should('exist')
        let codes
        cy.get('*[data-testid="field/lookup_secret_codes/text"]').then(($e) => {
          codes = $e.text().trim().split(', ')
        })
        cy.get('*[data-testid="field/lookup_secret_confirm/true"]').click()
        cy.get('*[data-testid="form-messages"]').should(
          'contain.text',
          'Your changes have been saved!'
        )

        cy.get('*[data-testid="field/lookup_secret_confirm/true"]').should(
          'not.exist'
        )
        cy.get('*[data-testid="field/lookup_secret_regenerate/true"]').should(
          'not.exist'
        )
        cy.get('*[data-testid="field/lookup_secret_codes/true"]').should(
          'not.exist'
        )

        cy.get('*[data-testid="field/lookup_secret_reveal/true"]').click()
        cy.get('*[data-testid="field/lookup_secret_regenerate/true"]').should(
          'exist'
        )
        cy.get('*[data-testid="field/lookup_secret_codes/text"]').then(($e) => {
          const actualCodes = $e.text().trim().split(', ')
          expect(actualCodes.join(', ')).to.eq(codes.join(', '))
        })

        let newCodes
        cy.get('*[data-testid="field/lookup_secret_regenerate/true"]').click()
        cy.get(
          '*[data-testid="field/lookup_secret_regenerate/true"]:disabled'
        ).should('not.exist')
        cy.get('*[data-testid="field/lookup_secret_codes/text"]').then(($e) => {
          newCodes = $e.text().trim().split(', ')
        })
        cy.get('*[data-testid="field/lookup_secret_confirm/true"]').click()
        cy.get('*[data-testid="field/lookup_secret_reveal/true"]').click()
        cy.get('*[data-testid="field/lookup_secret_codes/text"]').then(($e) => {
          const actualCodes = $e.text().trim().split(', ')
          expect(actualCodes.join(', ')).to.eq(newCodes.join(', '))
        })

        cy.visit(MOBILE_URL + '/Login?aal=aal2&refresh=true')

        // First use a wrong code
        cy.get('[data-testid=lookup_secret]').then(($e) => {
          console.log(codes)
          cy.wrap($e).type('1234')
        })
        cy.get('*[data-testid="field/method/lookup_secret"]').click()
        cy.get('*[data-testid="form-messages"]').should(
          'contain.text',
          'The backup recovery code is not valid.'
        )
        cy.get('[data-testid=lookup_secret]').then(($e) => {
          cy.wrap($e).type(newCodes[0])
        })
        cy.get('*[data-testid="field/method/lookup_secret"]').click()
        cy.get('[data-testid="session-content"]').should('contain', 'aal2')
        cy.get('[data-testid="session-content"]').should(
          'contain',
          'lookup_secret'
        )
      })
    })
  })
})
