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
