// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { APIResponse } from "playwright-core"
import { Session } from "@ory/kratos-client"

export type SessionWithResponse = {
  session: Session
  response: APIResponse
}
