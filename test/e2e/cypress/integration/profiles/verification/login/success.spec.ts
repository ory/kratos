import { APP_URL, appPrefix, gen } from '../../../../helpers'
import { routes as react } from '../../../../helpers/react'
import { routes as express } from '../../../../helpers/express'

context('Account Verification Login Success', () => {
  ;[
    {
      login: react.login,
      app: 'react' as 'react',
      profile: 'verification'
    },
    {
      login: express.login,
      app: 'express' as 'express',
      profile: 'verification'
    }
  ].forEach(({ profile, login, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        cy.enableLoginForVerifiedAddressOnly()
        cy.proxy(app)
      })

      it('is able to login after successful email verification', () => {
        cy.deleteMail()

        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.performEmailVerification({ expect: { email: identity.email } })

        cy.visit(login)

        cy.get(appPrefix(app) + 'input[name="password_identifier"]').type(
          identity.email
        )
        cy.get('input[name="password"]').type(identity.password)
        cy.get('button[value="password"]').click()

        cy.getSession()
      })
    })
  })
})
