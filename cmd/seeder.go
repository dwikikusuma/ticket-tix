package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const dbURL = "postgres://user:password@localhost:5433/ticket_tix_db?sslmode=disable"

const (
	totalEvents        = 50
	categoriesPerEvent = 3
	seatsPerCategory   = 10_000
	imagesPerEvent     = 3 // 1 primary + 2 gallery
)

const (
	minioEndpoint  = "localhost:9000"
	minioAccessKey = "minioadmin"
	minioSecretKey = "minioadmin123"
	minioBucket    = "ticket-bucket"
	minioUseSSL    = false
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

	// Random vibrant colors for dummy images so they're visually distinguishable
	imageColors = []color.RGBA{
		{220, 50, 50, 255},  // red
		{50, 120, 220, 255}, // blue
		{50, 180, 80, 255},  // green
		{220, 150, 30, 255}, // orange
		{150, 50, 220, 255}, // purple
		{30, 180, 180, 255}, // teal
		{220, 50, 150, 255}, // pink
		{100, 80, 60, 255},  // brown
	}
)

func main() {
	ctx := context.Background()

	// --- Connect Postgres ---
	fmt.Println("🔌 Connecting to Postgres...")
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect error: %v", err)
	}
	defer conn.Close(ctx)

	// --- Connect MinIO ---
	fmt.Println("🪣  Connecting to MinIO...")
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: minioUseSSL,
	})
	if err != nil {
		log.Fatalf("minio connect error: %v", err)
	}

	// Ensure bucket exists
	exists, err := minioClient.BucketExists(ctx, minioBucket)
	if err != nil {
		log.Fatalf("bucket check error: %v", err)
	}
	if !exists {
		if err = minioClient.MakeBucket(ctx, minioBucket, minio.MakeBucketOptions{}); err != nil {
			log.Fatalf("bucket create error: %v", err)
		}
		fmt.Printf("   Bucket '%s' created.\n", minioBucket)
	} else {
		fmt.Printf("   Bucket '%s' already exists.\n", minioBucket)
	}

	// --- Truncate ---
	fmt.Println("🧹 Truncating tables...")
	_, err = conn.Exec(ctx, `
		TRUNCATE TABLE tickets, event_categories, event_images, events RESTART IDENTITY CASCADE;
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

	// --- Seed events + images ---
	fmt.Printf("🎪 Seeding %d events with images...\n", totalEvents)
	eventIDs := make([]int, 0, totalEvents)

	for i := 0; i < totalEvents; i++ {
		name := fmt.Sprintf("%s #%d", eventNames[rand.Intn(len(eventNames))], i+1)
		location := locations[rand.Intn(len(locations))]
		start := time.Now().AddDate(0, 0, rand.Intn(180)+1)
		end := start.Add(time.Duration(rand.Intn(4)+2) * time.Hour)

		var eventID int
		err = conn.QueryRow(ctx, `
			INSERT INTO events (name, description, location, start_time, end_time)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			name,
			fmt.Sprintf("This is the description for %s", name),
			location, start, end,
		).Scan(&eventID)
		if err != nil {
			log.Fatalf("insert event error: %v", err)
		}
		eventIDs = append(eventIDs, eventID)

		// Upload images to MinIO and insert into event_images
		for imgIdx := 0; imgIdx < imagesPerEvent; imgIdx++ {
			isPrimary := imgIdx == 0
			imageLabel := "gallery"
			if isPrimary {
				imageLabel = "banner"
			}

			// Generate a dummy colored JPEG image
			imgColor := imageColors[rand.Intn(len(imageColors))]
			imgBytes, err := generateDummyImage(800, 600, imgColor)
			if err != nil {
				log.Fatalf("generate image error: %v", err)
			}

			// Upload to MinIO
			// key format: events/{eventID}/{banner|gallery-N}.jpg
			key := fmt.Sprintf("events/%d/%s.jpg", eventID, imageLabel)
			if !isPrimary {
				key = fmt.Sprintf("events/%d/gallery-%d.jpg", eventID, imgIdx)
			}

			_, err = minioClient.PutObject(
				ctx,
				minioBucket,
				key,
				bytes.NewReader(imgBytes),
				int64(len(imgBytes)),
				minio.PutObjectOptions{ContentType: "image/jpeg"},
			)
			if err != nil {
				log.Fatalf("minio upload error: %v", err)
			}

			// Insert into event_images
			_, err = conn.Exec(ctx, `
				INSERT INTO event_images (event_id, image_key, is_primary, display_order)
				VALUES ($1, $2, $3, $4)`,
				eventID, key, isPrimary, imgIdx,
			)
			if err != nil {
				log.Fatalf("insert event_image error: %v", err)
			}
		}

		fmt.Printf("   Event %d/%d seeded with %d images\n", i+1, totalEvents, imagesPerEvent)
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
	ticketStart := time.Now()

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

	fmt.Printf("✅ Inserted %d tickets in %.2f seconds.\n", count, time.Since(ticketStart).Seconds())

	// --- Update available_stock ---
	fmt.Println("📊 Updating available_stock...")
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
	fmt.Printf("   • category_types  : 2\n")
	fmt.Printf("   • events          : %d\n", totalEvents)
	fmt.Printf("   • event_images    : %d (in MinIO + DB)\n", totalEvents*imagesPerEvent)
	fmt.Printf("   • event_categories: %d\n", len(categoryRecords))
	fmt.Printf("   • tickets         : %d\n", count)
}

// generateDummyImage creates a simple solid-color JPEG as a placeholder
func generateDummyImage(width, height int, c color.RGBA) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Add a slight gradient so images aren't completely flat
			r := uint8(min(int(c.R)+x*20/width, 255))
			g := uint8(min(int(c.G)+y*20/height, 255))
			b := uint8(min(int(c.B), 255))
			img.SetRGBA(x, y, color.RGBA{r, g, b, 255})
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
