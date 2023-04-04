// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

/// <reference types="cypress" />

const got = require("got")
const CRI = require("chrome-remote-interface")
let criPort = 0,
  criClient = null

/**
 * @type {Cypress.PluginConfig}
 */
module.exports = (on) => {
  // `on` is used to hook into various events Cypress emits
  // `config` is the resolved Cypress config
}

function ensureRdpPort(args) {
  const existing = args.find(
    (arg) => arg.slice(0, 23) === "--remote-debugging-port",
  )

  if (existing) {
    return Number(existing.split("=")[1])
  }

  const port = 40000 + Math.round(Math.random() * 25000)
  args.push(`--remote-debugging-port=${port}`)
  return port
}
