CREATE TABLE "PolicyDecision" (
  "id" TEXT NOT NULL,
  "decision" "PolicyEffect" NOT NULL,
  "policyVersion" TEXT NOT NULL,
  "matchedRuleId" TEXT,
  "matchedRulePriority" INTEGER,
  "resource" TEXT NOT NULL,
  "action" TEXT NOT NULL,
  "role" TEXT,
  "reason" TEXT NOT NULL,
  "evaluatorVersion" TEXT NOT NULL,
  "evaluatedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT "PolicyDecision_pkey" PRIMARY KEY ("id")
);

ALTER TABLE "GoGateRequest"
  ADD COLUMN "idempotencyKey" TEXT,
  ADD COLUMN "executionId" TEXT,
  ADD COLUMN "provider" TEXT,
  ADD COLUMN "providerRequestId" TEXT,
  ADD COLUMN "executionAttempt" INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN "policyDecisionId" TEXT;

UPDATE "GoGateRequest"
SET "idempotencyKey" = 'legacy-' || "id"
WHERE "idempotencyKey" IS NULL;

ALTER TABLE "GoGateRequest"
  ALTER COLUMN "idempotencyKey" SET NOT NULL;

CREATE UNIQUE INDEX "GoGateRequest_idempotencyKey_key" ON "GoGateRequest"("idempotencyKey");
CREATE INDEX "PolicyDecision_policyVersion_idx" ON "PolicyDecision"("policyVersion");
CREATE INDEX "PolicyDecision_resource_action_idx" ON "PolicyDecision"("resource", "action");
CREATE INDEX "PolicyDecision_evaluatedAt_idx" ON "PolicyDecision"("evaluatedAt");
CREATE INDEX "GoGateRequest_executionId_idx" ON "GoGateRequest"("executionId");
CREATE INDEX "GoGateRequest_provider_providerRequestId_idx" ON "GoGateRequest"("provider", "providerRequestId");

ALTER TABLE "GoGateRequest"
  ADD CONSTRAINT "GoGateRequest_policyDecisionId_fkey"
  FOREIGN KEY ("policyDecisionId") REFERENCES "PolicyDecision"("id")
  ON DELETE SET NULL ON UPDATE CASCADE;
