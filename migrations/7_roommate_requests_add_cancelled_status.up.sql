CREATE TABLE roommate_requests_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    requester_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id INTEGER NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'cancelled')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO roommate_requests_new (id, requester_id, target_id, room_id, status, created_at)
SELECT id, requester_id, target_id, room_id, status, created_at
FROM roommate_requests;

DROP INDEX IF EXISTS idx_roommate_requests_pending_unique;
DROP INDEX IF EXISTS idx_roommate_requests_room_id;
DROP INDEX IF EXISTS idx_roommate_requests_target_id;
DROP INDEX IF EXISTS idx_roommate_requests_requester_id;
DROP TABLE roommate_requests;

ALTER TABLE roommate_requests_new RENAME TO roommate_requests;

CREATE INDEX idx_roommate_requests_requester_id ON roommate_requests(requester_id);
CREATE INDEX idx_roommate_requests_target_id ON roommate_requests(target_id);
CREATE INDEX idx_roommate_requests_room_id ON roommate_requests(room_id);
CREATE UNIQUE INDEX idx_roommate_requests_pending_unique
    ON roommate_requests(requester_id, target_id, room_id)
    WHERE status = 'pending';
