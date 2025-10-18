-- Drop triggers for updated_at (before dropping table)
DROP TRIGGER IF EXISTS update_todos_updated_at ON todos;
-- Drop todos table (this also drops all indexes automatically)
DROP TABLE IF EXISTS todos;
-- Drop custom types
DROP TYPE IF EXISTS todo_priority;
DROP TYPE IF EXISTS todo_status;