package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"ticket-tix/service/ticket/internal/infra/redis"
	"ticket-tix/service/ticket/internal/model"
	"time"
)

type StockSyncJob struct {
	counter *redis.StockCounter
	repo    model.TicketRepo
}

func NewStockSyncJob(counter *redis.StockCounter, repo model.TicketRepo) *StockSyncJob {
	return &StockSyncJob{
		counter: counter,
		repo:    repo,
	}
}

func (s *StockSyncJob) SeedAll(ctx context.Context) error {
	categories, err := s.repo.GetAllStandingEventCatStock(ctx)
	if err != nil {
		return fmt.Errorf("get standing categories: %w", err)
	}

	for _, cat := range categories {
		if err := s.counter.Seed(ctx, cat.EventCatID, cat.Stock); err != nil {
			return fmt.Errorf("seed category %d: %w", cat.EventCatID, err)
		}
	}
	return nil
}

func (s *StockSyncJob) SyncEventStockJob(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("context canceled, exiting syncEventStockJob")
			return
		case <-ticker.C:
			log.Println("starting syncEventStockJob")

			jobs, err := s.fetchEventToSync(ctx)
			if err != nil {
				log.Printf("failed to fetch events to sync: %v", err)
				continue
			}
			s.runWorkerPool(ctx, jobs, 5)

			log.Println("completed syncEventStockJob")
		}
	}
}

func (s *StockSyncJob) fetchEventToSync(ctx context.Context) ([]int64, error) {
	var res []int64
	events, err := s.repo.GetAllStandingEventCatStock(ctx)
	if err != nil {
		return res, err
	}

	for _, cat := range events {
		res = append(res, int64(cat.EventCatID))
	}
	return res, nil
}

func (s *StockSyncJob) runWorkerPool(ctx context.Context, jobs []int64, numWorker int32) {
	var wg sync.WaitGroup
	jobChan := make(chan int64, numWorker)

	for i := 0; i < int(numWorker); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				select {
				case <-ctx.Done():
					log.Println("context canceled, worker exiting")
					return
				default:
					s.processSyncEventStock(ctx, job)
				}
			}
		}()
	}

	for _, job := range jobs {
		jobChan <- job
	}
	close(jobChan)
	wg.Wait()
}

func (s *StockSyncJob) processSyncEventStock(ctx context.Context, eventID int64) {
	stock, err := s.counter.Get(ctx, int32(eventID))
	if err != nil {
		log.Printf("failed to get stock for event category %d: %v", eventID, err)
		return
	}

	if updateErr := s.repo.UpdateEventCategoryStock(ctx, int32(eventID), stock); updateErr != nil {
		log.Printf("failed to update stock for event category %d: %v", eventID, updateErr)
		return
	}
}
