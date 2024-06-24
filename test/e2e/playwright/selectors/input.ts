// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { Locator, Page } from "@playwright/test"

export interface InputLocator {
  input: Locator
  message: Locator
  label: Locator
}

export const createInputLocator = (page: Page, field: string): InputLocator => {
  const prefix = `[data-testid="node/input/${field}"]`
  return {
    input: page.locator(`${prefix} input`),
    label: page.locator(`${prefix} label`),
    message: page.locator(`${prefix} p`),
  }
}
