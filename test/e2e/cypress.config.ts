import { defineConfig } from 'cypress'
import got from 'got'
import CRI from 'chrome-remote-interface'
let criPort = 0,
  criClient = null

export default defineConfig({
  chromeWebSecurity: false,
  retries: {
    runMode: 6,
    openMode: 1
  },
  defaultCommandTimeout: 10000,
  requestTimeout: 10000,
  projectId: 'bc48bg',
  video: true,
  videoCompression: false,
  screenshotOnRunFailure: true,
  e2e: {
    videosFolder: 'cypress/videos',
    screenshotsFolder: 'cypress/screenshots',
    excludeSpecPattern: '**/*snapshots.js',
    supportFile: 'cypress/support/index.js',
    specPattern: '**/*.spec.{js,ts}',
    baseUrl: 'http://localhost:4455/',
    setupNodeEvents(on, config) {
      on('before:browser:launch', (browser, args) => {
        criPort = ensureRdpPort(args.args)
        console.log('criPort is', criPort)
      })

      on('task', {
        httpRequest(params) {
          return got(params).then(({ body }) => body)
        },
        // Reset chrome remote interface for clean state
        async resetCRI() {
          if (criClient) {
            const c = criClient
            criClient = null
            await c.close()
          }

          return Promise.resolve(true)
        },
        // Execute CRI command
        async sendCRI(args) {
          if (!criClient) {
            criClient = await CRI({ port: criPort })
          }

          return criClient.send(args.query, args.opts)
        }
      })
    }
  }
})

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
