// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import fs from "fs"
import { default_config } from "./default_config"

export default async function globalSetup() {
  await fs.promises.writeFile(
    "playwright/kratos.config.json",
    JSON.stringify(default_config),
  )
}
