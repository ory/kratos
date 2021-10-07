import { appPrefix, gen } from '../../../../helpers'
import { routes as react } from '../../../../helpers/react'
import { routes as express } from '../../../../helpers/express'

context('Account Verification Login Errors', () => {
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

      it('is unable to login as long as the email is not verified', () => {
        cy.deleteMail()

        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.visit(login)

        cy.get(appPrefix(app) + '[name="password_identifier"]').type(
          identity.email
        )
        cy.get('[name="password"]').type(identity.password)
        cy.get('[value="password"]').click()

        cy.get('[data-testid="ui/message/4000010"]').contains(
          'Account not active yet'
        )

        cy.noSession()
      })
    })
  })
})
