module.exports = {
  Introduction: [
    "index", "quickstart", "install",
  ],
  Concepts: [
    "concepts/index",
    "concepts/ui-user-interface",
    "concepts/identity-user-model",
    {
      label:'Identity Credentials',
      type: "category",
      items:[
        "concepts/credentials",
        "concepts/credentials/username-email-password",
        "concepts/credentials/openid-connect-oidc-oauth2",
      ]
    },
    "concepts/selfservice-flow-completion",
    "concepts/email-sms",
    "concepts/federation",
    "concepts/security"
  ],
  "Self Service": [
    "self-service/index",
    {
      type: "category",
      label: "Flows",
      items:
        [
        "self-service/flows/index",
        "self-service/flows/user-login-user-registration",
        "self-service/flows/user-logout",
        "self-service/flows/user-settings-profile-management",
        "self-service/flows/password-reset-account-recovery",
        "self-service/flows/user-facing-errors",
        "self-service/flows/verify-email-account-activation",
        "self-service/flows/2fa-mfa-multi-factor-authentication"
      ]
    },
    {
      type: "category",
      label: "Strategies", items: [
        "self-service/strategies/username-email-password",
        "self-service/strategies/openid-connect-social-sign-in-oauth2",
        "self-service/strategies/user-settings-profile"
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
    "guides/sign-in-with-github-google-facebook-linkedin",
    "guides/zero-trust-iap-proxy-identity-access-proxy",
    "guides/multi-tenancy-multitenant",
    "guides/high-availability-ha"
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
