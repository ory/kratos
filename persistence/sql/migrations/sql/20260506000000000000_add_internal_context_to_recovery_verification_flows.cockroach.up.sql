ALTER TABLE "selfservice_recovery_flows" ADD COLUMN IF NOT EXISTS "internal_context" json;
ALTER TABLE "selfservice_verification_flows" ADD COLUMN IF NOT EXISTS "internal_context" json;
