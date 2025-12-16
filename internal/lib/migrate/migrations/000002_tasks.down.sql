-- Drop trigger and function
DROP TRIGGER IF EXISTS trg_update_tasks_updated_at ON tasks;
DROP FUNCTION IF EXISTS update_tasks_updated_at();

-- Drop tasks table
DROP TABLE IF EXISTS tasks;

-- Drop task_status enum
DROP TYPE IF EXISTS task_status;
