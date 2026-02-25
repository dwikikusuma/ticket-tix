-- ==========================================
-- MIGRATION: update bookings table
-- ==========================================

-- 1. drop old indexes
DROP INDEX IF EXISTS idx_bookings_ticket_id;
DROP INDEX IF EXISTS idx_bookings_status;

-- 2. drop old FK constraint
ALTER TABLE bookings
    DROP CONSTRAINT bookings_ticket_id_fkey;

-- 3. modify ticket_id (NOT NULL -> nullable, no CASCADE)
ALTER TABLE bookings
    ALTER COLUMN ticket_id DROP NOT NULL,
    ADD CONSTRAINT bookings_ticket_id_fkey
        FOREIGN KEY (ticket_id) REFERENCES tickets(id);

-- 4. add unique on ticket_id
ALTER TABLE bookings
    ADD CONSTRAINT unique_ticket_booking UNIQUE (ticket_id);

-- 5. add new columns
ALTER TABLE bookings
    ADD COLUMN user_id INT NOT NULL,
    ADD COLUMN event_category_id INT NOT NULL REFERENCES event_categories(id);

-- 6. rebuild indexes
CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_event_category_id ON bookings(event_category_id);
CREATE INDEX idx_bookings_ticket_id ON bookings(ticket_id);
CREATE INDEX idx_bookings_status ON bookings(status);