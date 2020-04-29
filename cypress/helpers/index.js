const email = () =>
  Math.random().toString(36).substring(7) +
  '@' +
  Math.random().toString(36).substring(7)
const password = () => Math.random().toString(36)

const assertAddress = ({ isVerified, email }) => ({ identity }) => {
  expect(identity).to.have.property('addresses')
  expect(identity.addresses).to.have.length(1)

  const address = identity.addresses[0]
  expect(address.id).to.not.be.empty
  expect(address.verified).to.equal(isVerified)
  expect(address.value).to.equal(email)

  if (isVerified) {
    expect(address.verified_at).to.not.be.null
  } else {
    expect(address.verified_at).to.be.null
  }
}

const parseHtml = (html) => new DOMParser().parseFromString(html, 'text/html')

module.exports = {
  APP_URL: (Cypress.env('app_url') || 'http://127.0.0.1:4455').replace(
    /\/$/,
    ''
  ),
  MAIL_API: (Cypress.env('mail_url') || 'http://127.0.0.1:4437').replace(
    /\/$/,
    ''
  ),
  identity: 'foo@bar.com',
  password: 'JfiAKvhjA9',
  website: 'https://www.ory.sh/',
  parseHtml,
  gen: {
    email,
    password,
    identity: () => ({ email: email(), password: password() }),
  },
  assertAddress,

  // Format is
  //  http://127.0.0.1:4455/.ory/kratos/public/self-service/browser/flows/verification/email/confirm/OdTRmdMKe0DfF6ScaOFYgWJwoAprTxnA
  verifyHrefPattern: /^http:.*\/.ory\/kratos\/public\/self-service\/browser\/flows\/verification\/email\/confirm\/([a-zA-Z0-9]+)$/,

  // intervals define how long to wait for something,
  pollInterval: 100, // how long to wait before retry
  verifyLifespan: 5000, // how long to wait before retry
  privilegedLifespan: 5000, // how long to wait before retry
}
