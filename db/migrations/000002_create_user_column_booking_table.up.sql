-- ==========================================
-- MIGRATION: update bookings table
-- ==========================================

-- 1. drop old indexes
DROP INDEX IF EXISTS idx_bookings_ticket_id;
DROP INDEX IF EXISTS idx_bookings_status;

-- 2. modify ticket_id (NOT NULL -> nullable)
ALTER TABLE bookings
    ALTER COLUMN ticket_id DROP NOT NULL;

-- 3. add unique on ticket_id
ALTER TABLE bookings
    ADD CONSTRAINT unique_ticket_booking UNIQUE (ticket_id);

-- 4. add new columns
ALTER TABLE bookings
    ADD COLUMN user_id INT NOT NULL,
    ADD COLUMN event_id INT NOT NULL,
    ADD COLUMN event_category_id INT NOT NULL;

-- 5. rebuild indexes
CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_event_id ON bookings(event_id);
CREATE INDEX idx_bookings_event_category_id ON bookings(event_category_id);
CREATE INDEX idx_bookings_ticket_id ON bookings(ticket_id);
CREATE INDEX idx_bookings_status ON bookings(status);