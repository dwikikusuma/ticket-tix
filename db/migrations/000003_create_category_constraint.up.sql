ALTER TABLE event_categories
ADD CONSTRAINT chk_category_type
CHECK (
    NOT (category_type = 'STANDING' AND book_type = 'FIXED')
    );