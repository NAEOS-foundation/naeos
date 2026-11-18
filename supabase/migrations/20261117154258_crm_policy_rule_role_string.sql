-- AlterTable
ALTER TABLE "PolicyRule" DROP COLUMN "role",
ADD COLUMN     "role" TEXT NOT NULL DEFAULT 'MEMBER';

ALTER TABLE "PolicyRule" ALTER COLUMN "role" DROP DEFAULT;

-- CreateIndex
CREATE INDEX "PolicyRule_role_idx" ON "PolicyRule"("role");
