import { APP_URL, assertAddress, gen } from '../../../../helpers'

context('Verify', () => {
  describe('successful flow', () => {
    let identity

    before(() => {
      cy.deleteMail()
    })

    beforeEach(() => {
      identity = gen.identity()
      cy.register(identity)
      cy.deleteMail({ atLeast: 1 }) // clean up registration email

      cy.login(identity)
      cy.visit(APP_URL + '/verify')
    })

    it('should request verification and receive an email and verify it', () => {
      cy.get('input[name="to_verify"]').type(identity.email)
      cy.get('button[type="submit"]').click()

      cy.location('pathname').should('eq', '/')

      cy.verifyEmail({ expect: { email: identity.email } })

      cy.location('pathname').should('eq', '/')
    })

    it('should request verification for an email that does not exist yet', () => {
      const email = `not-${identity.email}`
      cy.get('input[name="to_verify"]').type(email)
      cy.get('button[type="submit"]').click()

      cy.getMail().should((message) => {
        expect(message.subject.trim()).to.equal(
          'Someone tried to verify this email address'
        )
        expect(message.fromAddress.trim()).to.equal('no-reply@ory.kratos.sh')
        expect(message.toAddresses).to.have.length(1)
        expect(message.toAddresses[0].trim()).to.equal(email)
      })

      cy.session().then(
        assertAddress({ isVerified: false, email: identity.email })
      )

      cy.location('pathname').should('eq', '/')
    })
  })
})
