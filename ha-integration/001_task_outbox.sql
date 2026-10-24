-- HA-only prototype migration. NOT wired to Semaphore's migration runner.
-- Apply only to a disposable PostgreSQL HA test database after review.
CREATE TABLE IF NOT EXISTS ha_task_outbox (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id bigint NOT NULL,
    fencing_token bigint NOT NULL CHECK (fencing_token > 0),
    event_type text NOT NULL CHECK (event_type IN ('task.completed')),
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    delivered_at timestamptz,
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    UNIQUE (task_id, fencing_token, event_type)
);
CREATE INDEX IF NOT EXISTS ha_task_outbox_pending
    ON ha_task_outbox (id) WHERE delivered_at IS NULL;
