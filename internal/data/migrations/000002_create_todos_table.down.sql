-- Drop todos table
DROP TABLE todos;
-- Drop indexes for todos
DROP INDEX idx_todos_user_id;
DROP INDEX idx_todos_status;
DROP INDEX idx_todos_priority;
DROP INDEX idx_todos_due_date;
DROP INDEX idx_todos_created_at;
DROP INDEX idx_todos_user_status;
DROP INDEX idx_todos_tags_gin;
DROP INDEX idx_todos_metadata_gin;
DROP INDEX idx_todos_title_trgm;
-- Drop triggers for updated_at
DROP TRIGGER update_todos_updated_at;
DROP FUNCTION update_updated_at_column;
-- Drop custom types
DROP TYPE todo_status;
DROP TYPE todo_priority;