-- Create task_status enum
CREATE TYPE task_status AS ENUM ('pending', 'running', 'waiting', 'completed', 'failed');

-- Create tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    task_type TEXT NOT NULL CHECK (task_type <> ''),
    state JSONB NOT NULL DEFAULT '{}',
    status task_status NOT NULL DEFAULT 'pending',
    worker_id TEXT,
    lease_expires_at TIMESTAMPTZ,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- worker_id and lease_expires_at must be set together when running
    CHECK (
        (status = 'running' AND worker_id IS NOT NULL AND lease_expires_at IS NOT NULL) OR
        (status <> 'running' AND worker_id IS NULL AND lease_expires_at IS NULL)
    ),
    -- error must be set iff status is failed
    CHECK (
        (status = 'failed' AND error IS NOT NULL) OR
        (status <> 'failed' AND error IS NULL)
    )
);

-- Index for finding claimable tasks (pending or expired leases)
CREATE INDEX IF NOT EXISTS idx_tasks_claimable ON tasks (status, lease_expires_at)
WHERE status IN ('pending', 'running');

-- Index for finding tasks by type
CREATE INDEX IF NOT EXISTS idx_tasks_type ON tasks (task_type);

-- Trigger to update updated_at on row modification
CREATE OR REPLACE FUNCTION update_tasks_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_tasks_updated_at ON tasks;

CREATE TRIGGER trg_update_tasks_updated_at
BEFORE UPDATE ON tasks
FOR EACH ROW
EXECUTE FUNCTION update_tasks_updated_at();
