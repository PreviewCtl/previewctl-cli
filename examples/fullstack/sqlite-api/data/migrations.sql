CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO notes (id, title, content) VALUES
    (1, 'Welcome', 'This is the first note from the seed data.'),
    (2, 'Setup', 'Preview environment provisioned by previewctl.'),
    (3, 'Testing', 'SQLite prestart seed copied the database file.');
