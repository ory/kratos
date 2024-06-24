// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

export const generateEmail = (prefix: "dev" | "crm" = "dev") =>
  Math.random().toString(36) + Math.random().toString(36) + "@ory.test"

export const generatePassword = () => Math.random().toString(36)
