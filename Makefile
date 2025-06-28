# Config
GOOS := js
GOARCH := wasm
DIST := httpserve/dist
WASM_OUT := $(DIST)/main.wasm
SERVER_SRC := ./cmd/server
HOST := localhost:8080

WASM_EXEC := $(shell go env GOROOT)/lib/wasm/wasm_exec.js 
WASM_EXEC_DIST := $(DIST)/wasm_exec.js

.PHONY: all build copy serve run open clean test lint

all: build copy

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(WASM_OUT) .

copy:
	echo "Using local wasm_exec.js from Go toolchain"; \
	cp $(WASM_EXEC) $(WASM_EXEC_DIST); \

runbrowser: all
	@echo "Starting server..."
	@go run $(SERVER_SRC) -host=$(HOST) & \
	SERVER_PID=$$!; \
	for i in $$(seq 1 30); do \
		curl -sSf http://$(HOST) >/dev/null && break; \
		echo "Waiting for server..."; \
		sleep 0.2; \
	done; \
	( \
		xdg-open http://$(HOST)/ 2>/dev/null || \
		open http://$(HOST)/ 2>/dev/null || \
		start http://$(HOST)/ 2>/dev/null || \
		echo "Please open http://$(HOST)/ manually" \
	); \
	wait $$SERVER_PID

clean:
	rm -f $(WASM_OUT) $(WASM_EXEC_DIST)

test: lint
	go test -v -cover ./...

lint:
	golangci-lint run ./...