import {APP_URL, assertVerifiableAddress, gen} from '../../../../helpers'


context('Verification Profile', () => {
  describe('Registration', () => {
    before(() => {
      cy.useConfigProfile('verification')
    })

    describe('successful flow', () => {
      beforeEach(() => {
        cy.longVerificationLifespan()
        cy.visit(APP_URL + '/auth/registration')
        cy.deleteMail()
      })

      afterEach(() => {
        cy.deleteMail()
    })

    const up = (value) => `up-${value}`
    const { email, password } = gen.identity()
    it('is able to verify the email address after sign up', () => {
      cy.register({ email, password })
      cy.login({ email, password })
      cy.session().then(assertVerifiableAddress({ isVerified: false, email }))

      cy.verifyEmail({ expect: { email } })
    })

    xit('sends the warning email on double sign up', () => {
      // FIXME https://github.com/ory/kratos/issues/133
      cy.clearCookies()
      cy.register({ email, password: up(password) })

      cy.verifyEmail({ expect: { email } })
    })

      it('is redirected to after_verification_return_to after verification', () => {
        cy.clearCookies()
        const {email, password} = gen.identity()
        cy.register({
          email,
          password,
          query: {after_verification_return_to: "http://127.0.0.1:4455/verification_callback"}
        })
        cy.login({email, password})
        cy.verifyEmail({expect: {email, redirectTo: "http://127.0.0.1:4455/verification_callback"}})
      })
    })
  })
})
