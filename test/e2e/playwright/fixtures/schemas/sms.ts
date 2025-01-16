// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

export default {
  $id: "https://schemas.ory.sh/presets/kratos/quickstart/email-password/identity.schema.json",
  $schema: "http://json-schema.org/draft-07/schema#",
  title: "Person",
  type: "object",
  properties: {
    traits: {
      type: "object",
      properties: {
        phone: {
          type: "string",
          format: "tel",
          title: "Your Phone Number",
          minLength: 3,
          "ory.sh/kratos": {
            credentials: {
              code: {
                identifier: true,
                via: "sms",
              },
            },
            verification: {
              via: "sms",
            },
          },
        },
      },
      required: ["phone"],
      additionalProperties: false,
    },
  },
}
