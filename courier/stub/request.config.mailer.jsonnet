function(ctx) {
  recipient: ctx.recipient,
  template_type: ctx.template_type,
  to: if "template_data" in ctx && "to" in ctx.template_data then ctx.template_data.to else null,
  recovery_code: if "template_data" in ctx && "recovery_code" in ctx.template_data then ctx.template_data.recovery_code else null,
  recovery_url: if "template_data" in ctx && "recovery_url" in ctx.template_data then ctx.template_data.recovery_url else null,
  verification_url: if "template_data" in ctx && "verification_url" in ctx.template_data then ctx.template_data.verification_url else null,
  verification_code: if "template_data" in ctx && "verification_code" in ctx.template_data then ctx.template_data.verification_code else null,
  subject: if "template_data" in ctx && "subject" in ctx.template_data then ctx.template_data.subject else null,
  body: if "template_data" in ctx && "body" in ctx.template_data then ctx.template_data.body else null,
  html_body: if "template_data" in ctx && "html_body" in ctx.template_data then ctx.template_data.html_body else null
}
