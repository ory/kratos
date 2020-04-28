module.exports = {
  APP_URL: (Cypress.env('app_url') || 'http://127.0.0.1:4455').replace(/\/$/, ''),
  MAIL_API:  (Cypress.env('mail_url') || 'http://127.0.0.1:4437').replace(/\/$/, ''),
  identity: 'foo@bar.com',
  password: 'JfiAKvhjA9',
  website: 'https://www.ory.sh/',
  gen: {
    email: () => Math.random().toString(36).substring(7) + "@" + Math.random().toString(36).substring(7),
    password: () => Math.random().toString(36)
  }
}
