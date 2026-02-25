-- ==========================================
-- MIGRATION DOWN: revert bookings table
-- ==========================================

-- 1. drop new indexes
DROP INDEX IF EXISTS idx_bookings_user_id;
DROP INDEX IF EXISTS idx_bookings_event_id;
DROP INDEX IF EXISTS idx_bookings_event_category_id;
DROP INDEX IF EXISTS idx_bookings_ticket_id;
DROP INDEX IF EXISTS idx_bookings_status;

-- 2. drop new columns
ALTER TABLE bookings
    DROP COLUMN user_id,
    DROP COLUMN event_id,
    DROP COLUMN event_category_id;

-- 3. drop unique constraint on ticket_id
ALTER TABLE bookings
    DROP CONSTRAINT unique_ticket_booking;

-- 4. revert ticket_id NOT NULL
ALTER TABLE bookings
    ALTER COLUMN ticket_id SET NOT NULL;

-- 5. restore old indexes
CREATE INDEX idx_bookings_ticket_id ON bookings(ticket_id);
CREATE INDEX idx_bookings_status ON bookings(status);