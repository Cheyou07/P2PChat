# variables
BIN_NAME := myapp
BUILD_DIR := build
SRC_DIR := cmd
SRC := $(wildcard $(SRC_DIR)/*.go)
FLAGS := 

default: run

build:
	go build -o $(BUILD_DIR)/$(BIN_NAME) $(SRC) $(FLAGS)

run: build
	./$(BUILD_DIR)/$(BIN_NAME)

clean:
	rm $(BUILD_DIR)/*

.PHONY: run build clean
