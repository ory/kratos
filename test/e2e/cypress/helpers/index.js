const email = () => Math.random().toString(36) + '@ory.sh'
const password = () => Math.random().toString(36)

const assertVerifiableAddress = ({ isVerified, email }) => ({ identity }) => {
  expect(identity).to.have.property('verifiable_addresses')
  expect(identity.verifiable_addresses).to.have.length(1)

  const address = identity.verifiable_addresses[0]
  expect(address.id).to.not.be.empty
  expect(address.verified).to.equal(isVerified)
  expect(address.value).to.equal(email)

  if (isVerified) {
    expect(address.verified_at).to.not.be.null
  } else {
    expect(address.verified_at).to.be.null
  }
}

const assertRecoveryAddress = ({ email }) => ({ identity }) => {
  expect(identity).to.have.property('recovery_addresses')
  expect(identity.recovery_addresses).to.have.length(1)

  const address = identity.recovery_addresses[0]
  expect(address.id).to.not.be.empty
  expect(address.value).to.equal(email)
}

const parseHtml = (html) => new DOMParser().parseFromString(html, 'text/html')

module.exports = {
  APP_URL: (Cypress.env('app_url') || 'http://127.0.0.1:4455').replace(
    /\/$/,
    ''
  ),
  MOBILE_URL: (Cypress.env('mobile_url') || 'http://127.0.0.1:4457').replace(
    /\/$/,
    ''
  ),
  KRATOS_ADMIN: (Cypress.env('kratos_admin') || 'http://127.0.0.1:4434')
    .replace()
    .replace(/\/$/, ''),
  KRATOS_PUBLIC: (Cypress.env('kratos_public') || 'http://127.0.0.1:4433')
    .replace()
    .replace(/\/$/, ''),
  MAIL_API: (Cypress.env('mail_url') || 'http://127.0.0.1:4437').replace(
    /\/$/,
    ''
  ),
  website: 'https://www.ory.sh/',
  parseHtml,
  gen: {
    email,
    password,
    identity: () => ({ email: email(), password: password() })
  },
  assertVerifiableAddress: assertVerifiableAddress,
  assertRecoveryAddress: assertRecoveryAddress,

  // Format is
  verifyHrefPattern: /^http:.*\/self-service\/verification\?(((&|)token|(&|)flow)=([\-a-zA-Z0-9]+)){2}$/,

  // intervals define how long to wait for something,
  pollInterval: 250, // how long to wait before retry

  // Adding 1+ second on top because MySQL doesn't do millisecs.
  verifyLifespan: 5000 + 1000,
  privilegedLifespan: 5000 + 1000
}
