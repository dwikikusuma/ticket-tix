-- ==========================================
-- MIGRATION DOWN: revert bookings table
-- ==========================================

-- 1. drop new indexes
DROP INDEX IF EXISTS idx_bookings_user_id;
DROP INDEX IF EXISTS idx_bookings_event_category_id;
DROP INDEX IF EXISTS idx_bookings_ticket_id;
DROP INDEX IF EXISTS idx_bookings_status;

-- 2. drop new columns
ALTER TABLE bookings
    DROP COLUMN user_id,
    DROP COLUMN event_category_id;

-- 3. drop unique constraint on ticket_id
ALTER TABLE bookings
    DROP CONSTRAINT unique_ticket_booking;

-- 4. revert ticket_id FK (nullable -> NOT NULL, add CASCADE back)
ALTER TABLE bookings
    DROP CONSTRAINT bookings_ticket_id_fkey;

ALTER TABLE bookings
    ALTER COLUMN ticket_id SET NOT NULL,
    ADD CONSTRAINT bookings_ticket_id_fkey
        FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE;

-- 5. restore old indexes
CREATE INDEX idx_bookings_ticket_id ON bookings(ticket_id);
CREATE INDEX idx_bookings_status ON bookings(status);