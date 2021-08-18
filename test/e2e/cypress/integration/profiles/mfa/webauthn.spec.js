import { APP_URL, gen, website } from '../../../helpers'

context('MFA Profile', () => {
  describe('Test WebAuthn', () => {
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
      cy.task('sendCRI', {
        query: 'WebAuthn.disable',
        opts: {}
      })
    })

    it('should be able to identify if the authenticator is wrong', () => {
      cy.visit(APP_URL + '/settings')

      // Set up virtual authenticator
      cy.task('sendCRI', {
        query: 'WebAuthn.enable',
        opts: {}
      }).then(() => {
        cy.task('sendCRI', {
          query: 'WebAuthn.addVirtualAuthenticator',
          opts: {
            options: {
              protocol: 'ctap2',
              transport: 'usb',
              hasResidentKey: true,
              hasUserVerification: true,
              isUserVerified: true
            }
          }
        }).then((addResult) => {
          cy.get('*[name="webauthn_register_displayname"]').type('key1')
          cy.get('*[name="webauthn_register_trigger"]').click()
          cy.get('*[name="webauthn_remove"]').should('have.length', 1)

          cy.task('sendCRI', {
            query: 'WebAuthn.removeVirtualAuthenticator',
            opts: addResult
          }).then(() => {
            cy.visit(APP_URL + '/auth/login?aal=aal2')
            cy.location().should((loc) => {
              expect(loc.href).to.include('/auth/login')
            })
            cy.get('*[name="webauthn_login_trigger"]').click()
            cy.location().should((loc) => {
              expect(loc.href).to.include('/auth/login')
            })
            cy.session({
              expectAal: 'aal1',
              expectMethods: ['password']
            })

            cy.task('sendCRI', {
              query: 'WebAuthn.addVirtualAuthenticator',
              opts: {
                options: {
                  protocol: 'ctap2',
                  transport: 'usb',
                  hasResidentKey: true,
                  hasUserVerification: true,
                  isUserVerified: true
                }
              }
            }).then((addResult) => {
              cy.visit(APP_URL + '/auth/login?aal=aal2')
              cy.location().should((loc) => {
                expect(loc.href).to.include('/auth/login')
              })
              cy.get('*[name="webauthn_login_trigger"]').click()

              cy.location().should((loc) => {
                expect(loc.href).to.include('/auth/login')
              })

              cy.session({
                expectAal: 'aal1',
                expectMethods: ['password']
              })
            })
          })
        })
      })
    })

    it('should be able to link multiple authenticators', () => {
      cy.visit(APP_URL + '/settings')

      // Set up virtual authenticator
      cy.task('sendCRI', {
        query: 'WebAuthn.enable',
        opts: {}
      }).then(() => {
        cy.task('sendCRI', {
          query: 'WebAuthn.addVirtualAuthenticator',
          opts: {
            options: {
              protocol: 'ctap2',
              transport: 'usb',
              hasResidentKey: true,
              hasUserVerification: true,
              isUserVerified: true
            }
          }
        }).then((addResult) => {
          cy.get('*[name="webauthn_register_displayname"]').type('key1')
          cy.get('*[name="webauthn_register_trigger"]').click()

          cy.get('*[name="webauthn_register_displayname"]').type('key2')
          cy.get('*[name="webauthn_register_trigger"]').click()

          cy.get('*[name="webauthn_remove"]').should('have.length', 2)

          cy.visit(APP_URL + '/auth/login?aal=aal2')
          cy.location().should((loc) => {
            expect(loc.href).to.include('/auth/login')
          })
          cy.get('*[name="webauthn_login_trigger"]').should('have.length', 1)
          cy.get('*[name="webauthn_login_trigger"]').click()
        })
      })
    })

    it('should be not be able to link provider if webauth is not enabled', () => {
      cy.visit(APP_URL + '/settings')
      cy.get('*[name="webauthn_register_displayname"]').type('my-key')
      cy.get('*[name="webauthn_register_trigger"]').click()
      cy.get('*[name="webauthn_remove"]').should('not.exist')
    })

    it('should be able to link a webauthn provider', () => {
      cy.visit(APP_URL + '/settings')

      // Set up virtual authenticator
      cy.task('sendCRI', {
        query: 'WebAuthn.enable',
        opts: {}
      }).then(() => {
        cy.task('sendCRI', {
          query: 'WebAuthn.addVirtualAuthenticator',
          opts: {
            options: {
              protocol: 'ctap2',
              transport: 'usb',
              hasResidentKey: true,
              hasUserVerification: true,
              isUserVerified: true
            }
          }
        }).then((addResult) => {
          // Signing up without a display name causes an error
          cy.get('*[name="webauthn_remove"]').should('not.exist')

          cy.get('*[name="webauthn_register_trigger"]').click()

          cy.get('form .messages .message').should(
            'contain.text',
            'length must be >= 1, but got 0'
          )

          // Setting up with key works
          cy.get('*[name="webauthn_register_displayname"]').type('my-key')

          // We need a workaround here. So first we click, then we submit
          cy.get('*[name="webauthn_register_trigger"]').click()

          cy.get('form .messages .message').should(
            'contain.text',
            'Your changes have been saved!'
          )
          cy.get('*[name="webauthn_remove"]').should('exist')

          cy.visit(APP_URL + '/auth/login?aal=aal2')
          cy.location().should((loc) => {
            expect(loc.href).to.include('/auth/login')
          })

          cy.get('button[name="webauthn_login_trigger"]').click()
          cy.location().should((loc) => {
            expect(loc.href).to.not.include('/auth/login')
          })

          cy.session({
            expectAal: 'aal2',
            expectMethods: ['password', 'webauthn']
          })
          cy.visit(APP_URL + '/settings')
          cy.get('*[name="webauthn_remove"]').click()
          cy.get('*[name="webauthn_remove"]').should('not.exist')

          cy.visit(APP_URL + '/auth/login?aal=aal2&refresh=true')
          cy.location().should((loc) => {
            expect(loc.href).to.include('/auth/login')
          })

          cy.get('button[name="webauthn_login_trigger"]').should('not.exist')
        })
      })
    })
  })
})
