-- name: CreateBooking :one
INSERT INTO bookings (ticket_id, status, user_id, event_category_id, event_id)
VALUES ($1, $2, $3, $4, $5)
    RETURNING *;