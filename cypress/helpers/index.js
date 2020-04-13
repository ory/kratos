module.exports = {
  APP_URL: (Cypress.env('app_url') || 'http://127.0.0.1:4455').replace(/\/$/,''),
  identity: 'foo@bar.com',
  password: 'JfiAKvhjA9'
}
