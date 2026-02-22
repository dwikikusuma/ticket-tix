-- name: InsertEvent :one
INSERT INTO events (name, description, location, start_time, end_time)
VALUES ($1, $2, $3, $4, $5)
    RETURNING *;

-- name: InsertEventImage :one
INSERT INTO event_images (event_id, image_key, is_primary, display_order)
VALUES ($1, $2, $3, $4)
    RETURNING *;

-- name: InsertEventCategory :one
INSERT INTO event_categories (event_id, name, category_type, price, book_type, total_capacity, available_stock)
VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING *;

-- name: InsertTicket :one
INSERT INTO tickets (event_category_id, seat_number, status, reserved_until)
VALUES ($1, $2, $3, $4)
    RETURNING *;

-- name: InsertBooking :one
INSERT INTO bookings (ticket_id, status)
VALUES ($1, $2)
    RETURNING *;