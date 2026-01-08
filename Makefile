dev:
	@/home/mmnb/go/bin/reflex -r '.go' -s -- go run main.go

dep_install:
	go mod tidy

run:
	docker compose build
	docker compose up