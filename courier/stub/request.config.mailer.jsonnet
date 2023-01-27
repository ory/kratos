function(ctx) {
  recipient: ctx.Recipient,
  template_type: ctx.TemplateType,
  to: if "TemplateData" in ctx && "To" in ctx.TemplateData then ctx.TemplateData.To else null,
  recovery_code: if "TemplateData" in ctx && "RecoveryCode" in ctx.TemplateData then ctx.TemplateData.RecoveryCode else null,
  recovery_url: if "TemplateData" in ctx && "RecoveryURL" in ctx.TemplateData then ctx.TemplateData.RecoveryURL else null,
  verification_url: if "TemplateData" in ctx && "VerificationURL" in ctx.TemplateData then ctx.TemplateData.VerificationURL else null,
  verification_code: if "TemplateData" in ctx && "VerificationCode" in ctx.TemplateData then ctx.TemplateData.VerificationCode else null,
  subject: if "TemplateData" in ctx && "Subject" in ctx.TemplateData then ctx.TemplateData.Subject else null,
  body: if "TemplateData" in ctx && "Body" in ctx.TemplateData then ctx.TemplateData.Body else null
}
