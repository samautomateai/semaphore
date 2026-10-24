-- Prototype only: run after 001_task_outbox.sql in a disposable HA database.
-- Leases permit crash recovery; delivery consumers MUST be idempotent.
ALTER TABLE ha_task_outbox
 ADD COLUMN IF NOT EXISTS delivery_owner text,
 ADD COLUMN IF NOT EXISTS delivery_lease_expires_at timestamptz;
CREATE INDEX IF NOT EXISTS ha_task_outbox_claimable
 ON ha_task_outbox (id)
 WHERE delivered_at IS NULL;
