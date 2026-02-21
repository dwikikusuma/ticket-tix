-- ==========================================
-- TICKETS
-- ==========================================
CREATE TABLE IF NOT EXISTS tickets (
                                       id SERIAL PRIMARY KEY,
                                       event_name VARCHAR(100),
                                       stadium VARCHAR(100),
                                       price INT,
                                       seat_id VARCHAR(50),
                                       status VARCHAR(20) DEFAULT 'AVAILABLE',
                                       event_date TIMESTAMP,
                                       CONSTRAINT unique_seat_event UNIQUE (seat_id, event_name)
);

CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_event_name ON tickets(event_name);
CREATE INDEX idx_tickets_event_date ON tickets(event_date);

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
-- BOOKINGS
-- ==========================================
CREATE TABLE IF NOT EXISTS bookings (
                                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                        ticket_id INT NOT NULL,
                                        status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
                                        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
                                        CONSTRAINT fk_ticket FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE
);

CREATE INDEX idx_bookings_ticket_id ON bookings(ticket_id);
CREATE INDEX idx_bookings_status ON bookings(status);