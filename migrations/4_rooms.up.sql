CREATE TABLE rooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conference_id INTEGER NOT NULL REFERENCES conferences(id) ON DELETE CASCADE,
    room_number TEXT NOT NULL,
    room_type TEXT NOT NULL CHECK (room_type IN ('single', 'double', 'triple')),
    price_per_night REAL NOT NULL,
    capacity INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(conference_id, room_number)
);
CREATE INDEX idx_rooms_conference ON rooms(conference_id);
