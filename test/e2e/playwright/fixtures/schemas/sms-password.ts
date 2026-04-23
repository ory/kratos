// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

export default {
  $id: "https://schemas.ory.sh/presets/kratos/test/sms-password/identity.schema.json",
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
              password: {
                identifier: true,
              },
              code: {
                identifier: true,
                via: "sms",
              },
            },
            verification: {
              via: "sms",
            },
            recovery: {
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
