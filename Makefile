ROOT=$(shell pwd)
BUILD_DIR=$(ROOT)/build
OUTBIN=$(BUILD_DIR)/main
MAIN=$(ROOT)/cmd/main.go

build: pre_build
	CGO_ENABLED=0 GOOS=linux go build -o $(OUTBIN) $(MAIN)

pre_build:
	mkdir -p $(ROOT)/build

watch:
	while true; do \
		make build; \
		inotifywait -qre close_write .; \
	done

test_lab:
	@echo "not implemented"
