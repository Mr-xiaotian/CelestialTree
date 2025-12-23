# ---------- basic ----------

OUT := bin

APP_NAME := celestialtree
BENCH_NAME := bench_emit

MAIN_BIN  := $(OUT)/$(APP_NAME)
BENCH_BIN := $(OUT)/$(BENCH_NAME)

MAIN_SRC  := main.go
MAIN_SRC  += $(wildcard internal/**/*.go)

BENCH_SRC := bench/bench_emit.go

# ---------- version ----------

VERSION ?= dev

GIT_COMMIT := $(shell git rev-parse --short HEAD 2>nul || echo unknown)
BUILD_TIME := $(shell go run internal/tools/now.go 2>nul || echo unknown)

LDFLAGS := -X celestialtree/internal/version.Version=$(VERSION) \
           -X celestialtree/internal/version.GitCommit=$(GIT_COMMIT) \
           -X celestialtree/internal/version.BuildTime=$(BUILD_TIME)

# ---------- phony ----------

.PHONY: all build run bench version

all: build

# ---------- build ----------

build: $(MAIN_BIN) $(BENCH_BIN)

$(OUT):
	mkdir $(OUT)

$(MAIN_BIN): $(OUT) $(MAIN_SRC)
	go build -ldflags "$(LDFLAGS)" -o $@ .

$(BENCH_BIN): $(OUT) $(BENCH_SRC)
	go build -o $@ $(BENCH_SRC)

# ---------- run ----------

run: $(MAIN_BIN)
	$(MAIN_BIN)

bench: $(BENCH_BIN)
	$(BENCH_BIN)

# ---------- info ----------

version:
	@echo $(APP_NAME) $(VERSION) ($(GIT_COMMIT)) built at $(BUILD_TIME)
