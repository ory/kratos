// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { Session as KratosSession } from "@ory/kratos-client"
import { OryKratosConfiguration } from "./config"

export interface MailMessage {
  fromAddress: string
  toAddresses: Array<string>
  body: string
  subject: string
}

export type Strategy = "code" | "link"
type app = "express" | "react"

declare global {
  namespace Cypress {
    interface Chainable {
      /**
       * Delete mails from the email server.
       *
       * @param options
       */
      deleteMail(options: { atLeast?: boolean }): Chainable<void>

      /**
       * Adds end enables a WebAuth authenticator key.
       */
      addVirtualAuthenticator(): Chainable<any>

      /**
       * Fetch the browser's Ory Session.
       *
       * @param opts
       */
      getSession(opts?: {
        expectAal?: "aal2" | "aal1"
        expectMethods?: Array<
          "password" | "webauthn" | "lookup_secret" | "totp"
        >
      }): Chainable<KratosSession>

      /**
       * Expect that the browser has no valid Ory Kratos Cookie Session.
       */
      noSession(): Chainable<Response<any>>

      /**
       * Log a user in
       *
       * @param opts
       */
      login(opts: {
        email: string
        password: string
        expectSession?: boolean
        cookieUrl?: string
      }): Chainable<Response<KratosSession | undefined>>

      /**
       * Sign up a user
       *
       * @param opts
       */
      register(opts: {
        email: string
        password: string
        query?: { [key: string]: string }
        fields?: { [key: string]: any }
      }): Chainable<Response<void>>

      /**
       * Register a user with a code
       *
       * @param opts
       */
      registerWithCode(opts: {
        email: string
        code?: string
        query?: { [key: string]: string }
      }): Chainable<Response<void>>

      /**
       * Updates a user's settings using an API flow
       *
       * @param opts
       */
      settingsApi(opts: {
        fields?: { [key: string]: any }
      }): Chainable<Response<void>>

      /**
       * Set the "privileged session lifespan" to a large value.
       */
      longPrivilegedSessionTime(): Chainable<void>

      /**
       * Return and optionally delete an email from the fake email server.
       *
       * @param opts
       */
      getMail(opts?: {
        removeMail: boolean
        expectedCount?: number
        email?: string
      }): Chainable<MailMessage>

      performEmailVerification(opts?: {
        expect?: { email?: string; redirectTo?: string }
        strategy?: Strategy
        useLinkFromEmail?: boolean
      }): Chainable<void>

      /**
       * Sets the Ory Kratos configuration profile.
       *
       * @param profile
       */
      useConfigProfile(profile: string): Chainable<void>

      /**
       * Register a new user with email + password via the API.
       *
       * Recommended to use if you need an Ory Session Token or alternatively the user created.
       *
       * @param opts
       */
      registerApi(opts?: {
        email: string
        password: string
        fields: { [key: string]: string }
      }): Chainable<KratosSession>

      /**
       * Submits a recovery flow via the API
       *
       * @param opts
       */
      recoverApi(opts: { email: string; returnTo?: string }): Chainable<void>

      /**
       * Submits a verification flow via the API
       *
       * @param opts
       */
      verificationApi(opts: {
        email: string
        returnTo?: string
        strategy?: Strategy
      }): Chainable<void>

      /**
       * Update the config file
       *
       * @param cb
       */
      updateConfigFile(cb: (arg: OryKratosConfiguration) => any): Chainable<any>

      /**
       * Submits a verification flow via the API
       *
       * @param opts
       */
      verificationApiExpired(opts: {
        email: string
        strategy?: Strategy
        returnTo?: string
      }): Chainable<void>

      /**
       *  Sets the hook.
       *
       * @param hooks
       */
      setupHooks(
        flow:
          | "registration"
          | "login"
          | "recovery"
          | "verification"
          | "settings",
        phase: "before" | "after",
        kind: "password" | "webauthn" | "oidc" | "code",
        hooks: Array<{ hook: string; config?: any }>,
      ): Chainable<void>

      /**
       *  Sets the post registration hook.
       *
       * @param hooks
       */
      setPostPasswordRegistrationHooks(
        hooks: Array<{ hook: string; config?: any }>,
      ): Chainable<void>

      /**
       * Sets the post code registration hook.
       *
       * @param hooks
       */
      setPostCodeRegistrationHooks(
        hooks: Array<{ hook: string; config?: any }>,
      ): Chainable<void>

      /**
       * Submits a verification flow via the Browser
       *
       * @param opts
       */
      verificationBrowser(opts: {
        email: string
        returnTo?: string
      }): Chainable<void>

      /**
       * Changes the config so that the login flow lifespan is very short.
       *
       *
       * Useful when testing expiry of login flows.
       * @see longLoginLifespan()
       */
      shortLoginLifespan(): Chainable<void>

      /**
       * Changes the config so that the login flow lifespan is very long.
       *
       * Useful when testing expiry of login flows.
       *
       * @see shortLoginLifespan()
       */
      longLoginLifespan(): Chainable<void>

      /**
       * Change the config so that `https://www.ory.sh/` is a allowed return to URL.
       */
      browserReturnUrlOry(): Chainable<void>

      /**
       * Change the courier recovery invalid and valid templates to remote base64 strings
       */
      remoteCourierRecoveryTemplates(): Chainable<void>

      /**
       * Resets the remote courier templates for the given template type to their default values
       */
      resetCourierTemplates(
        type: "recovery_code" | "recovery" | "verification",
      ): Chainable<void>

      /**
       * Change the courier recovery code invalid and valid templates to remote base64 strings
       */
      remoteCourierRecoveryCodeTemplates(): Chainable<void>

      /**
       * Changes the config so that the registration flow lifespan is very short.
       *
       * Useful when testing expiry of registration flows.
       *
       * @see longRegisterLifespan()
       */
      shortRegisterLifespan(): Chainable<void>

      /**
       * Changes the config so that the registration flow lifespan is very long.
       *
       * Useful when testing expiry of registration flows.
       *
       * @see shortRegisterLifespan()
       */
      longRegisterLifespan(): Chainable<void>

      /**
       * Changes the config so that the settings privileged lifespan is very long.
       *
       * Useful when testing privileged settings flows.
       *
       * @see longPrivilegedSessionTime()
       */
      shortPrivilegedSessionTime(): Chainable<void>

      /**
       * Re-authenticates a user.
       *
       * @param opts
       */
      reauth(opts: {
        expect: { email: string; success?: boolean }
        type: { email?: string; password?: string }
      }): Chainable<void>

      /**
       * Change the config file to support lookup secrets
       * @param value
       */
      useLookupSecrets(value: boolean): Chainable<void>

      /**
       * Re-authenticates a user.
       *
       * @param opts
       */
      reauthWithOtherAccount(opts: {
        previousUrl: string
        expect: { email: string; success?: boolean }
        type: { email?: string; password?: string }
      }): Chainable<void>

      /**
       * Do not require 2fa for /session/whoami
       */
      sessionRequiresNo2fa(): Chainable<void>

      /**
       * Require 2fa for /session/whoami if available
       */
      sessionRequires2fa(): Chainable<void>

      /**
       * Like sessionRequires2fa but sets this also for settings
       */
      requireStrictAal(): Chainable<void>

      /**
       * Like sessionRequiresNo2fa but sets this also for settings
       */
      useLaxAal(): Chainable<void>

      /**
       * Gets the lookup codes from the settings page
       */
      getLookupSecrets(): Chainable<Array<string>>

      /**
       * Expect the settings to be saved.
       */
      expectSettingsSaved(): Chainable<void>

      clearCookies(
        options?: Partial<Loggable & Timeoutable & { domain: null | string }>,
      ): Chainable<null>

      /**
       * A workaround for cypress not being able to clear cookies properly
       */
      clearAllCookies(): Chainable<null>

      /**
       * Submits a password form by clicking the button with method=password
       */
      submitPasswordForm(): Chainable<null>

      /**
       * Submits a profile form by clicking the button with method=profile
       */
      submitProfileForm(): Chainable<null>

      /**
       * Submits a code form by clicking the button with method=code
       */
      submitCodeForm(): Chainable<void>

      /**
       * Expect a CSRF error to occur
       *
       * @param opts
       */
      shouldHaveCsrfError(opts: { app: string }): Chainable<void>

      /**
       * Expects the app to error if a return_to is used which isn't allowed.
       *
       * @param init
       * @param opts
       */
      shouldErrorOnDisallowedReturnTo(
        init: string,
        opts: { app: string },
      ): Chainable<void>

      /**
       * Expect that the second factor login screen is shown
       */
      shouldShow2FAScreen(): Chainable<void>

      /** Click a webauthn button
       *
       * @param type
       */
      clickWebAuthButton(type: "login" | "register"): Chainable<void>

      /**
       * Sign up a user using Social Sign In
       *
       * @param opts
       */
      registerOidc(opts: {
        app: app
        email?: string
        website?: string
        scopes?: Array<string>
        rememberLogin?: boolean
        rememberConsent?: boolean
        acceptLogin?: boolean
        acceptConsent?: boolean
        expectSession?: boolean
        route?: string
      }): Chainable<void>

      /**
       * Sign in a user using Social Sign In
       *
       * @param opts
       */
      loginOidc(opts: {
        app: app
        expectSession?: boolean
        url?: string
        preTriggerHook?: () => void
      }): Chainable<void>

      /**
       * Triggers a Social Sign In flow for the given provider
       *
       * @param app
       * @param provider
       */
      triggerOidc(app: "react" | "express", provider?: string): Chainable<void>

      /**
       * Changes the config so that the recovery privileged lifespan is very long.
       *
       * Useful when testing privileged recovery flows.
       *
       * @see shortPrivilegedRecoveryTime()
       */
      longRecoveryLifespan(): Chainable<void>

      /**
       * Changes the config so that the recovery privileged lifespan is very short.
       *
       * Useful when testing privileged recovery flows.
       *
       * @see shortPrivilegedRecoveryTime()
       */
      shortRecoveryLifespan(): Chainable<void>

      /**
       * Changes the config so that the verification privileged lifespan is very long.
       *
       * Useful when testing recovery/verification flows.
       *
       * @see shortLinkLifespan()
       */
      longVerificationLifespan(): Chainable<void>

      /**
       * Changes the config so that the verification privileged lifespan is very short.
       *
       * Useful when testing privileged verification flows.
       *
       * @see shortPrivilegedVerificationTime()
       */
      shortVerificationLifespan(): Chainable<void>

      /**
       * Log a user out
       */
      logout(): Chainable<void>

      /**
       * Deletes all mail in the mail mock server.
       *
       * @param opts
       */
      deleteMail(opts?: { atLeast?: number }): Chainable<void>

      /**
       * Changes the config so that the link lifespan is very long.
       *
       * Useful when testing recovery/verification flows.
       *
       * @see shortLinkLifespan()
       */
      longLinkLifespan(): Chainable<void>

      /**
       * Changes the config so that the link lifespan is very short.
       *
       * Useful when testing recovery/verification flows.
       *
       * @see longLinkLifespan()
       */
      shortLinkLifespan(): Chainable<void>

      /**
       * Changes the config so that the code lifespan is very short.
       *
       * Useful when testing recovery/verification flows.
       *
       * @see longCodeLifespan()
       */
      shortCodeLifespan(): Chainable<void>

      /**
       * Sets the `lifespan` of a strategy to 1ms (a short value)
       *
       * Useful to test the behavior if the subject of the strategy expired
       *
       * @param s the strategy
       */
      shortLifespan(s: Strategy): Chainable<void>

      /**
       * Sets the `lifespan` of a strategy to 1m
       *
       * @param s the strategy
       */
      longLifespan(s: Strategy): Chainable<void>

      /**
       * Changes the config so that the code lifespan is very long.
       *
       * Useful when testing recovery/verification flows.
       *
       * @see shortCodeLifespan()
       */
      longCodeLifespan(): Chainable<void>

      /**
       * Expect a recovery email which is expired.
       *
       * @param opts
       */
      recoverEmailButExpired(opts?: {
        expect: { email: string }
      }): Chainable<void>

      /**
       * Expect a recovery email with a recovery code.
       *
       * @param opts
       */
      recoveryEmailWithCode(opts?: {
        expect: { email: string; enterCode?: boolean }
      }): Chainable<void>

      /**
       * Expect a verification email which is expired.
       *
       * @param opts
       */
      verifyEmailButExpired(opts?: {
        expect: { email: string }
        strategy?: Strategy
      }): Chainable<void>

      /**
       * Sets the strategy to use for verification
       *
       * @param strategy the Strategy
       */
      useVerificationStrategy(strategy: Strategy): Chainable<void>

      /**
       * Disables verification
       */
      disableVerification(): Chainable<void>

      /**
       * Enables verification
       */
      enableVerification(): Chainable<void>

      /**
       * Enables recovery
       */
      enableRecovery(): Chainable<void>

      /**
       * Sets the recovery strategy to use
       */
      useRecoveryStrategy(strategy: Strategy): Chainable<void>

      /**
       * Disables a specific recovery strategy
       *
       * @param strategy the recovery strategy to disable
       */
      disableRecoveryStrategy(strategy: Strategy): Chainable<void>

      /**
       * Disabled recovery
       */
      disableRecovery(): Chainable<void>

      /**
       * Disables registration
       */
      disableRegistration(): Chainable<void>

      /**
       * Enables registration
       */
      enableRegistration(): Chainable<void>

      /**
       * Expect a recovery email which is valid.
       *
       * @param opts
       */
      recoverEmail(opts: {
        expect: { email: string }
        shouldVisit?: boolean
      }): Chainable<MailMessage>

      /**
       * Expect a verification email which is valid.
       *
       * @param opts
       */
      verifyEmail(opts: {
        expect: { email: string; password?: string; redirectTo?: string }
        strategy?: Strategy
        shouldVisit?: boolean
      }): Chainable<void>

      /**
       * Configures a hook which only allows verified email addresses to sign in.
       */
      enableLoginForVerifiedAddressOnly(): Chainable<void>

      /**
       * Sets the value for the `notify_unknown_recipients` key for a flow
       *
       * @param flow the flow for which to set the config value
       * @param value the value, defaults to true
       */
      notifyUnknownRecipients(
        flow: "recovery" | "verification",
        value?: boolean,
      ): Chainable<void>

      /**
       * Sign a user in via the API and return the session.
       *
       * @param opts
       */
      loginApi(opts: {
        email: string
        password: string
      }): Chainable<{ session: KratosSession }>

      /**
       * Same as loginApi but uses dark magic to avoid cookie issues.
       *
       * @param opts
       */
      loginApiWithoutCookies(opts: {
        email: string
        password: string
      }): Chainable<{ session: KratosSession }>

      /**
       * Which app to proxy
       */
      proxy(app: "react" | "express"): Chainable<void>

      /**
       * Log a user in on mobile
       *
       * @param opts
       */
      loginMobile(opts: { email: string; password: string }): Chainable<void>

      /**
       * Set the identity schema
       * @param schema
       */
      setIdentitySchema(schema: string): Chainable<void>

      /**
       * Set the default schema
       * @param id
       */
      setDefaultIdentitySchema(id: string): Chainable<void>

      /**
       * Remove the specified attribute from the given HTML elements
       */
      removeAttribute(selectors: string[], attribute: string): Chainable<void>

      /**
       * Add an input element to the DOM as a child of the given parent
       */
      addInputElement(
        parent: string,
        attribute: string,
        value: string,
      ): Chainable<void>

      /**
       * Fetches the courier messages from the admin API
       */
      getCourierMessages(): Chainable<
        { recipient: string; template_type: string }[]
      >

      /**
       * Enable the verification UI after registration hook
       */
      enableVerificationUIAfterRegistration(
        strategy: "password" | "oidc" | "webauthn",
      ): Chainable<void>

      /**
       * Extracts a verification code from the received email
       */
      getVerificationCodeFromEmail(email: string): Chainable<string>

      /**
       * Enables the registration code method
       * @param enable
       */
      enableRegistrationViaCode(enable: boolean): Chainable<void>

      /**
       * Extracts a registration code from the received email
       */
      getRegistrationCodeFromEmail(
        email: string,
        opts?: { expectedCount: number },
      ): Chainable<string>

      /**
       * Extracts a login code from the received email
       */
      getLoginCodeFromEmail(
        email: string,
        opts?: { expectedCount: number },
      ): Chainable<string>
    }
  }
}
