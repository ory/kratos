// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { defineConfig, devices } from "@playwright/test"
import * as dotenv from "dotenv"

dotenv.config({ path: __dirname + "/playwright/playwright.env" })

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
  testDir: "./playwright/tests",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1,
  workers: 1,
  reporter: process.env.CI ? [["github"], ["html"], ["list"]] : "html",

  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  use: {
    trace: process.env.CI ? "retain-on-failure" : "on",
  },

  /* Configure projects for major browsers */
  projects: [
    {
      name: "mobile-chrome",
      testMatch: "mobile/**/*.spec.ts",
      use: {
        ...devices["Pixel 5"],
        baseURL: "http://localhost:19006",
      },
    },
    {
      name: "chromium",
      testMatch: "desktop/**/*.spec.ts",
      use: {
        ...devices["Desktop Chrome"],
        baseURL: "http://localhost:4455",
      },
    },
  ],

  webServer: [
    {
      command: [
        "cp test/e2e/playwright/kratos.base-config.json test/e2e/playwright/kratos.config.json",
        "go run -tags sqlite,json1 . migrate sql -e --yes",
        "go run -tags sqlite,json1 . serve --watch-courier --dev -c test/e2e/playwright/kratos.config.json",
      ].join(" && "),
      cwd: "../..",
      url: "http://localhost:4433/health/ready",
      env: {
        DSN: dbToDsn(),
        COURIER_SMTP_CONNECTION_URI:
          "smtp://localhost:8026/?disable_starttls=true",
      },
      timeout: 5 * 60 * 1000, // 5 minutes
    },
    {
      command: "go tool MailHog -smtp-bind-addr=localhost:8026",
      cwd: "../..",
      reuseExistingServer: false,
      url: "http://localhost:8025/",
    },
    {
      command: "go run test/e2e/mock/httptarget/main.go",
      cwd: "../..",
      reuseExistingServer: false,
      url: "http://localhost:4471/health",
    },
  ],
})

function dbToDsn(): string {
  switch (process.env.DB) {
    case "postgres":
      return process.env.TEST_DATABASE_POSTGRESQL
    case "cockroach":
      return process.env.TEST_DATABASE_COCKROACHDB
    case "mysql":
      return process.env.TEST_DATABASE_MYSQL
    case "sqlite":
      return process.env.TEST_DATABASE_SQLITE
    default:
      return "memory"
  }
}
