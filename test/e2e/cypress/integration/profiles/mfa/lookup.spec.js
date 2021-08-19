import { APP_URL, gen, website } from '../../../helpers'

context('MFA Profile', () => {
  describe('Test Lookup Secrets', () => {
    before(() => {
      cy.useConfigProfile('mfa')
    })

    let email = gen.email()
    let password = gen.password()

    beforeEach(() => {
      cy.clearCookies()
      email = gen.email()
      password = gen.password()
      cy.registerApi({ email, password, fields: { 'traits.website': website } })
      cy.login({ email, password })
      cy.longPrivilegedSessionTime()
    })

    it('should go through several lookup secret lifecycles', () => {
      cy.visit(APP_URL + '/settings')

      cy.get('p[data-testid="text-lookup_secret_codes-label"]').should(
        'not.exist'
      )
      cy.get('p[data-testid="text-lookup_secret_codes-content"]').should(
        'not.exist'
      )
      cy.get('button[name="lookup_secret_confirm"]').should('not.exist')
      cy.get('button[name="lookup_secret_regenerate"]').click()
      cy.get('p[data-testid="text-lookup_secret_codes-label"]').should(
        'contain.text',
        'These are your back up recovery codes.'
      )
      cy.get('p[data-testid="text-lookup_secret_codes-content"]').should(
        'not.be.empty'
      )

      let codes
      let codesText
      cy.get('p[data-testid="text-lookup_secret_codes-content"]').then(($e) => {
        codesText = $e.text()
        codes = codesText.trim().split(', ')
      })

      cy.get('button[name="lookup_secret_confirm"]').click()
      cy.get('form .messages .message').should(
        'contain.text',
        'Your changes have been saved!'
      )

      cy.get('button[name="lookup_secret_reveal"]').should('exist')
      cy.get('p[data-testid="text-lookup_secret_codes-content"]').should(
        'not.exist'
      )
      cy.get('button[name="lookup_secret_confirm"]').should('not.exist')
      cy.get('button[name="lookup_secret_regenerate"]').should('not.exist')

      cy.get('button[name="lookup_secret_reveal"]').click()
      cy.get('p[data-testid="text-lookup_secret_codes-content"]').should(
        ($e) => {
          expect($e.text()).to.equal(codesText)
        }
      )

      // Try to log in with a recovery code now
      cy.visit(APP_URL + '/auth/login?aal=aal2&refresh=true')
      cy.location().should((loc) => {
        expect(loc.href).to.include('/auth/login')
      })

      cy.get('*[name="method"][value="lookup_secret"]').should('exist')
      cy.get('*[name="method"][value="password"]').should('not.exist')

      // Type an invalid code
      cy.get('input[name="lookup_secret"]').should('exist')
      cy.get('input[name="lookup_secret"]').type('invalid-code')
      cy.get('*[name="method"][value="lookup_secret"]').click()
      cy.get('form .messages .message').should(
        'contain.text',
        'The backup recovery code is not valid.'
      )

      // Type a valid code
      cy.get('input[name="lookup_secret"]').should('exist')
      cy.get('input[name="lookup_secret"]').should('have.value', '')
      cy.get('input[name="lookup_secret"]').then(($e) => {
        cy.wrap($e).type(codes[0])
      })
      cy.get('*[name="method"][value="lookup_secret"]').click()

      let authenticatedAt
      cy.session({
        expectAal: 'aal2',
        expectMethods: ['password', 'lookup_secret', 'lookup_secret']
      }).then((session) => {
        authenticatedAt = session.authenticated_at
        expect(session.authenticator_assurance_level).to.equal('aal2')
      })

      // Retry auth with the used code
      cy.visit(APP_URL + '/auth/login?aal=aal2&refresh=true')
      cy.location().should((loc) => {
        expect(loc.href).to.include('/auth/login')
      })
      cy.get('input[name="lookup_secret"]').then(($e) => {
        cy.wrap($e).type(codes[0])
      })
      cy.get('*[name="method"][value="lookup_secret"]').click()

      // Use a valid code
      cy.get('form .messages .message').should(
        'contain.text',
        'This backup recovery code has already been used.'
      )
      cy.get('input[name="lookup_secret"]').then(($e) => {
        cy.wrap($e).type(codes[1])
      })
      cy.get('*[name="method"][value="lookup_secret"]').click()

      cy.session({
        expectAal: 'aal2',
        expectMethods: [
          'password',
          'lookup_secret',
          'lookup_secret',
          'lookup_secret'
        ]
      }).then((session) => {
        expect(session.authenticatedAt).to.not.equal(authenticatedAt)
      })

      // Going back to the settings UI we should see that the codes have been "used"
      cy.visit(APP_URL + '/settings')
      cy.get('button[name="lookup_secret_reveal"]').click()
      cy.get('p[data-testid="text-lookup_secret_codes-content"]').should(
        ($e) => {
          let newCodes = codes
          newCodes[0] = 'used'
          newCodes[1] = 'used'
          expect($e.text()).to.contain(newCodes.join(', '))
        }
      )

      // Regenerating the codes means the old one become invalid
      cy.get('*[name=lookup_secret_regenerate]').click()
      let regenCodes
      let regenCodesText
      cy.get('p[data-testid="text-lookup_secret_codes-content"]').then(($e) => {
        regenCodesText = $e.text()
        regenCodes = regenCodesText.trim().split(', ')
      })

      // Confirm it
      cy.get('*[name=lookup_secret_confirm]').click()
      cy.get('*[name="lookup_secret_reveal"]').click()
      cy.get('p[data-testid="text-lookup_secret_codes-content"]').should(
        ($e) => {
          expect($e.text()).to.equal(regenCodesText)
        }
      )

      // Log in and see if we can use the old / new keys
      cy.visit(APP_URL + '/auth/login?aal=aal2&refresh=true')
      cy.location().should((loc) => {
        expect(loc.href).to.include('/auth/login')
      })

      // Using an old code fails
      cy.get('input[name="lookup_secret"]').then(($e) => {
        cy.wrap($e).type(codes[3])
      })
      cy.get('*[name="method"][value="lookup_secret"]').click()
      cy.get('form .messages .message').should(
        'contain.text',
        'The backup recovery code is not valid.'
      )

      // Using a new code succeeds
      cy.get('input[name="lookup_secret"]').then(($e) => {
        cy.wrap($e).type(regenCodes[0])
      })
      cy.get('*[name="method"][value="lookup_secret"]').click()

      // Going back to the settings UI we should see that the codes have been "used"
      cy.visit(APP_URL + '/settings')
      cy.get('button[name="lookup_secret_reveal"]').click()
      cy.get('p[data-testid="text-lookup_secret_codes-content"]').should(
        ($e) => {
          let newCodes = regenCodes
          newCodes[0] = 'used'
          expect($e.text()).to.contain(newCodes.join(', '))
        }
      )
    })

    it('should end up at login screen if trying to reveal without privileged session', () => {
      cy.shortPrivilegedSessionTime()
      cy.visit(APP_URL + '/settings')
      cy.get('button[name="lookup_secret_regenerate"]').click()
      cy.reauth({
        expect: { email },
        type: { email: email, password: password }
      })

      let codes
      cy.get('p[data-testid="text-lookup_secret_codes-content"]').then(($e) => {
        codes = $e.text().trim().split(', ')
      })

      cy.shortPrivilegedSessionTime()
      cy.get('button[name="lookup_secret_confirm"]').click()
      cy.reauth({
        expect: { email },
        type: { email: email, password: password }
      })

      cy.shortPrivilegedSessionTime()
      cy.get('button[name="lookup_secret_reveal"]').click()
      cy.reauth({
        expect: { email },
        type: { email: email, password: password }
      })
    })

    it('should not show lookup as an option if not configured', () => {
      cy.visit(APP_URL + '/auth/login?aal=aal2')
      cy.location().should((loc) => {
        expect(loc.href).to.include('/auth/login')
      })

      cy.get('*[name="method"][value="totp"]').should('not.exist')
      cy.get('*[name="method"][value="lookup_secret"]').should('not.exist')
      cy.get('*[name="method"][value="password"]').should('not.exist')
      cy.get('form .messages .message').should(
        'contain.text',
        'Please complete the second authentication challenge.'
      )
    })
  })
})
