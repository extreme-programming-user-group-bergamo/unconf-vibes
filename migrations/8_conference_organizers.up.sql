CREATE TABLE conference_organizers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conference_id INTEGER NOT NULL REFERENCES conferences(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'admin' CHECK (role IN ('owner', 'admin')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(conference_id, user_id)
);
CREATE INDEX idx_conference_organizers_conference ON conference_organizers(conference_id);
CREATE INDEX idx_conference_organizers_user ON conference_organizers(user_id);
