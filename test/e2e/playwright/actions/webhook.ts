// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { request } from "@playwright/test"
import retry from "promise-retry"
import { retryOptions } from "../lib/config"

export const WEBHOOK_TARGET = "http://127.0.0.1:4471"

const baseUrl = WEBHOOK_TARGET

/**
 * Fetches a documented (hopefully) created by web hook
 *
 * @param key
 */
export async function fetchDocument(key: string) {
  const r = await request.newContext()

  return retry(async (retry) => {
    const res = await r.get(documentUrl(key))
    if (res.status() !== 200) {
      const body = await res.text()
      const message = `Expected response code 200 but received ${res.status()}: ${body}`
      return retry(message)
    }
    return await res.json()
  }, retryOptions)
}

/**
 * Fetches a documented (hopefully) created by web hook
 *
 * @param key
 */
export async function deleteDocument(key: string) {
  const r = await request.newContext()

  return retry(async (retry) => {
    const res = await r.delete(documentUrl(key))
    if (res.status() !== 204) {
      const body = await res.text()
      const message = `Expected response code 204 but received ${res.status()}: ${body}`
      return retry(message)
    }
    return
  }, retryOptions)
}

/**
 * Returns the URL for a specific document
 *
 * @param key
 */
export function documentUrl(key: string) {
  return `${baseUrl}/documents/${key}`
}
