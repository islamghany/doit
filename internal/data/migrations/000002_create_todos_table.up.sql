-- Create custom types
CREATE TYPE todo_status AS ENUM (
    'pending',
    'in_progress',
    'completed',
    'archived'
);
CREATE TYPE todo_priority AS ENUM ('low', 'medium', 'high', 'urgent');
-- Todos table
CREATE TABLE todos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status todo_status NOT NULL DEFAULT 'pending',
    priority todo_priority NOT NULL DEFAULT 'medium',
    tags TEXT [] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    due_date TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT title_not_empty CHECK (length(trim(title)) > 0)
);
-- Create indexes for todos
CREATE INDEX idx_todos_user_id ON todos(user_id);
CREATE INDEX idx_todos_status ON todos(status);
CREATE INDEX idx_todos_priority ON todos(priority);
CREATE INDEX idx_todos_due_date ON todos(due_date);
CREATE INDEX idx_todos_created_at ON todos(created_at DESC);
CREATE INDEX idx_todos_user_status ON todos(user_id, status);
CREATE INDEX idx_todos_tags_gin ON todos USING gin(tags);
CREATE INDEX idx_todos_metadata_gin ON todos USING gin(metadata);
CREATE INDEX idx_todos_title_trgm ON todos USING gin(title gin_trgm_ops);
-- Triggers for updated_at
CREATE TRIGGER update_todos_updated_at BEFORE
UPDATE ON todos FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();