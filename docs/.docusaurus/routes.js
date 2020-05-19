
import React from 'react';
import ComponentCreator from '@docusaurus/ComponentCreator';

export default [
  
{
  path: '/kratos/docs/',
  component: ComponentCreator('/kratos/docs/'),
  exact: true,
  
},
{
  path: '/kratos/docs/versions',
  component: ComponentCreator('/kratos/docs/versions'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/:route',
  component: ComponentCreator('/kratos/docs/next/:route'),
  
  routes: [
{
  path: '/kratos/docs/next/concepts/credentials',
  component: ComponentCreator('/kratos/docs/next/concepts/credentials'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/concepts/credentials/openid-connect-oidc-oauth2',
  component: ComponentCreator('/kratos/docs/next/concepts/credentials/openid-connect-oidc-oauth2'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/concepts/credentials/username-email-password',
  component: ComponentCreator('/kratos/docs/next/concepts/credentials/username-email-password'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/concepts/email-sms',
  component: ComponentCreator('/kratos/docs/next/concepts/email-sms'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/concepts/federation',
  component: ComponentCreator('/kratos/docs/next/concepts/federation'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/concepts/identity-user-model',
  component: ComponentCreator('/kratos/docs/next/concepts/identity-user-model'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/concepts/index',
  component: ComponentCreator('/kratos/docs/next/concepts/index'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/concepts/securing-applications',
  component: ComponentCreator('/kratos/docs/next/concepts/securing-applications'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/concepts/security',
  component: ComponentCreator('/kratos/docs/next/concepts/security'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/concepts/selfservice-flow-completion',
  component: ComponentCreator('/kratos/docs/next/concepts/selfservice-flow-completion'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/concepts/ui-user-interface',
  component: ComponentCreator('/kratos/docs/next/concepts/ui-user-interface'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/debug/csrf',
  component: ComponentCreator('/kratos/docs/next/debug/csrf'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/fallback/error',
  component: ComponentCreator('/kratos/docs/next/fallback/error'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/fallback/login',
  component: ComponentCreator('/kratos/docs/next/fallback/login'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/fallback/mfa',
  component: ComponentCreator('/kratos/docs/next/fallback/mfa'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/fallback/registration',
  component: ComponentCreator('/kratos/docs/next/fallback/registration'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/fallback/settings',
  component: ComponentCreator('/kratos/docs/next/fallback/settings'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/fallback/verify',
  component: ComponentCreator('/kratos/docs/next/fallback/verify'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/further-reading/comparison',
  component: ComponentCreator('/kratos/docs/next/further-reading/comparison'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/further-reading/contrib',
  component: ComponentCreator('/kratos/docs/next/further-reading/contrib'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/guides/high-availability-ha',
  component: ComponentCreator('/kratos/docs/next/guides/high-availability-ha'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/guides/multi-tenancy-multitenant',
  component: ComponentCreator('/kratos/docs/next/guides/multi-tenancy-multitenant'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/guides/sign-in-with-github-google-facebook-linkedin',
  component: ComponentCreator('/kratos/docs/next/guides/sign-in-with-github-google-facebook-linkedin'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/guides/zero-trust-iap-proxy-identity-access-proxy',
  component: ComponentCreator('/kratos/docs/next/guides/zero-trust-iap-proxy-identity-access-proxy'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/index',
  component: ComponentCreator('/kratos/docs/next/index'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/install',
  component: ComponentCreator('/kratos/docs/next/install'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/quickstart',
  component: ComponentCreator('/kratos/docs/next/quickstart'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/reference/api',
  component: ComponentCreator('/kratos/docs/next/reference/api'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/reference/configuration',
  component: ComponentCreator('/kratos/docs/next/reference/configuration'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/reference/html-forms',
  component: ComponentCreator('/kratos/docs/next/reference/html-forms'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/reference/json-schema-json-paths',
  component: ComponentCreator('/kratos/docs/next/reference/json-schema-json-paths'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/reference/jsonnet',
  component: ComponentCreator('/kratos/docs/next/reference/jsonnet'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/sdk',
  component: ComponentCreator('/kratos/docs/next/sdk'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/self-service',
  component: ComponentCreator('/kratos/docs/next/self-service'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/self-service/flows/2fa-mfa-multi-factor-authentication',
  component: ComponentCreator('/kratos/docs/next/self-service/flows/2fa-mfa-multi-factor-authentication'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/self-service/flows/password-reset-account-recovery',
  component: ComponentCreator('/kratos/docs/next/self-service/flows/password-reset-account-recovery'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/self-service/flows/user-facing-errors',
  component: ComponentCreator('/kratos/docs/next/self-service/flows/user-facing-errors'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/self-service/flows/user-login-user-registration',
  component: ComponentCreator('/kratos/docs/next/self-service/flows/user-login-user-registration'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/self-service/flows/user-login-user-registration/openid-connect-social-sign-in-oauth2',
  component: ComponentCreator('/kratos/docs/next/self-service/flows/user-login-user-registration/openid-connect-social-sign-in-oauth2'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/self-service/flows/user-login-user-registration/username-email-password',
  component: ComponentCreator('/kratos/docs/next/self-service/flows/user-login-user-registration/username-email-password'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/self-service/flows/user-logout',
  component: ComponentCreator('/kratos/docs/next/self-service/flows/user-logout'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/self-service/flows/user-settings',
  component: ComponentCreator('/kratos/docs/next/self-service/flows/user-settings'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/self-service/flows/user-settings/change-password',
  component: ComponentCreator('/kratos/docs/next/self-service/flows/user-settings/change-password'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/self-service/flows/user-settings/link-unlink-openid-connect-oauth2',
  component: ComponentCreator('/kratos/docs/next/self-service/flows/user-settings/link-unlink-openid-connect-oauth2'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/self-service/flows/user-settings/user-profile-management',
  component: ComponentCreator('/kratos/docs/next/self-service/flows/user-settings/user-profile-management'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/self-service/flows/verify-email-account-activation',
  component: ComponentCreator('/kratos/docs/next/self-service/flows/verify-email-account-activation'),
  exact: true,
  
},
{
  path: '/kratos/docs/next/self-service/hooks/index',
  component: ComponentCreator('/kratos/docs/next/self-service/hooks/index'),
  exact: true,
  
}],
},
{
  path: '/kratos/docs/v0.1/:route',
  component: ComponentCreator('/kratos/docs/v0.1/:route'),
  
  routes: [
{
  path: '/kratos/docs/v0.1/concepts/credentials',
  component: ComponentCreator('/kratos/docs/v0.1/concepts/credentials'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/concepts/email-sms',
  component: ComponentCreator('/kratos/docs/v0.1/concepts/email-sms'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/concepts/federation',
  component: ComponentCreator('/kratos/docs/v0.1/concepts/federation'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/concepts/identity-user-model',
  component: ComponentCreator('/kratos/docs/v0.1/concepts/identity-user-model'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/concepts/index',
  component: ComponentCreator('/kratos/docs/v0.1/concepts/index'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/concepts/securing-applications',
  component: ComponentCreator('/kratos/docs/v0.1/concepts/securing-applications'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/concepts/security',
  component: ComponentCreator('/kratos/docs/v0.1/concepts/security'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/concepts/ui-user-interface',
  component: ComponentCreator('/kratos/docs/v0.1/concepts/ui-user-interface'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/further-reading/comparison',
  component: ComponentCreator('/kratos/docs/v0.1/further-reading/comparison'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/further-reading/contrib',
  component: ComponentCreator('/kratos/docs/v0.1/further-reading/contrib'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/index',
  component: ComponentCreator('/kratos/docs/v0.1/index'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/install',
  component: ComponentCreator('/kratos/docs/v0.1/install'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/quickstart',
  component: ComponentCreator('/kratos/docs/v0.1/quickstart'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/reference/api',
  component: ComponentCreator('/kratos/docs/v0.1/reference/api'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/reference/configuration',
  component: ComponentCreator('/kratos/docs/v0.1/reference/configuration'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/reference/html-forms',
  component: ComponentCreator('/kratos/docs/v0.1/reference/html-forms'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/reference/json-schema-json-paths',
  component: ComponentCreator('/kratos/docs/v0.1/reference/json-schema-json-paths'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/sdk/api',
  component: ComponentCreator('/kratos/docs/v0.1/sdk/api'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/sdk/index',
  component: ComponentCreator('/kratos/docs/v0.1/sdk/index'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/self-service/flows/index',
  component: ComponentCreator('/kratos/docs/v0.1/self-service/flows/index'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/self-service/flows/password-reset-account-recovery',
  component: ComponentCreator('/kratos/docs/v0.1/self-service/flows/password-reset-account-recovery'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/self-service/flows/user-facing-errors',
  component: ComponentCreator('/kratos/docs/v0.1/self-service/flows/user-facing-errors'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/self-service/flows/user-login-user-registration',
  component: ComponentCreator('/kratos/docs/v0.1/self-service/flows/user-login-user-registration'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/self-service/flows/user-logout',
  component: ComponentCreator('/kratos/docs/v0.1/self-service/flows/user-logout'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/self-service/flows/user-profile-management',
  component: ComponentCreator('/kratos/docs/v0.1/self-service/flows/user-profile-management'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/self-service/flows/verify-email-account-activation',
  component: ComponentCreator('/kratos/docs/v0.1/self-service/flows/verify-email-account-activation'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/self-service/strategies/index',
  component: ComponentCreator('/kratos/docs/v0.1/self-service/strategies/index'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/self-service/strategies/openid-connect-social-sign-in-oauth2',
  component: ComponentCreator('/kratos/docs/v0.1/self-service/strategies/openid-connect-social-sign-in-oauth2'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/self-service/strategies/username-email-password',
  component: ComponentCreator('/kratos/docs/v0.1/self-service/strategies/username-email-password'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/self-service/workflows/jobs/after',
  component: ComponentCreator('/kratos/docs/v0.1/self-service/workflows/jobs/after'),
  exact: true,
  
},
{
  path: '/kratos/docs/v0.1/self-service/workflows/jobs/before',
  component: ComponentCreator('/kratos/docs/v0.1/self-service/workflows/jobs/before'),
  exact: true,
  
}],
},
{
  path: '/kratos/docs/:route',
  component: ComponentCreator('/kratos/docs/:route'),
  
  routes: [
{
  path: '/kratos/docs/concepts/credentials',
  component: ComponentCreator('/kratos/docs/concepts/credentials'),
  exact: true,
  
},
{
  path: '/kratos/docs/concepts/credentials/openid-connect-oidc-oauth2',
  component: ComponentCreator('/kratos/docs/concepts/credentials/openid-connect-oidc-oauth2'),
  exact: true,
  
},
{
  path: '/kratos/docs/concepts/credentials/username-email-password',
  component: ComponentCreator('/kratos/docs/concepts/credentials/username-email-password'),
  exact: true,
  
},
{
  path: '/kratos/docs/concepts/email-sms',
  component: ComponentCreator('/kratos/docs/concepts/email-sms'),
  exact: true,
  
},
{
  path: '/kratos/docs/concepts/federation',
  component: ComponentCreator('/kratos/docs/concepts/federation'),
  exact: true,
  
},
{
  path: '/kratos/docs/concepts/identity-user-model',
  component: ComponentCreator('/kratos/docs/concepts/identity-user-model'),
  exact: true,
  
},
{
  path: '/kratos/docs/concepts/index',
  component: ComponentCreator('/kratos/docs/concepts/index'),
  exact: true,
  
},
{
  path: '/kratos/docs/concepts/securing-applications',
  component: ComponentCreator('/kratos/docs/concepts/securing-applications'),
  exact: true,
  
},
{
  path: '/kratos/docs/concepts/security',
  component: ComponentCreator('/kratos/docs/concepts/security'),
  exact: true,
  
},
{
  path: '/kratos/docs/concepts/selfservice-flow-completion',
  component: ComponentCreator('/kratos/docs/concepts/selfservice-flow-completion'),
  exact: true,
  
},
{
  path: '/kratos/docs/concepts/ui-user-interface',
  component: ComponentCreator('/kratos/docs/concepts/ui-user-interface'),
  exact: true,
  
},
{
  path: '/kratos/docs/debug/csrf',
  component: ComponentCreator('/kratos/docs/debug/csrf'),
  exact: true,
  
},
{
  path: '/kratos/docs/fallback/error',
  component: ComponentCreator('/kratos/docs/fallback/error'),
  exact: true,
  
},
{
  path: '/kratos/docs/fallback/login',
  component: ComponentCreator('/kratos/docs/fallback/login'),
  exact: true,
  
},
{
  path: '/kratos/docs/fallback/mfa',
  component: ComponentCreator('/kratos/docs/fallback/mfa'),
  exact: true,
  
},
{
  path: '/kratos/docs/fallback/registration',
  component: ComponentCreator('/kratos/docs/fallback/registration'),
  exact: true,
  
},
{
  path: '/kratos/docs/fallback/settings',
  component: ComponentCreator('/kratos/docs/fallback/settings'),
  exact: true,
  
},
{
  path: '/kratos/docs/fallback/verify',
  component: ComponentCreator('/kratos/docs/fallback/verify'),
  exact: true,
  
},
{
  path: '/kratos/docs/further-reading/comparison',
  component: ComponentCreator('/kratos/docs/further-reading/comparison'),
  exact: true,
  
},
{
  path: '/kratos/docs/further-reading/contrib',
  component: ComponentCreator('/kratos/docs/further-reading/contrib'),
  exact: true,
  
},
{
  path: '/kratos/docs/guides/high-availability-ha',
  component: ComponentCreator('/kratos/docs/guides/high-availability-ha'),
  exact: true,
  
},
{
  path: '/kratos/docs/guides/multi-tenancy-multitenant',
  component: ComponentCreator('/kratos/docs/guides/multi-tenancy-multitenant'),
  exact: true,
  
},
{
  path: '/kratos/docs/guides/zero-trust-iap-proxy-identity-access-proxy',
  component: ComponentCreator('/kratos/docs/guides/zero-trust-iap-proxy-identity-access-proxy'),
  exact: true,
  
},
{
  path: '/kratos/docs/index',
  component: ComponentCreator('/kratos/docs/index'),
  exact: true,
  
},
{
  path: '/kratos/docs/install',
  component: ComponentCreator('/kratos/docs/install'),
  exact: true,
  
},
{
  path: '/kratos/docs/quickstart',
  component: ComponentCreator('/kratos/docs/quickstart'),
  exact: true,
  
},
{
  path: '/kratos/docs/reference/api',
  component: ComponentCreator('/kratos/docs/reference/api'),
  exact: true,
  
},
{
  path: '/kratos/docs/reference/configuration',
  component: ComponentCreator('/kratos/docs/reference/configuration'),
  exact: true,
  
},
{
  path: '/kratos/docs/reference/html-forms',
  component: ComponentCreator('/kratos/docs/reference/html-forms'),
  exact: true,
  
},
{
  path: '/kratos/docs/reference/json-schema-json-paths',
  component: ComponentCreator('/kratos/docs/reference/json-schema-json-paths'),
  exact: true,
  
},
{
  path: '/kratos/docs/sdk/api',
  component: ComponentCreator('/kratos/docs/sdk/api'),
  exact: true,
  
},
{
  path: '/kratos/docs/sdk/index',
  component: ComponentCreator('/kratos/docs/sdk/index'),
  exact: true,
  
},
{
  path: '/kratos/docs/self-service/flows/2fa-mfa-multi-factor-authentication',
  component: ComponentCreator('/kratos/docs/self-service/flows/2fa-mfa-multi-factor-authentication'),
  exact: true,
  
},
{
  path: '/kratos/docs/self-service/flows/index',
  component: ComponentCreator('/kratos/docs/self-service/flows/index'),
  exact: true,
  
},
{
  path: '/kratos/docs/self-service/flows/password-reset-account-recovery',
  component: ComponentCreator('/kratos/docs/self-service/flows/password-reset-account-recovery'),
  exact: true,
  
},
{
  path: '/kratos/docs/self-service/flows/user-facing-errors',
  component: ComponentCreator('/kratos/docs/self-service/flows/user-facing-errors'),
  exact: true,
  
},
{
  path: '/kratos/docs/self-service/flows/user-login-user-registration',
  component: ComponentCreator('/kratos/docs/self-service/flows/user-login-user-registration'),
  exact: true,
  
},
{
  path: '/kratos/docs/self-service/flows/user-logout',
  component: ComponentCreator('/kratos/docs/self-service/flows/user-logout'),
  exact: true,
  
},
{
  path: '/kratos/docs/self-service/flows/user-settings-profile-management',
  component: ComponentCreator('/kratos/docs/self-service/flows/user-settings-profile-management'),
  exact: true,
  
},
{
  path: '/kratos/docs/self-service/flows/verify-email-account-activation',
  component: ComponentCreator('/kratos/docs/self-service/flows/verify-email-account-activation'),
  exact: true,
  
},
{
  path: '/kratos/docs/self-service/hooks/index',
  component: ComponentCreator('/kratos/docs/self-service/hooks/index'),
  exact: true,
  
},
{
  path: '/kratos/docs/self-service/index',
  component: ComponentCreator('/kratos/docs/self-service/index'),
  exact: true,
  
},
{
  path: '/kratos/docs/self-service/strategies/openid-connect-social-sign-in-oauth2',
  component: ComponentCreator('/kratos/docs/self-service/strategies/openid-connect-social-sign-in-oauth2'),
  exact: true,
  
},
{
  path: '/kratos/docs/self-service/strategies/user-settings-profile',
  component: ComponentCreator('/kratos/docs/self-service/strategies/user-settings-profile'),
  exact: true,
  
},
{
  path: '/kratos/docs/self-service/strategies/username-email-password',
  component: ComponentCreator('/kratos/docs/self-service/strategies/username-email-password'),
  exact: true,
  
}],
},
  
  {
    path: '*',
    component: ComponentCreator('*')
  }
];
