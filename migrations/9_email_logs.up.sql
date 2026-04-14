CREATE TABLE email_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    booking_id INTEGER NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    email_type TEXT NOT NULL,
    recipient TEXT NOT NULL,
    subject TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('sent', 'failed')),
    attempt INTEGER NOT NULL CHECK (attempt > 0),
    error_details TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_email_logs_booking_id ON email_logs(booking_id);
CREATE INDEX idx_email_logs_created_at ON email_logs(created_at);
