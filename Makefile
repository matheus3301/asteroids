APP_NAME := asteroids
BUILD_DIR := bin

.PHONY: build run clean

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/asteroids

run: build
	./$(BUILD_DIR)/$(APP_NAME)

clean:
	rm -rf $(BUILD_DIR)
