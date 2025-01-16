// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import type { OperationOptions } from "retry"

export type RetryOptions = OperationOptions

export const retryOptions: RetryOptions = {
  retries: 20,
  factor: 1,
  maxTimeout: 500,
  minTimeout: 250,
  randomize: false,
}
