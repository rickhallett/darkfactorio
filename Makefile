.PHONY: test gate-sample gate-sample-adversarial build-dfgate build-dfgatev01 build-dflearn learning-touch learning-check

GOCACHE ?= $(CURDIR)/.cache/go-build
GO := GOCACHE=$(GOCACHE) go

test:
	$(GO) test ./...

build-dfgate:
	$(GO) build -o ./bin/dfgate ./cmd/dfgate

build-dfgatev01:
	$(GO) build -o ./bin/dfgatev01 ./cmd/dfgatev01

build-dflearn:
	$(GO) build -o ./bin/dflearn ./cmd/dflearn

learning-touch:
	$(GO) run ./cmd/dflearn touch --source-project darkfactorio --summary "manual learning checkpoint"

learning-check:
	$(GO) run ./cmd/dflearn check --base HEAD~1 --head HEAD

gate-sample:
	$(GO) run ./cmd/dfgatev01 -input runs/examples/window-sample.ndjson -window w-2026-02-l4-01 -criteria profiles/level4-gate-v0.1-baseline.json -output text

gate-sample-adversarial:
	$(GO) run ./cmd/dfgatev01 -input runs/examples/window-sample.ndjson -window w-2026-02-l4-01 -criteria profiles/level4-gate-v0.1-adversarial.json -output text
