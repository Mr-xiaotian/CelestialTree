# ---------- basic ----------

OUT := bin

NOW_NAME        := now
APP_NAME        := celestialtree
BENCH_HTTP_NAME := bench_emit_http
BENCH_GRPC_NAME := bench_emit_grpc

NOW_BIN        := $(OUT)/$(NOW_NAME)
MAIN_BIN       := $(OUT)/$(APP_NAME)
BENCH_HTTP_BIN := $(OUT)/$(BENCH_HTTP_NAME)
BENCH_GRPC_BIN := $(OUT)/$(BENCH_GRPC_NAME)

NOW_SRC   := cmd/now/main.go
MAIN_SRC  := cmd/celestialtree/main.go
MAIN_SRC  += $(wildcard internal/**/*.go)

BENCH_HTTP_PKG := ./bench/http/emit.go
BENCH_GRPC_PKG := ./bench/grpc/emit.go

# ---------- version ----------

VERSION ?= dev

ifeq ($(OS),Windows_NT)
  NULLDEV := nul
else
  NULLDEV := /dev/null
endif

GIT_COMMIT := $(shell git rev-parse --short HEAD 2>$(NULLDEV) || echo unknown)

# ---------- phony ----------

.PHONY: all build run bench bench-http bench-grpc version

all: build

# ---------- build ----------

build: $(MAIN_BIN) $(BENCH_HTTP_BIN) $(BENCH_GRPC_BIN) $(NOW_BIN)

$(OUT):
	mkdir $(OUT)

$(NOW_BIN): $(OUT) $(NOW_SRC)
	go build -o $@ $(NOW_SRC)

$(MAIN_BIN): $(OUT) $(NOW_BIN) $(MAIN_SRC)
	go build -ldflags "\
	-X github.com/Mr-xiaotian/CelestialTree/internal/version.Version=$(VERSION) \
	-X github.com/Mr-xiaotian/CelestialTree/internal/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/Mr-xiaotian/CelestialTree/internal/version.BuildTime=$(shell bin/now)" \
	-o $@ ./cmd/celestialtree

$(BENCH_HTTP_BIN): $(OUT) $(BENCH_HTTP_PKG)
	go build -o $@ $(BENCH_HTTP_PKG)

$(BENCH_GRPC_BIN): $(OUT) $(BENCH_GRPC_PKG)
	go build -o $@ $(BENCH_GRPC_PKG)

# ---------- run ----------

run: $(MAIN_BIN)
	$(MAIN_BIN)

# ---------- bench ----------

bench: bench-http bench-grpc

bench-http: $(BENCH_HTTP_BIN)
	$(BENCH_HTTP_BIN)

bench-grpc: $(BENCH_GRPC_BIN)
	$(BENCH_GRPC_BIN)

# ---------- info ----------

version:
	@echo $(APP_NAME) $(VERSION)($(GIT_COMMIT)) built at $(shell bin/now)
