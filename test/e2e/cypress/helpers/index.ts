export const email = () => Math.random().toString(36) + '@ory.sh'

export const password = () => Math.random().toString(36)

export const assertVerifiableAddress = ({ isVerified, email }) => (session) => {
  const { identity } = session
  expect(identity).to.have.property('verifiable_addresses')
  expect(identity.verifiable_addresses).to.have.length(1)

  const address = identity.verifiable_addresses[0]
  expect(address.id).to.not.be.empty
  expect(address.verified).to.equal(isVerified)
  expect(address.value).to.equal(email)

  if (isVerified) {
    expect(address.verified_at).to.not.be.null
  } else {
    expect(address).to.not.have.property('verified_at')
  }
}

export const assertRecoveryAddress = ({ email }) => ({ identity }) => {
  expect(identity).to.have.property('recovery_addresses')
  expect(identity.recovery_addresses).to.have.length(1)

  const address = identity.recovery_addresses[0]
  expect(address.id).to.not.be.empty
  expect(address.value).to.equal(email)
}

export const parseHtml = (html) =>
  new DOMParser().parseFromString(html, 'text/html')

export const APP_URL = (
  Cypress.env('app_url') || 'http://localhost:4455'
).replace(/\/$/, '')

export const MOBILE_URL = (
  Cypress.env('mobile_url') || 'http://localhost:4457'
).replace(/\/$/, '')
export const SPA_URL = (
  Cypress.env('react_url') || 'http://localhost:4455'
).replace(/\/$/, '')
export const KRATOS_ADMIN = (
  Cypress.env('kratos_admin') || 'http://localhost:4434'
)
  .replace()
  .replace(/\/$/, '')

export const KRATOS_PUBLIC = (
  Cypress.env('kratos_public') || 'http://localhost:4433'
)
  .replace()
  .replace(/\/$/, '')

export const MAIL_API = (
  Cypress.env('mail_url') || 'http://localhost:4437'
).replace(/\/$/, '')

export const website = 'https://www.ory.sh/'

export const gen = {
  email,
  password,
  identity: () => ({ email: email(), password: password() }),
  identityWithWebsite: () => ({
    email: email(),
    password: password(),
    fields: { 'traits.website': 'https://www.ory.sh' }
  })
}

// Format is
export const verifyHrefPattern = /^http:.*\/self-service\/verification\?(((&|)token|(&|)flow)=([\-a-zA-Z0-9]+)){2}$/

// intervals define how long to wait for something,
export const pollInterval = 250 // how long to wait before retry

// Adding 1+ second on top because MySQL doesn't do millisecs.
export const verifyLifespan = 5000 + 1000
export const privilegedLifespan = 5000 + 1000

export const appPrefix = (app) => `[data-testid="app-${app}"] `
