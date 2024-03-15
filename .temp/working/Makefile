# variables
BIN_NAME := myapp
BUILD_DIR := build
SRC_DIR := cmd
SRC := $(wildcard $(SRC_DIR)/*.go)
FLAGS := 

default: run

# Build
build:
	go build -o $(BUILD_DIR)/$(BIN_NAME) $(SRC) $(FLAGS)

run: build
	./$(BUILD_DIR)/$(BIN_NAME)

clean:
	rm -rf $(BUILD_DIR)

.PHONY: run build clean
