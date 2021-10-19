import { appPrefix, gen, website } from '../../../helpers'
import { routes as react } from '../../../helpers/react'
import { routes as express } from '../../../helpers/express'

context('2FA WebAuthn', () => {
  before(() => {
    cy.task('resetCRI')
  })
  ;[
    {
      login: express.login,
      settings: express.settings,
      base: express.base,
      app: 'express' as 'express',
      profile: 'mfa'
    },
    {
      login: react.login,
      settings: react.settings,
      base: react.base,
      app: 'react' as 'react',
      profile: 'spa'
    }
  ].forEach(({ settings, login, profile, app, base }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      let email = gen.email()
      let password = gen.password()

      beforeEach(() => {
        cy.clearAllCookies()

        email = gen.email()
        password = gen.password()
        cy.registerApi({
          email,
          password,
          fields: { 'traits.website': website }
        })

        cy.login({ email, password })

        cy.longPrivilegedSessionTime()
        cy.task('sendCRI', {
          query: 'WebAuthn.disable',
          opts: {}
        })
      })

      it('should be able to identify if the authenticator is wrong', () => {
        cy.visit(settings)

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
            cy.get(
              appPrefix(app) + '[name="webauthn_register_displayname"]'
            ).type('key1')

            cy.clickWebAuthButton('register')

            cy.get('*[name="webauthn_remove"]').should('have.length', 1)

            cy.task('sendCRI', {
              query: 'WebAuthn.removeVirtualAuthenticator',
              opts: addResult
            }).then(() => {
              cy.visit(login + '?aal=aal2&refresh=true')
              cy.location().should((loc) => {
                expect(loc.href).to.include('/login')
              })
              cy.clickWebAuthButton('login')
              cy.location().should((loc) => {
                expect(loc.href).to.include('/login')
              })
              cy.getSession({
                expectAal: 'aal2',
                expectMethods: ['password', 'webauthn']
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
                cy.visit(login + '?aal=aal2&refresh=true')
                cy.location().should((loc) => {
                  expect(loc.href).to.include('/login')
                })
                cy.clickWebAuthButton('login')

                cy.location().should((loc) => {
                  expect(loc.href).to.include('/login')
                })

                cy.getSession({
                  expectAal: 'aal2',
                  expectMethods: ['password', 'webauthn']
                })
              })
            })
          })
        })
      })

      it('should be able to link multiple authenticators', () => {
        cy.visit(settings)

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
            cy.clickWebAuthButton('register')

            cy.get('*[name="webauthn_register_displayname"]').type('key2')
            cy.clickWebAuthButton('register')

            cy.get('*[name="webauthn_remove"]').should('have.length', 2)

            cy.visit(login + '?aal=aal2&refresh=true')
            cy.location().should((loc) => {
              expect(loc.href).to.include('/login')
            })
            cy.get('*[name="webauthn_login_trigger"]').should('have.length', 1)
            cy.clickWebAuthButton('login')
          })
        })
      })

      it('should be not be able to link provider if webauth is not enabled', () => {
        cy.visit(settings)
        cy.get('*[name="webauthn_register_displayname"]').type('my-key')
        cy.clickWebAuthButton('register')
        cy.get('*[name="webauthn_remove"]').should('not.exist')
      })

      it('should be able to link a webauthn provider', () => {
        cy.visit(settings)

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

            cy.clickWebAuthButton('register')

            cy.get('[data-testid="ui/message/4000002"]').should(
              'contain.text',
              'Property webauthn_register_displayname is missing.'
            )

            // Setting up with key works
            cy.get('*[name="webauthn_register_displayname"]').type('my-key')

            // We need a workaround here. So first we click, then we submit
            cy.clickWebAuthButton('register')

            cy.expectSettingsSaved()
            cy.get('*[name="webauthn_remove"]').should('exist')

            cy.visit(login + '?aal=aal2&refresh=true')
            cy.location().should((loc) => {
              expect(loc.href).to.include('/login')
            })

            cy.get('*[name="webauthn_login_trigger"]').should('have.length', 1)
            cy.clickWebAuthButton('login')
            cy.location().should((loc) => {
              expect(loc.href).to.not.include('/login')
            })

            cy.getSession({
              expectAal: 'aal2',
              expectMethods: ['password', 'webauthn', 'webauthn']
            })
            cy.visit(settings)
            cy.get('*[name="webauthn_remove"]').click()
            cy.get('*[name="webauthn_remove"]').should('not.exist')

            cy.visit(login + '?aal=aal2&refresh=true')
            cy.location().should((loc) => {
              expect(loc.href).to.include('/login')
            })

            cy.get('button[name="webauthn_login_trigger"]').should('not.exist')
            cy.get('[data-testid="ui/message/1010003"]').should('exist')
          })
        })
      })
    })
  })
})
