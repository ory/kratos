/// <reference types="cypress" />
// ***********************************************************
// This example plugins/index.js can be used to load plugins
//
// You can change the location of this file or turn off loading
// the plugins file with the 'pluginsFile' configuration option.
//
// You can read more here:
// https://on.cypress.io/plugins-guide
// ***********************************************************

// This function is called when a project is opened or re-opened (e.g. due to
// the project's config changing)

// WebAuthn contents taken from
//    https://github.com/OWASP/SSO_Project/commit/ce7269540e0b9895e08d5269d1fee1bca570c0e4#
const CRI = require('chrome-remote-interface')
let criPort = 0,
  criClient = null

/**
 * @type {Cypress.PluginConfig}
 */
module.exports = (on, config) => {
  // `on` is used to hook into various events Cypress emits
  // `config` is the resolved Cypress config

  on('before:browser:launch', (browser, args) => {
    criPort = ensureRdpPort(args.args)
    console.log('criPort is', criPort)
  })

  on('task', {
    // Reset chrome remote interface for clean state
    async resetCRI() {
      if (criClient) {
        await criClient.close()
        criClient = null
      }
      return Promise.resolve(true)
    },
    // Execute CRI command
    async sendCRI(args) {
      criClient = criClient || (await CRI({ port: criPort }))
      return criClient.send(args.query, args.opts)
    }
  })
}

function ensureRdpPort(args) {
  const existing = args.find(
    (arg) => arg.slice(0, 23) === '--remote-debugging-port'
  )

  if (existing) {
    return Number(existing.split('=')[1])
  }

  const port = 40000 + Math.round(Math.random() * 25000)
  args.push(`--remote-debugging-port=${port}`)
  return port
}
