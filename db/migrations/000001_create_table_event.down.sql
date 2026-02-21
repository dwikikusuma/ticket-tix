-- Drop in reverse dependency order
DROP INDEX IF EXISTS idx_bookings_status;
DROP INDEX IF EXISTS idx_bookings_ticket_id;
DROP TABLE IF EXISTS bookings;

DROP INDEX IF EXISTS idx_one_primary_per_event;
DROP INDEX IF EXISTS idx_event_images_display_order;
DROP INDEX IF EXISTS idx_event_images_event_id;
DROP TABLE IF EXISTS event_images;

DROP INDEX IF EXISTS idx_events_start_time;
DROP INDEX IF EXISTS idx_events_location;
DROP INDEX IF EXISTS idx_events_name;
DROP TABLE IF EXISTS events;

DROP INDEX IF EXISTS idx_tickets_event_date;
DROP INDEX IF EXISTS idx_tickets_event_name;
DROP INDEX IF EXISTS idx_tickets_status;
DROP TABLE IF EXISTS tickets;