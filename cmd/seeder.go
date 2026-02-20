package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
)

const dbURL = "postgres://user:password@localhost:5433/ticket_tix_db?sslmode=disable"

const (
	totalEvents        = 50
	categoriesPerEvent = 3
	seatsPerCategory   = 10_000
)

var (
	eventNames = []string{
		"Rock Fest", "Jazz Night", "EDM Rave", "Classical Evening", "Pop Extravaganza",
		"Hip Hop Summit", "Country Fair", "Blues Festival", "Metal Madness", "Indie Vibes",
	}
	locations = []string{
		"Jakarta International Stadium", "Gelora Bung Karno", "ICE BSD Arena",
		"Istora Senayan", "Trans Studio Bandung", "Allianz Ecopark Ancol",
	}
	categoryDefs = []struct {
		name      string
		catType   string
		bookType  string
		priceBase int
	}{
		{"VIP", "SEATED", "FIXED", 1_500_000},
		{"Regular", "SEATED", "FIXED", 500_000},
		{"General Admission", "STANDING", "FLEXIBLE", 250_000},
	}
)

func main() {
	ctx := context.Background()

	fmt.Println("🔌 Connecting...")
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect error: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Println("🧹 Truncating tables...")
	_, err = conn.Exec(ctx, `
		TRUNCATE TABLE tickets, event_categories, events RESTART IDENTITY CASCADE;
	`)
	if err != nil {
		log.Fatalf("truncate error: %v", err)
	}

	// --- Seed category_types ---
	fmt.Println("📦 Seeding category_types...")
	_, err = conn.Exec(ctx, `
		INSERT INTO category_types (name) VALUES ('SEATED'), ('STANDING')
		ON CONFLICT DO NOTHING;
	`)
	if err != nil {
		log.Fatalf("category_types error: %v", err)
	}

	// --- Seed events ---
	fmt.Printf("🎪 Seeding %d events...\n", totalEvents)
	eventIDs := make([]int, 0, totalEvents)
	for i := 0; i < totalEvents; i++ {
		name := fmt.Sprintf("%s #%d", eventNames[rand.Intn(len(eventNames))], i+1)
		location := locations[rand.Intn(len(locations))]
		start := time.Now().AddDate(0, 0, rand.Intn(180)+1)
		end := start.Add(time.Duration(rand.Intn(4)+2) * time.Hour)

		var id int
		err = conn.QueryRow(ctx, `
			INSERT INTO events (name, description, location, start_time, end_time)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			name,
			fmt.Sprintf("This is the description for %s", name),
			location, start, end,
		).Scan(&id)
		if err != nil {
			log.Fatalf("insert event error: %v", err)
		}
		eventIDs = append(eventIDs, id)
	}

	// --- Seed event_categories ---
	fmt.Println("🗂️  Seeding event categories...")
	type categoryRecord struct {
		id       int
		bookType string
		catType  string
	}
	categoryRecords := make([]categoryRecord, 0, totalEvents*categoriesPerEvent)

	for _, eventID := range eventIDs {
		for _, cat := range categoryDefs {
			price := float64(cat.priceBase + rand.Intn(500_000))
			var id int
			err = conn.QueryRow(ctx, `
				INSERT INTO event_categories
					(event_id, name, category_type, price, book_type, total_capacity, available_stock)
				VALUES ($1, $2, $3, $4, $5, $6, $6) RETURNING id`,
				eventID, cat.name, cat.catType, price, cat.bookType, seatsPerCategory,
			).Scan(&id)
			if err != nil {
				log.Fatalf("insert category error: %v", err)
			}
			categoryRecords = append(categoryRecords, categoryRecord{
				id:       id,
				bookType: cat.bookType,
				catType:  cat.catType,
			})
		}
	}

	// --- Seed tickets via CopyFrom ---
	totalTickets := len(categoryRecords) * seatsPerCategory
	fmt.Printf("🎟️  Seeding %d tickets via COPY...\n", totalTickets)
	start := time.Now()

	idx := 0
	catIdx := 0
	seatIdx := 0

	count, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"tickets"},
		[]string{"event_category_id", "seat_number", "status"},
		pgx.CopyFromFunc(func() ([]any, error) {
			if idx >= totalTickets {
				return nil, nil
			}

			cat := categoryRecords[catIdx]

			var seatNumber *string
			if cat.bookType == "FIXED" {
				s := fmt.Sprintf("%s-%d", cat.catType[:3], seatIdx+1)
				seatNumber = &s
			}

			// Weighted status: 75% AVAILABLE, 10% RESERVED, 15% SOLD
			status := "AVAILABLE"
			r := rand.Intn(100)
			if r < 10 {
				status = "RESERVED"
			} else if r < 25 {
				status = "SOLD"
			}

			row := []any{cat.id, seatNumber, status}

			seatIdx++
			if seatIdx >= seatsPerCategory {
				seatIdx = 0
				catIdx++
			}
			idx++

			if idx%200_000 == 0 {
				fmt.Printf("   ... %d / %d tickets\n", idx, totalTickets)
			}

			return row, nil
		}),
	)
	if err != nil {
		log.Fatalf("❌ CopyFrom failed: %v", err)
	}

	fmt.Printf("✅ Inserted %d tickets in %.2f seconds.\n", count, time.Since(start).Seconds())

	// --- Update available_stock to reflect real ticket statuses ---
	fmt.Println("📊 Updating available_stock to reflect real data...")
	_, err = conn.Exec(ctx, `
		UPDATE event_categories ec
		SET available_stock = (
			SELECT COUNT(*) FROM tickets t
			WHERE t.event_category_id = ec.id
			AND t.status = 'AVAILABLE'
		)
	`)
	if err != nil {
		log.Fatalf("update available_stock error: %v", err)
	}

	fmt.Println("\n🎉 All done! Summary:")
	fmt.Printf("   • category_types : 2\n")
	fmt.Printf("   • events         : %d\n", totalEvents)
	fmt.Printf("   • event_categories: %d\n", len(categoryRecords))
	fmt.Printf("   • tickets        : %d\n", count)
}
