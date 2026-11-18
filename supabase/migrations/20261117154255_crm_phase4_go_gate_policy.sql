-- CreateEnum
CREATE TYPE "PolicyEffect" AS ENUM ('ALLOW', 'DENY');

-- CreateEnum
CREATE TYPE "GoGateStatus" AS ENUM ('READY', 'WAITING_FOR_GO', 'APPROVED', 'EXECUTING', 'EXECUTED', 'FAILED', 'EXPIRED', 'REJECTED');

-- CreateEnum
CREATE TYPE "GoGateActionType" AS ENUM ('SEND_EMAIL', 'SEND_MESSAGE', 'PUBLISH_POST', 'CONTACT_PROSPECT', 'CREATE_ISSUE', 'TRIGGER_WORKFLOW', 'MODIFY_EXTERNAL_SYSTEM');

-- CreateTable
CREATE TABLE "PolicyRule" (
    "id" TEXT NOT NULL,
    "resource" TEXT NOT NULL,
    "action" TEXT NOT NULL,
    "role" "UserRole" NOT NULL,
    "effect" "PolicyEffect" NOT NULL,
    "priority" INTEGER NOT NULL DEFAULT 0,
    "enabled" BOOLEAN NOT NULL DEFAULT true,
    "policyVersion" TEXT NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "PolicyRule_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "GoGateRequest" (
    "id" TEXT NOT NULL,
    "actionType" "GoGateActionType" NOT NULL,
    "target" TEXT NOT NULL,
    "payload" JSONB,
    "status" "GoGateStatus" NOT NULL DEFAULT 'READY',
    "policyVersion" TEXT,
    "reason" TEXT,
    "requestedBy" TEXT,
    "approvedBy" TEXT,
    "expiresAt" TIMESTAMP(3),
    "executedAt" TIMESTAMP(3),
    "verifiedAt" TIMESTAMP(3),
    "providerResponse" JSONB,
    "result" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "GoGateRequest_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "PolicyRule_resource_action_idx" ON "PolicyRule"("resource", "action");

-- CreateIndex
CREATE INDEX "PolicyRule_role_idx" ON "PolicyRule"("role");

-- CreateIndex
CREATE INDEX "PolicyRule_enabled_idx" ON "PolicyRule"("enabled");

-- CreateIndex
CREATE INDEX "PolicyRule_policyVersion_idx" ON "PolicyRule"("policyVersion");

-- CreateIndex
CREATE INDEX "GoGateRequest_status_idx" ON "GoGateRequest"("status");

-- CreateIndex
CREATE INDEX "GoGateRequest_requestedBy_idx" ON "GoGateRequest"("requestedBy");

-- CreateIndex
CREATE INDEX "GoGateRequest_actionType_idx" ON "GoGateRequest"("actionType");

-- CreateIndex
CREATE INDEX "GoGateRequest_createdAt_idx" ON "GoGateRequest"("createdAt");

-- AddForeignKey
ALTER TABLE "GoGateRequest" ADD CONSTRAINT "GoGateRequest_requestedBy_fkey" FOREIGN KEY ("requestedBy") REFERENCES "User"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "GoGateRequest" ADD CONSTRAINT "GoGateRequest_approvedBy_fkey" FOREIGN KEY ("approvedBy") REFERENCES "User"("id") ON DELETE SET NULL ON UPDATE CASCADE;
