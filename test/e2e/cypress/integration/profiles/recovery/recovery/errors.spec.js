import { APP_URL, gen, parseHtml } from '../../../../helpers'

context('Recovery', () => {
  describe('error flow', () => {
    let identity

    before(() => {
      cy.deleteMail()
    })

    beforeEach(() => {
      cy.longRecoveryLifespan()
      cy.visit(APP_URL + '/recovery')
    })

    it('should receive a stub email when recovering a non-existent account', () => {
      const email = gen.email()
      cy.get('input[name="email"]').type(email)
      cy.get('button[value="link"]').click()

      cy.location('pathname').should('eq', '/recovery')
      cy.get('.messages .message.info').should(
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
      cy.get('button[value="link"]').click()
      cy.get('.messages .message').should(
        'contain.text',
        'Property email is missing'
      )
    })

    it('is unable to recover the email address if the code is expired', () => {
      cy.shortRecoveryLifespan()
      identity = gen.identity()
      cy.register(identity)
      cy.visit(APP_URL + '/recovery')

      cy.get('input[name="email"]').type(identity.email)
      cy.get('button[value="link"]').click()

      cy.wait(4000)

      cy.recoverEmailButExpired({ expect: { email: identity.email } })

      cy.get('.messages .message.error').should(
        'contain.text',
        'The recovery flow expired'
      )

      cy.noSession()
    })

    it('is unable to recover the account if the code is incorrect', () => {
      identity = gen.identity()
      cy.register(identity)
      cy.visit(APP_URL + '/recovery')

      cy.get('input[name="email"]').type(identity.email)
      cy.get('button[value="link"]').click()

      cy.getMail().then((mail) => {
        const link = parseHtml(mail.body).querySelector('a')
        cy.visit(link.href + '-not') // add random stuff to the confirm challenge
        cy.get('.messages .message.error').should(
          'have.text',
          'The recovery token is invalid or has already been used. Please retry the flow.'
        )
        cy.noSession()
      })
    })

    it('is unable to recover the account using the token twice', () => {
      identity = gen.identity()
      cy.register(identity)
      cy.visit(APP_URL + '/recovery')

      cy.get('input[name="email"]').type(identity.email)
      cy.get('button[value="link"]').click()

      cy.getMail().then((mail) => {
        const link = parseHtml(mail.body).querySelector('a')

        cy.visit(link.href) // add random stuff to the confirm challenge
        cy.session()
        cy.logout()

        cy.visit(link.href)
        cy.get('.messages .message.error').should(
          'have.text',
          'The recovery token is invalid or has already been used. Please retry the flow.'
        )
        cy.noSession()
      })
    })
  })
})
