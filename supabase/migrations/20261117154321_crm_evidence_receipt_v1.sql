-- Evidence Receipt + Verification V1
CREATE TYPE "VerificationStatus" AS ENUM ('PENDING', 'VERIFIED', 'FAILED');

CREATE TABLE "EvidenceReceipt" (
  "id" TEXT NOT NULL,
  "goGateRequestId" TEXT NOT NULL,
  "executionId" TEXT NOT NULL,
  "idempotencyKey" TEXT NOT NULL,
  "policyDecisionId" TEXT,
  "policyVersion" TEXT,
  "provider" TEXT,
  "providerRequestId" TEXT,
  "requestDigest" TEXT NOT NULL,
  "providerResponseDigest" TEXT,
  "verificationStatus" "VerificationStatus" NOT NULL DEFAULT 'PENDING',
  "verifiedAt" TIMESTAMP(3),
  "verifierVersion" TEXT NOT NULL DEFAULT 'evidence-v1',
  "receiptHash" TEXT NOT NULL,
  "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updatedAt" TIMESTAMP(3) NOT NULL,

  CONSTRAINT "EvidenceReceipt_pkey" PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "EvidenceReceipt_goGateRequestId_key" ON "EvidenceReceipt"("goGateRequestId");
CREATE UNIQUE INDEX "EvidenceReceipt_receiptHash_key" ON "EvidenceReceipt"("receiptHash");
CREATE INDEX "EvidenceReceipt_executionId_idx" ON "EvidenceReceipt"("executionId");
CREATE INDEX "EvidenceReceipt_idempotencyKey_idx" ON "EvidenceReceipt"("idempotencyKey");
CREATE INDEX "EvidenceReceipt_policyDecisionId_idx" ON "EvidenceReceipt"("policyDecisionId");
CREATE INDEX "EvidenceReceipt_verificationStatus_idx" ON "EvidenceReceipt"("verificationStatus");

ALTER TABLE "EvidenceReceipt"
  ADD CONSTRAINT "EvidenceReceipt_goGateRequestId_fkey"
  FOREIGN KEY ("goGateRequestId") REFERENCES "GoGateRequest"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "EvidenceReceipt"
  ADD CONSTRAINT "EvidenceReceipt_policyDecisionId_fkey"
  FOREIGN KEY ("policyDecisionId") REFERENCES "PolicyDecision"("id") ON DELETE SET NULL ON UPDATE CASCADE;
