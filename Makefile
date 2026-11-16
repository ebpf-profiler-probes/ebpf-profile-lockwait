.PHONY: generate build test fmt example

# Prefer Homebrew LLVM when present (Apple clang cannot target BPF).
LLVM_PREFIX := $(shell brew --prefix llvm 2>/dev/null)
ifneq ($(LLVM_PREFIX),)
export PATH := $(LLVM_PREFIX)/bin:$(PATH)
endif

# Include go install bin (builder/ocb lives here).
export PATH := $(shell go env GOPATH)/bin:$(PATH)

# ocb installs as "builder": go install go.opentelemetry.io/collector/cmd/builder@v0.165.0
OCB ?= $(shell command -v ocb 2>/dev/null || command -v builder 2>/dev/null)
ifeq ($(OCB),)
OCB := go run go.opentelemetry.io/collector/cmd/builder@v0.165.0
endif

# Profiler is Linux-only; default to cross-compile from macOS.
GOOS ?= linux
GOARCH ?= $(shell go env GOARCH)

generate:
	go generate ./lockwait/...

test:
	go test ./...

# Build a custom collector that includes lockwait → ./dist/otelcol-lockwait
build: generate
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(OCB) --config manifest.yaml

fmt:
	gofmt -w ./lockwait ./example

example:
	go run ./example
