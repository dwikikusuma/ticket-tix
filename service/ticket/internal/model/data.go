package model

import (
	"io"
	"time"
)

type EventData struct {
	ID          int32
	Name        string
	Description string
	Location    string
	StartTime   time.Time
	EndTime     time.Time
}

type FileData struct {
	Filename     string
	Size         int64
	ContentType  string
	Reader       io.Reader
	IsPrimary    bool
	DisplayOrder int
}

type InsertTicketRequest struct {
	Event EventData
	Files []FileData
}

type EventCategoryData struct {
	EventID           int32
	CategoryID        int32
	CategoryType      string
	Price             float64
	BookType          string
	TotalCapacity     int32
	AvailableCapacity int32
}

type EventImageData struct {
	EventID      int32
	Key          string
	IsPrimary    bool
	DisplayOrder int
}

type EventDetailsData struct {
	EventData
	Images     []EventImageData
	Categories []EventCategoryData
}
