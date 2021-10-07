import {gen, parseHtml} from '../../../../helpers'
import {routes as react} from "../../../../helpers/react";
import {routes as express} from "../../../../helpers/express";

const fields = {'traits.website': 'https://www.ory.sh'}

context('Account Recovery Errors', () => {
  [
    {
      recovery: react.recovery,
      app: 'react', profile: 'spa'
    },
    {
      recovery: express.recovery,
      app: 'express', profile: 'recovery'
    }
  ].forEach(({recovery, profile, app,}) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
      })

      beforeEach(() => {
        cy.deleteMail()
        cy.longRecoveryLifespan()
        cy.longLinkLifespan()
        cy.disableVerification()
        cy.enableRecovery()
      })

      it('should receive a stub email when recovering a non-existent account', () => {
        cy.visit(recovery)

        const email = gen.email()
        cy.get('input[name="email"]').type(email)
        cy.get('button[value="link"]').click()

        cy.location('pathname').should('eq', '/recovery')
        cy.get('[data-testid="ui/message/1060002"]').should(
          'have.text',
          'An email containing a recovery link has been sent to the email address you provided.'
        )
        cy.get('input[name="email"]').should('have.value', email)

        cy.getMail().should((message) => {
          expect(message.subject.trim()).to.equal('Account access attempted')
          expect(message.fromAddress.trim()).to.equal('no-reply@ory.kratos.sh')
          expect(message.toAddresses).to.have.length(1)
          expect(message.toAddresses[0].trim()).to.equal(email)

          const link = parseHtml(message.body).querySelector('a')
          expect(link).to.be.null
        })
      })

      it('should cause form errors', () => {
        cy.visit(recovery)

        cy.get('button[value="link"]').click()
        cy.get('[data-testid="ui/message/4000002"]').should(
          'contain.text',
          'Property email is missing.'
        )
        cy.get('[name="method"][value="link"]').should(
          'exist'
        )
      })

      it('is unable to recover the email address if the code is expired', () => {
        cy.shortLinkLifespan()
        const identity = gen.identity()
        cy.registerApi({...identity, fields})
        cy.recoverApi({email: identity.email})
        cy.recoverEmailButExpired({expect: {email: identity.email}})

        cy.get('[data-testid="ui/message/4060005"]').should(
          'contain.text',
          'The recovery flow expired'
        )

        cy.noSession()
      })

      it('is unable to recover the account if the code is incorrect', () => {
        const identity = gen.identity()
        cy.registerApi({...identity, fields})
        cy.recoverApi({email: identity.email})

        cy.getMail().then((mail) => {
          console.log(mail)
          const link = parseHtml(mail.body).querySelector('a')
          cy.visit(link.href + '-not') // add random stuff to the confirm challenge
          cy.get('[data-testid="ui/message/4060004"]').should(
            'have.text',
            'The recovery token is invalid or has already been used. Please retry the flow.'
          )
          cy.noSession()
        })
      })

      it('is unable to recover the account using the token twice', () => {
        const identity = gen.identity()
        cy.registerApi({...identity, fields})
        cy.recoverApi({email: identity.email})

        cy.getMail().then((mail) => {
          const link = parseHtml(mail.body).querySelector('a')

          cy.visit(link.href) // add random stuff to the confirm challenge
          cy.getSession()
          cy.logout()

          cy.visit(link.href)
          cy.get('[data-testid="ui/message/4060004"]').should(
            'have.text',
            'The recovery token is invalid or has already been used. Please retry the flow.'
          )
          cy.noSession()
        })
      })
    })
  })
})
