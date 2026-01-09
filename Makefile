.PHONY: wire run run-worker build dev dev-worker generate hot

# Generate wire_gen.go
wire:
	wire gen ./cmd/app ./cmd/worker ./cmd/scheduler

# Go generate (buat jalankan semua go:generate, termasuk wire)
generate:
	go generate ./...

# Run app
run:
	go run ./cmd/app

# Run worker
run-worker:
	go run ./cmd/worker

# Run scheduler
run-scheduler:
	go run ./cmd/scheduler

# Build app binary
build:
	go build -o myapp ./cmd/app

# Build worker binary
build-worker:
	go build -o myworker ./cmd/worker

# Build scheduler binary
build-scheduler:
	go build -o myscheduler ./cmd/scheduler

# Dev mode: auto generate wire, then run app
dev:
	make generate
	make run

# Dev mode for worker
dev-worker:
	make generate
	make run-worker

# Dev mode for scheduler
dev-scheduler:
	make generate
	make run-scheduler

# Hot reload mode with air
hot:
	air