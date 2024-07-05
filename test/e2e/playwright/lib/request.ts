// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { APIResponse } from "playwright-core"
import { expect } from "../fixtures"
import { OperationOptions } from "retry"

export type RetryOptions = OperationOptions

export const retryOptions: RetryOptions = {
  retries: 20,
  factor: 1,
  maxTimeout: 500,
  minTimeout: 250,
  randomize: false,
}

export async function expectJSONResponse<T>(
  res: APIResponse,
  { statusCode = 200, message }: { statusCode?: number; message?: string } = {},
): Promise<T> {
  await expect(res).toMatchResponseData({
    statusCode,
    failureHint: message,
  })
  try {
    return (await res.json()) as T
  } catch (e) {
    const body = await res.text()
    throw Error(
      `Expected to be able to parse body as json: ${e} (body: ${body})`,
    )
  }
}
