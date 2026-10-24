-- HA-only prototype migration; run after 002_outbox_delivery.sql.
-- A poison event remains inspectable instead of being silently discarded.
ALTER TABLE ha_task_outbox
 ADD COLUMN IF NOT EXISTS available_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 ADD COLUMN IF NOT EXISTS dead_lettered_at timestamptz,
 ADD COLUMN IF NOT EXISTS last_error text;
CREATE INDEX IF NOT EXISTS ha_task_outbox_retry
 ON ha_task_outbox (available_at,id)
 WHERE delivered_at IS NULL AND dead_lettered_at IS NULL;
