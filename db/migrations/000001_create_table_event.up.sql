-- ==========================================
-- CATEGORY TYPES
-- ==========================================
CREATE TABLE IF NOT EXISTS category_types (
                                              name VARCHAR(50) PRIMARY KEY -- 'SEATED', 'STANDING'
);

INSERT INTO category_types (name) VALUES ('SEATED'), ('STANDING');

-- ==========================================
-- EVENTS
-- ==========================================
CREATE TABLE IF NOT EXISTS events (
                                      id SERIAL PRIMARY KEY,
                                      name VARCHAR(100) NOT NULL,
                                      description TEXT,
                                      location VARCHAR(255) NOT NULL,
                                      start_time TIMESTAMP NOT NULL,
                                      end_time TIMESTAMP NOT NULL,
                                      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_name ON events(name);
CREATE INDEX idx_events_location ON events(location);
CREATE INDEX idx_events_start_time ON events(start_time);

-- ==========================================
-- EVENT IMAGES
-- ==========================================
CREATE TABLE IF NOT EXISTS event_images (
                                            id SERIAL PRIMARY KEY,
                                            event_id INT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
                                            image_key VARCHAR(500) NOT NULL,
                                            is_primary BOOLEAN DEFAULT FALSE,
                                            display_order INT DEFAULT 0,
                                            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_event_images_event_id ON event_images(event_id);
CREATE INDEX idx_event_images_display_order ON event_images(event_id, display_order);
CREATE UNIQUE INDEX idx_one_primary_per_event ON event_images(event_id) WHERE is_primary = TRUE;

-- ==========================================
-- EVENT CATEGORIES
-- ==========================================
CREATE TABLE IF NOT EXISTS event_categories (
                                                id SERIAL PRIMARY KEY,
                                                event_id INT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
                                                name VARCHAR(50) NOT NULL,
                                                category_type VARCHAR(50) REFERENCES category_types(name),
                                                price DECIMAL(12, 2) NOT NULL,
                                                book_type VARCHAR(20) NOT NULL,         -- 'FIXED', 'FLEXIBLE'
                                                total_capacity INT NOT NULL,
                                                available_stock INT NOT NULL,
                                                UNIQUE (event_id, name)
);

CREATE INDEX idx_event_categories_event_id ON event_categories(event_id);
CREATE INDEX idx_event_categories_available_stock ON event_categories(available_stock);

-- ==========================================
-- TICKETS
-- ==========================================
CREATE TABLE IF NOT EXISTS tickets (
                                       id SERIAL PRIMARY KEY,
                                       event_category_id INT NOT NULL REFERENCES event_categories(id) ON DELETE CASCADE,
                                       seat_number VARCHAR(20),
                                       status VARCHAR(20) DEFAULT 'AVAILABLE', -- 'AVAILABLE', 'RESERVED', 'SOLD'
                                       reserved_until TIMESTAMP,
                                       version INT DEFAULT 0,
                                       CONSTRAINT unique_seat_per_category UNIQUE (event_category_id, seat_number)
);

CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_event_category_id ON tickets(event_category_id);

-- ==========================================
-- BOOKINGS
-- ==========================================
CREATE TABLE IF NOT EXISTS bookings (
                                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                        ticket_id INT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
                                        status VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- 'PENDING', 'CONFIRMED', 'CANCELLED'
                                        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_bookings_ticket_id ON bookings(ticket_id);
CREATE INDEX idx_bookings_status ON bookings(status);