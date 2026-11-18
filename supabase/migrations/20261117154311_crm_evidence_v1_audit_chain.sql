ALTER TABLE "AuditEvent"
  ADD COLUMN "schemaVersion" TEXT NOT NULL DEFAULT '1',
  ADD COLUMN "sequence" INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN "previousEventHash" TEXT,
  ADD COLUMN "eventHash" TEXT;

WITH ranked AS (
  SELECT
    "id",
    row_number() OVER (ORDER BY "createdAt", "id")::INTEGER AS sequence,
    lag("id") OVER (ORDER BY "createdAt", "id") AS previous_id
  FROM "AuditEvent"
),
hashed AS (
  SELECT
    "id",
    sequence,
    previous_id,
    md5("id" || '|' || sequence::TEXT || '|' || COALESCE(previous_id, '')) AS event_hash
  FROM ranked
)
UPDATE "AuditEvent" a
SET
  "sequence" = h.sequence,
  "previousEventHash" = p."eventHash",
  "eventHash" = h.event_hash
FROM hashed h
LEFT JOIN "AuditEvent" p ON p."id" = h.previous_id
WHERE a."id" = h."id";

ALTER TABLE "AuditEvent"
  ALTER COLUMN "eventHash" SET NOT NULL;

CREATE UNIQUE INDEX "AuditEvent_eventHash_key" ON "AuditEvent"("eventHash");
CREATE INDEX "AuditEvent_sequence_idx" ON "AuditEvent"("sequence");
