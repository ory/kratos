// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { OryKratosConfiguration } from "../../shared/config"

export class ConfigBuilder {
  constructor(readonly config: OryKratosConfiguration) {}

  public build() {
    return this.config
  }

  public longRecoveryLifespan() {
    this.config.selfservice.flows.recovery.lifespan = "1h"
    return this
  }

  public longLinkLifespan() {
    this.config.selfservice.methods.link.config.lifespan = "1m"
    return this
  }

  public disableVerification() {
    this.config.selfservice.flows.verification.enabled = false
    return this
  }

  public disableCodeMfa() {
    this.config.selfservice.methods.code.mfa_enabled = false
    return this
  }

  public enableRecovery() {
    if (!this.config.selfservice.flows.recovery) {
      this.config.selfservice.flows.recovery = {}
    }
    this.config.selfservice.flows.recovery.enabled = true
    return this
  }

  public disableRecovery() {
    this.config.selfservice.flows.recovery.enabled = false
    return this
  }

  public enableVerification() {
    this.config.selfservice.flows.verification.enabled = true
    return this
  }

  public useRecoveryStrategy(strategy: "link" | "code") {
    if (!this.config.selfservice.flows.recovery) {
      this.config.selfservice.flows.recovery = {}
    }
    this.config.selfservice.flows.recovery.use = strategy
    if (!this.config.selfservice.methods[strategy]) {
      this.config.selfservice.methods[strategy] = {}
    }
    this.config.selfservice.methods[strategy].enabled = true
    return this
  }

  public disableRecoveryStrategy(strategy: "link" | "code") {
    this.config.selfservice.methods[strategy].enabled = false
    return this
  }

  public notifyUnknownRecipients(
    flow: "recovery" | "verification",
    value: boolean,
  ) {
    this.config.selfservice.flows[flow].notify_unknown_recipients = value
    return this
  }

  public longCodeLifespan() {
    this.config.selfservice.methods.code.config.lifespan = "1m"
    return this
  }

  public shortCodeLifespan() {
    this.config.selfservice.methods.code.config.lifespan = "1ms"
    return this
  }

  public longPrivilegedSessionTime() {
    this.config.selfservice.flows.settings.privileged_session_max_age = "5m"
    return this
  }

  public requireStrictAal() {
    this.config.selfservice.flows.settings.required_aal = "highest_available"
    this.config.session.whoami.required_aal = "highest_available"
    return this
  }

  public enableLoginForVerifiedAddressOnly() {
    this.config.selfservice.flows.login["after"] = {
      password: { hooks: [{ hook: "require_verified_address" }] },
    }
    return this
  }

  public longVerificationLifespan() {
    this.config.selfservice.flows.verification.lifespan = "1m"
    return this
  }

  public longLifespan(strategy: "link" | "code") {
    this.config.selfservice.methods[strategy].config.lifespan = "1m"
    return this
  }

  public useVerificationStrategy(strategy: "link" | "code") {
    this.config.selfservice.flows.verification.use = strategy
    if (!this.config.selfservice.methods[strategy]) {
      this.config.selfservice.methods[strategy] = {}
    }
    this.config.selfservice.methods[strategy].enabled = true
    return this
  }

  public resetCourierTemplates(
    type: "recovery" | "verification" | "recovery_code" | "verification_code",
  ) {
    if (
      this.config.courier?.templates &&
      type in this.config.courier.templates
    ) {
      delete this.config.courier.templates[type]
    }
    return this
  }

  public useLaxSessionAal() {
    this.config.session.whoami.required_aal = "aal1"
    return this
  }

  public useLaxSettingsFlowAal() {
    this.config.selfservice.flows.settings.required_aal = "aal1"
    return this
  }

  public useLaxAal() {
    return this.useLaxSessionAal().useLaxSettingsFlowAal()
  }

  public useHighestSessionAal() {
    this.config.session.whoami.required_aal = "highest_available"
    return this
  }

  public useHighestSettingsFlowAal() {
    this.config.selfservice.flows.settings.required_aal = "highest_available"
    return this
  }

  public useHighestAvailable() {
    return this.useHighestSessionAal().useHighestSettingsFlowAal()
  }

  public enableCode() {
    this.config.selfservice.methods.code.enabled = true
    return this
  }

  public enableCodeMFA() {
    this.config.selfservice.methods.code.mfa_enabled = true
    return this
  }
}
