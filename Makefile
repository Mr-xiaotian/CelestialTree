OUT := bin

MAIN_BIN  := $(OUT)/celestialtree.exe
BENCH_BIN := $(OUT)/bench_emit.exe

MAIN_SRC  := main.go
MAIN_SRC  += $(wildcard internal/tree/*.go)

BENCH_SRC := tools/bench/bench_emit.go

.PHONY: all build run bench clean

all: build

# ---------- build ----------

build: $(MAIN_BIN) $(BENCH_BIN)

$(OUT):
	mkdir -p $(OUT)

$(MAIN_BIN): $(OUT) $(MAIN_SRC)
	go build -o $@ .

$(BENCH_BIN): $(OUT) $(BENCH_SRC)
	go build -o $@ $(BENCH_SRC)

# ---------- run ----------

run: $(MAIN_BIN)
	$(MAIN_BIN)

bench_emit: $(BENCH_BIN)
	$(BENCH_BIN)

# ---------- clean ----------

clean:
	rm -rf $(OUT)
