module.exports = {
  Introduction: [
    "index", "quickstart", "install",
  ],
  Concepts: [
    "concepts/index",
    "concepts/ui-user-interface",
    "concepts/identity-user-model",
    "concepts/credentials",
    "concepts/email-sms",
    "concepts/federation",
    "concepts/security"
  ],
  "Self Service": [
    {
      type: "category",
      label: "Flows", items: [
        "self-service/flows/index",
        "self-service/flows/user-login-user-registration",
        "self-service/flows/user-logout",
        "self-service/flows/user-settings-profile-management",
        "self-service/flows/password-reset-account-recovery",
        "self-service/flows/user-facing-errors",
        "self-service/flows/verify-email-account-activation"

      ]
    },
    {
      type: "category",
      label: "Strategies", items: [
        "self-service/strategies/index",
        "self-service/strategies/username-email-password",
        "self-service/strategies/openid-connect-social-sign-in-oauth2"
      ]
    },
    {
      type: "category",
      label: "Hooks", items: [
        "self-service/hooks/index"
      ]
    }
  ],
  Guides: [
    "guides/zero-trust-iap-proxy-identity-access-proxy",
    "guides/multi-tenancy-multitenant"
  ],
  "Reference": [
    "reference/configuration",
    "reference/json-schema-json-paths",
    "reference/html-forms",
    "reference/api"
  ],
  "Debug & Help": [
    "debug/csrf"
  ],
  SDKs: ["sdk/index"],
  "Further Reading": [
    "further-reading/comparison"
  ],
};
