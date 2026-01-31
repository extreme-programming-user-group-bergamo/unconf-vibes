# 9. Database Schema

## 9.1 Schema DDL

```sql
-- Users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    github_id TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL,
    display_name TEXT NOT NULL,
    privacy_setting TEXT NOT NULL DEFAULT 'public' CHECK (privacy_setting IN ('public', 'private')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_github_id ON users(github_id);

-- Conferences table
CREATE TABLE conferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    location TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    capacity INTEGER NOT NULL,
    hotel_email TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_conferences_slug ON conferences(slug);

-- Conference organizers junction table
CREATE TABLE conference_organizers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conference_id INTEGER NOT NULL REFERENCES conferences(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'admin' CHECK (role IN ('owner', 'admin')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(conference_id, user_id)
);

-- Rooms table
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

-- Bookings table
CREATE TABLE bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id INTEGER NOT NULL REFERENCES rooms(id) ON DELETE RESTRICT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    conference_id INTEGER NOT NULL REFERENCES conferences(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'requested' CHECK (status IN ('requested', 'confirmed', 'cancelled')),
    privacy_setting TEXT NOT NULL DEFAULT 'public' CHECK (privacy_setting IN ('public', 'private')),
    notes TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    confirmed_at DATETIME,
    cancelled_at DATETIME
);

CREATE INDEX idx_bookings_room ON bookings(room_id);
CREATE INDEX idx_bookings_user ON bookings(user_id);
CREATE INDEX idx_bookings_conference ON bookings(conference_id);

-- Partial unique index: one active booking per user per conference
CREATE UNIQUE INDEX idx_bookings_user_conference_active 
    ON bookings(user_id, conference_id) 
    WHERE status != 'cancelled';

-- Roommate requests table
CREATE TABLE roommate_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    requester_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id INTEGER NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    responded_at DATETIME,
    CHECK (requester_id != target_id)
);

-- Email log for audit trail
CREATE TABLE email_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    booking_id INTEGER REFERENCES bookings(id) ON DELETE SET NULL,
    email_type TEXT NOT NULL CHECK (email_type IN ('booking', 'cancellation', 'confirmation')),
    recipient TEXT NOT NULL,
    subject TEXT NOT NULL,
    sent_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'sent' CHECK (status IN ('sent', 'failed', 'queued')),
    error_message TEXT
);
```

---
