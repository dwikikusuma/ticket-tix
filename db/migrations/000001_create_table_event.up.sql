CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    location VARCHAR(255) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS category_types (
    name VARCHAR(50) PRIMARY KEY -- 'SEATED', 'STANDING'
);

CREATE TABLE IF NOT EXISTS event_categories (
    id SERIAL PRIMARY KEY,
    event_id INT REFERENCES events(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL, -- e.g., 'VIP', 'General Admission'
    category_type VARCHAR(50) REFERENCES category_types(name),
    price DECIMAL(12, 2) NOT NULL,
    book_type VARCHAR(20) NOT NULL, -- 'FIXED', 'FLEXIBLE'
    total_capacity INT NOT NULL,
    available_stock INT NOT NULL, -- Useful for quick checks before hitting the tickets table
    UNIQUE (event_id, name)
);

CREATE TABLE IF NOT EXISTS tickets (
    id SERIAL PRIMARY KEY,
    event_category_id INT REFERENCES event_categories(id) ON DELETE CASCADE,
    seat_number VARCHAR(20),
    status VARCHAR(20) DEFAULT 'AVAILABLE', -- 'AVAILABLE', 'RESERVED', 'SOLD'
    reserved_until TIMESTAMP,
    version INT DEFAULT 0,
    CONSTRAINT unique_seat_per_category UNIQUE (event_category_id, seat_number)
);