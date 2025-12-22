OUT := bin
BIN := $(OUT)/main.exe

SRC := main.go
SRC += $(wildcard internal/tree/*.go)

.PHONY: all run

all: run

run: $(BIN)
	$(BIN)

# 编译规则
$(BIN): $(SRC)
	go build -o $@ .
