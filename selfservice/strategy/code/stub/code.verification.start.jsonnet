local getFlowId(ctx) =
  if std.objectHas(ctx, "VerificationURL") then
    local start = std.findSubstr("flow=", ctx.VerificationURL);
    std.substr(ctx.VerificationURL, start[0]+5, 36)
  else
    "error_getting_flow_id";

local getOperator(ctx) =
  if std.objectHas(ctx, "TransientPayload") && std.objectHas(ctx.TransientPayload, "application") then ctx.TransientPayload.application
  else "monta";

function(ctx) {
  to: ctx.To,
  [if std.objectHas(ctx, "TransientPayload") && std.objectHas(ctx.TransientPayload, "language") then 'language']: ctx.TransientPayload.language,
  [if std.objectHas(ctx, "TransientPayload") && std.objectHas(ctx.TransientPayload, "application") then 'application']: ctx.TransientPayload.application,
  template: if std.objectHas(ctx, "VerificationCode") then 'sms_localisation.verification_code_with_link' else 'sms_localisation.account_activation_passcode',
  templateParameters: {
    "host": "portal.monta.app",
    "flow_id": getFlowId(ctx),
    "operator": getOperator(ctx),
  },
}
