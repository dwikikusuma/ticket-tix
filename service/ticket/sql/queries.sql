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

-- name: DeleteEventImage :exec
DELETE FROM event_images
WHERE image_key = $1;

-- name: GetEventDetails :one
SELECT * FROM events
WHERE id = $1;

-- name: GetEventCategories :many
SELECT * FROM event_categories
WHERE event_id = $1;

-- name: GetEventImages :many
SELECT * FROM event_images
WHERE event_id = $1
ORDER BY display_order;

-- name: BrowseEvents :many
SELECT e.id, e.name, e.description, e.location, e.start_time, e.end_time, e.created_at, ei.image_key
FROM events e
JOIN event_images ei ON e.id = ei.event_id AND ei.is_primary = true
WHERE
    (sqlc.arg(event_name)::text = '' OR e.name ILIKE '%' || sqlc.arg(event_name) || '%') AND
    (sqlc.arg(location)::text = '' OR e.location ILIKE '%' || sqlc.arg(location) || '%') AND
    (sqlc.arg(start_date)::timestamp = '0001-01-01 00:00:00' OR e.start_time >= sqlc.arg(start_date)) AND
    (sqlc.arg(end_date)::timestamp = '0001-01-01 00:00:00' OR e.start_time <= sqlc.arg(end_date)) AND
    (
        sqlc.arg(cursor_time)::timestamp = '0001-01-01 00:00:00' OR
        e.start_time > sqlc.arg(cursor_time)::timestamp OR
        (e.start_time = sqlc.arg(cursor_time)::timestamp AND e.id > sqlc.arg(cursor_id)::int)
        )
ORDER BY e.start_time ASC, e.id ASC
    LIMIT sqlc.arg(page_size);