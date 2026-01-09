# ---------- basic ----------

OUT := bin

APP_NAME   := celestialtree
BENCH_NAME := bench_emit
NOW_NAME   := now

MAIN_BIN  := $(OUT)/$(APP_NAME)
BENCH_BIN := $(OUT)/$(BENCH_NAME)
NOW_BIN   := $(OUT)/$(NOW_NAME)

MAIN_SRC  := main.go
MAIN_SRC  += $(wildcard internal/**/*.go)
BENCH_SRC := bench/bench_emit.go
NOW_SRC   := internal/tools/now.go

# ---------- version ----------

VERSION ?= dev

ifeq ($(OS),Windows_NT)
  NULLDEV := nul
else
  NULLDEV := /dev/null
endif

GIT_COMMIT := $(shell git rev-parse --short HEAD 2>$(NULLDEV) || echo unknown)

# ---------- phony ----------

.PHONY: all build run bench version

all: build

# ---------- build ----------

build: $(MAIN_BIN) $(BENCH_BIN) $(NOW_BIN)

$(OUT):
	mkdir $(OUT)

$(MAIN_BIN): $(OUT) $(NOW_BIN) $(MAIN_SRC)
	go build -ldflags "\
	-X celestialtree/internal/version.Version=$(VERSION) \
	-X celestialtree/internal/version.GitCommit=$(GIT_COMMIT) \
	-X celestialtree/internal/version.BuildTime=$(shell bin/now)" \
	-o $@ .

$(BENCH_BIN): $(OUT) $(BENCH_SRC)
	go build -o $@ $(BENCH_SRC)

$(NOW_BIN): $(OUT) $(NOW_SRC)
	go build -o $@ $(NOW_SRC)

# ---------- run ----------

run: $(MAIN_BIN)
	$(MAIN_BIN)

bench: $(BENCH_BIN)
	$(BENCH_BIN)

# ---------- info ----------

version:
	@echo $(APP_NAME) $(VERSION) ($(GIT_COMMIT)) built at $(BUILD_TIME)
