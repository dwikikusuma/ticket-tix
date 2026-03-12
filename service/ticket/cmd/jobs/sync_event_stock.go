package jobs

import (
	"context"
	"log"
	"sync"
	"ticket-tix/service/ticket/internal/infra/redis"
	"ticket-tix/service/ticket/internal/model"
	"time"
)

func syncEventCatStock(ctx context.Context, counter *redis.StockCounter, repo model.TicketRepo) {
	categoriesData, err := repo.GetAllStandingEventCatStock(ctx)
	if err != nil {
		log.Fatalf("failed to get all standing event category stock: %v", err)
	}

	for _, cat := range categoriesData {
		if seedErr := counter.Seed(ctx, cat.EventCatID, cat.Stock); seedErr != nil {
			log.Printf("failed to seed stock for event category %d: %v", cat.EventCatID, seedErr)
			continue
		}
	}
}

func syncEventStockJob(ctx context.Context, counter *redis.StockCounter, repo model.TicketRepo) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("context canceled, exiting syncEventStockJob")
			return
		case <-ticker.C:
			log.Println("starting syncEventStockJob")

			jobs, err := fetchEventToSync(ctx, repo)
			if err != nil {
				log.Printf("failed to fetch events to sync: %v", err)
				continue
			}

			log.Println("completed syncEventStockJob")

		}
	}
}

func fetchEventToSync(ctx context.Context, repo model.TicketRepo) ([]int64, error) {
	var res []int64
	events, err := repo.GetAllStandingEventCatStock(ctx)
	if err != nil {
		log.Printf("failed to get all standing event category stock: %v", err)
		return res, err
	}

	for _, cat := range events {
		res = append(res, int64(cat.EventCatID))
	}
	return res, nil
}

func runWorkerPool(jobs []int64, numWorker int32) {
	var wg *sync.WaitGroup
	jobChan := make(chan int64, numWorker)

}
