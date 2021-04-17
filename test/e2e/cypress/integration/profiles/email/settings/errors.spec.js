// TODO: implement wrong credentials when reauthing
// TODO: implement other account when reauthing
import {
  APP_URL,
  gen,
  password,
  privilegedLifespan,
  website,
} from '../../../../helpers'

context('Settings Flow Errors', () => {
  let email = gen.email()
  let password = gen.password()

  const emailSecond = `second-${gen.email()}`
  const passwordSecond = gen.password()

  const up = (value) => `not-${value}`
  const down = (value) => value.replace(/not-/, '')

  before(() => {
    console.log('before')
    cy.register({
      email: emailSecond,
      password: passwordSecond,
      fields: { 'traits.website': 'https://github.com/ory/kratos' },
    })
    cy.clearCookies()
    cy.register({ email, password, fields: { 'traits.website': website } })
  })

  beforeEach(() => {
    console.log('beforeAll')
    cy.clearCookies()
    cy.login({ email, password })
    cy.visit(APP_URL + '/settings')
  })
  describe('profile', () => {
    beforeEach(() => {
      cy.longPrivilegedSessionTime()
    })

    it('fails with validation errors', () => {
      cy.get('input[name="traits.website"]')
        .clear()
        .type('http://s')
      cy.get('button[value="profile"]').click()
      cy.get('.messages .message').should(
        'contain.text',
        'length must be >= 10'
      )
    })

    it('fails because reauth is another person', () => {
      cy.get('input[name="traits.email"]').clear().type(up(email))
      cy.shortPrivilegedSessionTime()
      cy.get('button[value="profile"]').click()

      cy.reauth({
        expect: { email },
        type: { email: emailSecond, password: passwordSecond },
      })

      cy.get('.messages .message').should(
        'contain.text',
        'You must restart the flow because the resumable session was initiated by another person.'
      )

      // Try to log in with updated credentials -> should fail
      cy.clearCookies()
      cy.login({ email: up(email), password, expectSession: false })
    })

    it('does not update data because resumable session was removed', () => {
      cy.get('input[name="traits.email"]').clear().type(up(email))
      cy.shortPrivilegedSessionTime()
      cy.get('button[value="profile"]').click()

      cy.clearCookies()
      cy.login({ email, password })

      cy.session().should((session) => {
        const { identity } = session
        expect(identity.traits.email).to.equal(email)
      })
    })

    it('does not update without re-auth', () => {
      cy.get('input[name="traits.email"]').clear().type(up(email))
      cy.shortPrivilegedSessionTime() // wait for the privileged session to time out
      cy.get('button[value="profile"]').click()

      cy.visit(APP_URL + '/')

      cy.session().should((session) => {
        const { identity } = session
        expect(identity.traits.email).to.equal(email)
      })
    })

    it('does not resume another failed request', () => {
      // checks here that we're checking settingsRequest.id == cookie.stored.id
      cy.get('input[name="traits.email"]').clear().type(up(email))
      cy.shortPrivilegedSessionTime() // wait for the privileged session to time out
      cy.get('button[value="profile"]').click()

      cy.visit(APP_URL + '/settings')
      cy.get('input[name="traits.website"]')
        .clear()
        .type('http://github.com/aeneasr')
      cy.get('button[value="profile"]').click()

      cy.session().should((session) => {
        const { identity } = session
        expect(identity.traits.email).to.equal(email) // this is NOT up(email)
        expect(identity.traits.website).to.equal('http://github.com/aeneasr') // this is NOT up(email)
      })
    })
  })

  describe('password', () => {
    beforeEach(() => {
      cy.longPrivilegedSessionTime()
    })
    afterEach(() => {
      cy.longPrivilegedSessionTime()
    })

    it('fails if password policy is violated', () => {
      cy.get('input[name="password"]').clear().type('123456')
      cy.get('button[value="password"]').click()
      cy.get('.messages .message').should(
        'contain.text',
        'data breaches'
      )
    })

    it('fails because reauth is another person', () => {
      cy.get('input[name="password"]').clear().type(up(password))
      cy.shortPrivilegedSessionTime() // wait for the privileged session to time out
      cy.get('button[value="password"]').click()

      cy.reauth({
        expect: { email },
        type: { email: emailSecond, password: passwordSecond },
      })

      cy.get('.messages .message').should(
        'contain.text',
        'You must restart the flow because the resumable session was initiated by another person.'
      )

      // Try to log in with updated credentials -> should fail
      cy.clearCookies()
      cy.login({ email, password: up(password), expectSession: false })
    })

    it('does not update without re-auth', () => {
      cy.get('input[name="password"]').clear().type(up(password))
      cy.shortPrivilegedSessionTime() // wait for the privileged session to time out
      cy.get('button[value="password"]').click()

      cy.visit(APP_URL + '/')
      cy.clearCookies()
      cy.login({ email, password: up(password), expectSession: false })
    })

    it('does not update data because resumable session was removed', () => {
      cy.get('input[name="password"]').clear().type(up(password))
      cy.shortPrivilegedSessionTime() // wait for the privileged session to time out
      cy.get('button[value="password"]').click()

      cy.clearCookies()
      cy.login({ email, password })
      cy.clearCookies()
      cy.login({ email, password: up(password), expectSession: false })
    })

    it('does not resume another queued request', () => {
      // checks here that we're checking settingsRequest.id == cookie.stored.id
      cy.get('input[name="password"]')
        .clear()
        .type(up(up(password)))
      cy.shortPrivilegedSessionTime() // wait for the privileged session to time out
      cy.get('button[value="password"]').click()

      password = up(password)
      cy.visit(APP_URL + '/settings')
      cy.get('input[name="password"]').clear().type(password)
      cy.get('button[value="password"]').click()

      cy.reauth({ expect: { email }, type: { password: down(password) } })

      cy.clearCookies()
      cy.login({ email, password })

      cy.clearCookies()
      cy.login({ email, password: up(password), expectSession: false })
    })
  })
})
