APP_NAME := asteroids
BUILD_DIR := bin

.PHONY: build run clean test fmt fmt-check lint vet

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/asteroids

run: build
	./$(BUILD_DIR)/$(APP_NAME)

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./internal/game/ -v

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files not formatted:" && gofmt -l . && exit 1)

vet:
	go vet ./...

lint: fmt-check vet
	golangci-lint run
