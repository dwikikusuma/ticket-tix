-- name: InsertEvents :one
INSERT INTO events (name, description, location, start_time, end_time, created_at)
VALUES ($1,$2, $3, $4, $5, $6)
RETURNING *;

-- name: InsertEventsCategory :one
INSERT INTO event_categories (
    id, event_id, category_type, price,
    book_type, total_capacity, available_stock
) VALUES (
 $1,$2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: InsertTickets :one
INSERT INTO tickets (event_category_id, seat_number, status, reserved_until, version)
VALUES ($1,$2, $3, $4, $5)
RETURNING *;

-- name: GetEventDetailByID :one
SELECT