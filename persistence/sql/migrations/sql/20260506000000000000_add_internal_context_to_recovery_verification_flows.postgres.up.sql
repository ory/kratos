ALTER TABLE "selfservice_recovery_flows" ADD COLUMN IF NOT EXISTS "internal_context" jsonb;
ALTER TABLE "selfservice_verification_flows" ADD COLUMN IF NOT EXISTS "internal_context" jsonb;
