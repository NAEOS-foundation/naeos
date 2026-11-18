-- Additive, backward-compatible external identity mapping.
-- Existing CRM CUID IDs, password hashes, sessions, and ownership relations are unchanged.
-- Supabase Auth account linking must be performed by a trusted flow after ownership proof.
ALTER TABLE "User"
  ADD COLUMN "supabaseAuthUserId" UUID;

CREATE UNIQUE INDEX "User_supabaseAuthUserId_key"
  ON "User"("supabaseAuthUserId");
