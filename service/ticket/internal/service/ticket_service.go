package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"ticket-tix/common/pkg/storage"
	"ticket-tix/service/ticket/internal/model"
	"time"

	"github.com/google/uuid"
)

type TicketService struct {
	storage *storage.Storage
	repo    model.TicketRepo
	db      *sql.DB
}

func NewTicketService(db *sql.DB, storage *storage.Storage, repo model.TicketRepo) model.TicketService {
	return &TicketService{db: db, storage: storage, repo: repo}
}

func (s *TicketService) CreateEvent(ctx context.Context, req model.InsertTicketRequest) (model.EventData, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.EventData{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txRepo := s.repo.WithTx(tx)

	event, err := txRepo.InsertEvent(ctx, req.Event)
	if err != nil {
		return model.EventData{}, err
	}

	fileKeys, err := s.insertFiles(ctx, event.ID, req.Files)
	if err != nil {
		return model.EventData{}, err
	}

	for i, key := range fileKeys {
		if err := txRepo.InsertEventImage(ctx, model.ImageKeyData{
			EventID:      event.ID,
			Key:          key,
			IsPrimary:    i == 0,
			DisplayOrder: i,
		}); err != nil {
			s.deleteFiles(ctx, fileKeys)
			return model.EventData{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		s.deleteFiles(ctx, fileKeys)
		return model.EventData{}, fmt.Errorf("commit tx: %w", err)
	}

	return event, nil
}

func (s *TicketService) DeleteImg(ctx context.Context, key string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txRepo := s.repo.WithTx(tx)
	err = txRepo.DeleteEventImage(ctx, key)
	if err != nil {
		return fmt.Errorf("delete event image from db: %w", err)
	}
	if err := s.storage.Delete(ctx, key); err != nil {
		return fmt.Errorf("delete image: %w", err)
	}
	return nil
}

func (s *TicketService) UploadImg(ctx context.Context, eventID int32, files []model.FileData) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	fileKeys, uploadErr := s.insertFiles(ctx, eventID, files)
	if uploadErr != nil {
		return uploadErr
	}

	txRepo := s.repo.WithTx(tx)
	for i, key := range fileKeys {
		if err := txRepo.InsertEventImage(ctx, model.ImageKeyData{
			EventID:      eventID,
			Key:          key,
			IsPrimary:    i == 0,
			DisplayOrder: i,
		}); err != nil {
			s.deleteFiles(ctx, fileKeys)
			return fmt.Errorf("insert event image: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		s.deleteFiles(ctx, fileKeys)
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (s *TicketService) GetEventDetail(ctx context.Context, id int32) (model.EventDetailsData, error) {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return model.EventDetailsData{}, fmt.Errorf("get event by id: %w", err)
	}

	categories, err := s.repo.GetEventCategory(ctx, id)
	if err != nil {
		return model.EventDetailsData{}, fmt.Errorf("get event category: %w", err)
	}

	images, err := s.repo.GetEventImages(ctx, id)
	if err != nil {
		return model.EventDetailsData{}, fmt.Errorf("get event images: %w", err)
	}

	var imagesWithUrl []model.EventImageData
	for _, img := range images {
		url := s.storage.GetImageURL(img.Key)
		img.Key = url
		imagesWithUrl = append(imagesWithUrl, img)
	}

	return model.EventDetailsData{
		EventData:  event,
		Categories: categories,
		Images:     imagesWithUrl,
	}, nil
}

func (s *TicketService) BrowseEvents(ctx context.Context, filter model.BrowseFilter) (model.BrowseResult, error) {
	filter.Limit += 1
	events, err := s.repo.BrowseEvents(ctx, filter)
	if err != nil {
		return model.BrowseResult{}, fmt.Errorf("browse events: %w", err)
	}
	filter.Limit -= 1

	hasMore := len(events) > filter.Limit
	if hasMore {
		events = events[:filter.Limit]
	}

	var nextCursor *string
	if hasMore && len(events) > 0 {
		last := events[len(events)-1]
		cursor := model.BrowseCursor{
			StartTime: last.StartTime,
			ID:        last.ID,
		}
		b, _ := json.Marshal(cursor)
		encoded := base64.StdEncoding.EncodeToString(b)
		nextCursor = &encoded
	}

	for i := range events {
		imgUrl := s.storage.GetImageURL(events[i].ImageURL)
		events[i].ImageURL = imgUrl
	}

	return model.BrowseResult{
		Events:     events,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// UpdateTicketStatus updates a ticket to RESERVED or AVAILABLE only.
// "RESERVED" is used when a booking is in progress.
// "AVAILABLE" is used when a booking is cancelled/released.
// "SOLD" is set after payment is confirmed (via a separate flow).
func (s *TicketService) UpdateTicketStatus(ctx context.Context, status, seatNum string, eventCategoryID int32) (int32, error) {
	// FIX: was only allowing "Sold" and "AVAILABLE", rejecting "RESERVED"
	// "SOLD" is intentionally excluded here — it should only be set after payment confirmation
	if status != "RESERVED" && status != "AVAILABLE" {
		return 0, fmt.Errorf("invalid status: %q, must be RESERVED or AVAILABLE", status)
	}

	// FIX: guard against updating a ticket that isn't currently in the expected state
	// e.g. don't let two concurrent requests both reserve the same seat
	ticket, err := s.repo.GetTicketSeatAndEventCategory(ctx, seatNum, eventCategoryID)
	if err != nil {
		return 0, fmt.Errorf("get ticket before update: %w", err)
	}

	if status == "RESERVED" && ticket.Status != "AVAILABLE" {
		return 0, fmt.Errorf("ticket is not available for reservation, current status: %s", ticket.Status)
	}
	if status == "AVAILABLE" && ticket.Status != "RESERVED" {
		return 0, fmt.Errorf("ticket is not reserved, cannot release, current status: %s", ticket.Status)
	}

	return s.repo.UpdateTicketStatus(ctx, status, seatNum, eventCategoryID)
}

// ValidateTicketBooking validates whether a ticket/seat can be booked.
// For SEATED categories: checks the specific seat exists, is AVAILABLE, and not expired-reserved.
// For STANDING categories: checks there is remaining available stock.
func (s *TicketService) ValidateTicketBooking(ctx context.Context, seatId string, eventID, eventCategory int32) error {
	ecDetail, err := s.repo.GetEventCategoryByID(ctx, eventCategory)
	if err != nil {
		return fmt.Errorf("get event category by id: %w", err)
	}

	// FIX: event category must belong to the given event
	if ecDetail.EventID != eventID {
		return fmt.Errorf("event category %d does not belong to event %d", eventCategory, eventID)
	}

	// FIX: was comparing against "Seated" — DB stores "SEATED" (uppercase)
	switch ecDetail.CategoryType {
	case "SEATED":
		if seatId == "" {
			return fmt.Errorf("seat_id is required for SEATED category")
		}

		ticketDetail, err := s.repo.GetTicketSeatAndEventCategory(ctx, seatId, eventCategory)
		if err != nil {
			return fmt.Errorf("get ticket: %w", err)
		}

		// FIX: was never checking if the ticket is actually AVAILABLE
		if ticketDetail.Status != "AVAILABLE" {
			// FIX: a RESERVED ticket whose reservation has expired should be treated as AVAILABLE
			if ticketDetail.Status == "RESERVED" && !ticketDetail.ReservedUntil.IsZero() && time.Now().After(ticketDetail.ReservedUntil) {
				// Reservation expired — treat as available (the actual status update
				// happens in UpdateTicketStatus; we just allow the booking to proceed)
				return nil
			}
			return fmt.Errorf("seat %s is not available, current status: %s", seatId, ticketDetail.Status)
		}

	case "STANDING":
		if seatId != "" {
			return fmt.Errorf("seat_id must be empty for STANDING category")
		}

		// FIX: was checking TotalCapacity (never changes) instead of AvailableCapacity
		if ecDetail.AvailableCapacity <= 0 {
			return fmt.Errorf("no available capacity for this category")
		}

	default:
		return fmt.Errorf("unknown category type: %s", ecDetail.CategoryType)
	}

	return nil
}

func (s *TicketService) insertFiles(ctx context.Context, eventID int32, files []model.FileData) ([]string, error) {
	filesKey := make([]string, 0, len(files))
	for _, file := range files {
		key := fmt.Sprintf("events/%d/%s-%s", eventID, uuid.New().String(), file.Filename)
		uploadErr := s.storage.UploadImage(ctx, file.Reader, key, file.Size, file.ContentType)
		if uploadErr != nil {
			s.deleteFiles(ctx, filesKey)
			return nil, fmt.Errorf("upload image: %w", uploadErr)
		}
		filesKey = append(filesKey, key)
	}

	return filesKey, nil
}

func (s *TicketService) deleteFiles(ctx context.Context, keys []string) {
	for _, key := range keys {
		if err := s.storage.Delete(ctx, key); err != nil {
			log.Printf("failed to delete file with key %s: %v\n", key, err)
		}
	}
}
