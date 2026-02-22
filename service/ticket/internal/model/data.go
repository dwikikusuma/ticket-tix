package model

import (
	"io"
	"time"
)

type EventData struct {
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
