APP_NAME := asteroids
BUILD_DIR := bin
WEIGHTS := weights.gob

.PHONY: build build-train build-watch run clean test fmt fmt-check lint vet train watch

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/asteroids

build-train:
	go build -o $(BUILD_DIR)/train ./cmd/train

build-watch:
	go build -o $(BUILD_DIR)/watch ./cmd/watch

run: build
	./$(BUILD_DIR)/$(APP_NAME)

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./internal/game/ ./internal/ai/ -v

train: build-train
	./$(BUILD_DIR)/train -output $(WEIGHTS)

watch: build-watch
	./$(BUILD_DIR)/watch $(WEIGHTS)

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files not formatted:" && gofmt -l . && exit 1)

vet:
	go vet ./...

lint: fmt-check vet
	golangci-lint run
