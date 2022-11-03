// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { APP_URL, SPA_URL } from "./index"

export const routes = {
  base: APP_URL,
  login: APP_URL + "/login",
  registration: APP_URL + "/registration",
  settings: APP_URL + "/settings",
  recovery: APP_URL + "/recovery",
  verification: APP_URL + "/verification",
}
