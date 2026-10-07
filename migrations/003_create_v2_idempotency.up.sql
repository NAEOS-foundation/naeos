CREATE TABLE IF NOT EXISTS naeos_v2_idempotency (
    idempotency_key VARCHAR(255) PRIMARY KEY,
    fingerprint VARCHAR(64) NOT NULL,
    status VARCHAR(16) NOT NULL,
    headers TEXT NOT NULL,
    body TEXT NOT NULL,
    response_status INTEGER NOT NULL,
    expires_at BIGINT NOT NULL

);