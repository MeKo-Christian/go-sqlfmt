-- SQLite Migration: Task management application
-- Version: 003
-- Description: Task management with projects, tags, and time tracking

-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- Create projects table
CREATE TABLE projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    color TEXT DEFAULT '#3B82F6', -- hex color
    is_archived BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CHECK (length(name) > 0),
    CHECK (color LIKE '#%')
);

-- Create tags table
CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    color TEXT DEFAULT '#6B7280',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CHECK (length(name) > 0),
    CHECK (color LIKE '#%')
);

-- Create tasks table
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER,
    title TEXT NOT NULL,
    description TEXT,
    priority INTEGER DEFAULT 2 CHECK (priority BETWEEN 1 AND 4), -- 1=urgent, 2=high, 3=normal, 4=low
    status TEXT DEFAULT 'todo' CHECK (status IN ('todo', 'in_progress', 'done', 'cancelled')),
    due_date DATE,
    estimated_hours REAL CHECK (estimated_hours >= 0),
    actual_hours REAL DEFAULT 0 CHECK (actual_hours >= 0),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL,
    CHECK (length(title) > 0),
    CHECK (due_date IS NULL OR due_date >= date('now')),
    CHECK (actual_hours <= estimated_hours * 2) -- reasonable overtime limit
);

-- Create task_tags junction table
CREATE TABLE task_tags (
    task_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (task_id, tag_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Create time_entries table for time tracking
CREATE TABLE time_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    duration_minutes INTEGER GENERATED ALWAYS AS (
        CASE
            WHEN end_time IS NOT NULL THEN
                CAST((julianday(end_time) - julianday(start_time)) * 24 * 60 AS INTEGER)
            ELSE NULL
        END
    ) STORED,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    CHECK (end_time IS NULL OR end_time >= start_time),
    CHECK (duration_minutes IS NULL OR duration_minutes > 0)
);

-- Create indexes for performance
CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_time_entries_task_id ON time_entries(task_id);
CREATE INDEX idx_time_entries_start_time ON time_entries(start_time);

-- Create triggers for updated_at timestamps
CREATE TRIGGER update_projects_updated_at
    AFTER UPDATE ON projects
    FOR EACH ROW
    BEGIN
        UPDATE projects SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER update_tasks_updated_at
    AFTER UPDATE ON tasks
    FOR EACH ROW
    BEGIN
        UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

-- Insert sample data
INSERT INTO projects (name, description, color) VALUES
('Website Redesign', 'Complete overhaul of company website', '#EF4444'),
('Mobile App', 'Native mobile application development', '#10B981'),
('Database Migration', 'Migrate legacy database to new system', '#F59E0B');

INSERT INTO tags (name, color) VALUES
('urgent', '#DC2626'),
('backend', '#7C3AED'),
('frontend', '#06B6D4'),
('design', '#EC4899'),
('testing', '#84CC16');

-- Insert sample tasks
INSERT INTO tasks (project_id, title, description, priority, status, due_date, estimated_hours) VALUES
(1, 'Design new homepage mockups', 'Create wireframes and mockups for the new homepage design', 2, 'in_progress', '2024-02-15', 16),
(1, 'Implement responsive navigation', 'Build responsive navigation component with mobile menu', 3, 'todo', '2024-02-20', 8),
(2, 'Set up CI/CD pipeline', 'Configure automated testing and deployment pipeline', 1, 'done', '2024-01-30', 12),
(3, 'Data migration script', 'Write script to migrate data from old database', 1, 'in_progress', '2024-02-10', 24);

-- Associate tags with tasks
INSERT INTO task_tags (task_id, tag_id) VALUES
(1, 4), -- design
(2, 3), -- frontend
(3, 2), -- backend
(4, 2); -- backend

-- Insert sample time entries
INSERT INTO time_entries (task_id, start_time, end_time, description) VALUES
(1, '2024-01-15 09:00:00', '2024-01-15 12:00:00', 'Initial design concepts'),
(3, '2024-01-16 14:00:00', '2024-01-16 17:30:00', 'CI/CD setup and configuration'),
(4, '2024-01-17 10:00:00', '2024-01-17 15:00:00', 'Database schema analysis');

-- Create views for common queries
CREATE VIEW active_tasks AS
SELECT
    t.id,
    t.title,
    p.name as project_name,
    t.priority,
    t.status,
    t.due_date,
    t.estimated_hours,
    t.actual_hours
FROM tasks t
LEFT JOIN projects p ON t.project_id = p.id
WHERE t.status IN ('todo', 'in_progress')
ORDER BY
    CASE t.priority
        WHEN 1 THEN 1
        WHEN 2 THEN 2
        WHEN 3 THEN 3
        WHEN 4 THEN 4
    END,
    t.due_date ASC;

CREATE VIEW project_summary AS
SELECT
    p.id,
    p.name,
    COUNT(t.id) as total_tasks,
    COUNT(CASE WHEN t.status = 'done' THEN 1 END) as completed_tasks,
    COUNT(CASE WHEN t.status = 'in_progress' THEN 1 END) as in_progress_tasks,
    ROUND(
        CASE
            WHEN COUNT(t.id) > 0 THEN
                COUNT(CASE WHEN t.status = 'done' THEN 1 END) * 100.0 / COUNT(t.id)
            ELSE 0
        END,
        1
    ) as completion_percentage,
    SUM(t.estimated_hours) as total_estimated_hours,
    SUM(t.actual_hours) as total_actual_hours
FROM projects p
LEFT JOIN tasks t ON p.id = t.project_id
WHERE p.is_archived = FALSE
GROUP BY p.id, p.name;