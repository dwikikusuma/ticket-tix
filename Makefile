DB_URL=postgres://user:password@localhost:5433/ticket_tix_db?sslmode=disable
MIGRATIONS_PATH=$(CURDIR)/db/migrations

.PHONY: migrate-up migrate-down migrate-force

migrate-up:
	docker run --rm -v $(MIGRATIONS_PATH):/migrations --network host migrate/migrate \
	   -path=/migrations/ -database "$(DB_URL)" up

migrate-down:
	docker run --rm -v $(MIGRATIONS_PATH):/migrations --network host migrate/migrate \
	   -path=/migrations/ -database "$(DB_URL)" down 1

migrate-force:
	docker run --rm -v $(MIGRATIONS_PATH):/migrations --network host migrate/migrate \
	   -path=/migrations/ -database "$(DB_URL)" force 1