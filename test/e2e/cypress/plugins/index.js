/// <reference types="cypress" />

const got = require('got')
const CRI = require('chrome-remote-interface')
let criPort = 0,
  criClient = null

/**
 * @type {Cypress.PluginConfig}
 */
module.exports = (on) => {
  // `on` is used to hook into various events Cypress emits
  // `config` is the resolved Cypress config
  on('before:browser:launch', (browser, args) => {
    criPort = ensureRdpPort(args.args)
    console.log('criPort is', criPort)
  })

  on('task', {
    httpRequest(params) {
      return got(params).then(({body})=>body)
    },
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
