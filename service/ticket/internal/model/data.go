package model

import (
	"io"
	"time"
)

type EventData struct {
	ID          int32     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	ImageURL    string    `json:"image_url,omitempty"`
}

type FileData struct {
	Filename     string    `json:"filename"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"contentType"`
	Reader       io.Reader `json:"reader"`
	IsPrimary    bool      `json:"is_primary"`
	DisplayOrder int       `json:"display_order"`
}

type InsertTicketRequest struct {
	Event EventData  `json:"event"`
	Files []FileData `json:"files"`
}

type EventCategoryData struct {
	EventID           int32   `json:"event_id"`
	CategoryID        int32   `json:"category_id"`
	CategoryType      string  `json:"category_type"`
	Price             float64 `json:"price"`
	BookType          string  `json:"book_type"`
	TotalCapacity     int32   `json:"total_capacity"`
	AvailableCapacity int32   `json:"available_capacity"`
}

type EventImageData struct {
	EventID      int32  `json:"event_id"`
	Key          string `json:"key"`
	IsPrimary    bool   `json:"is_primary"`
	DisplayOrder int    `json:"display_order"`
}

type EventDetailsData struct {
	EventData
	Images     []EventImageData    `json:"images"`
	Categories []EventCategoryData `json:"categories"`
}

type BrowseCursor struct {
	StartTime time.Time `json:"start_time"`
	ID        int32     `json:"id"`
}

type BrowseFilter struct {
	EventName string        `json:"event_name"`
	Location  string        `json:"location"`
	StartDate time.Time     `json:"start_date"`
	EndDate   time.Time     `json:"end_date"`
	Cursor    *BrowseCursor `json:"cursor"`
	Limit     int           `json:"limit"`
}

type BrowseResult struct {
	Events     []EventData `json:"events"`
	NextCursor *string     `json:"next_cursor"`
	HasMore    bool        `json:"has_more"`
}
