.PHONY: test gate-sample gate-sample-adversarial build-dfgate build-dfgatev01 build-dflearn build-dfwindowv01 build-dfcorpusv01 learning-touch learning-check window-advance window-advance-high corpus-adversarial

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

build-dfwindowv01:
	$(GO) build -o ./bin/dfwindowv01 ./cmd/dfwindowv01

build-dfcorpusv01:
	$(GO) build -o ./bin/dfcorpusv01 ./cmd/dfcorpusv01

learning-touch:
	$(GO) run ./cmd/dflearn touch --source-project darkfactorio --summary "manual learning checkpoint"

learning-check:
	$(GO) run ./cmd/dflearn check --base HEAD~1 --head HEAD

window-advance:
	$(GO) run ./cmd/dfwindowv01 --window $(WINDOW) --append $(or $(APPEND),2)

window-advance-high:
	$(GO) run ./cmd/dfwindowv01 --window $(WINDOW) --append $(or $(APPEND),2) --quality high

corpus-adversarial:
	$(GO) run ./cmd/dfcorpusv01 --inputs runs/w-2026-02-l4-02.ndjson,runs/w-2026-02-l4-03.ndjson --criteria profiles/level4-gate-v0.1-adversarial.json --output text

gate-sample:
	$(GO) run ./cmd/dfgatev01 -input runs/examples/window-sample.ndjson -window w-2026-02-l4-01 -criteria profiles/level4-gate-v0.1-baseline.json -output text

gate-sample-adversarial:
	$(GO) run ./cmd/dfgatev01 -input runs/examples/window-sample.ndjson -window w-2026-02-l4-01 -criteria profiles/level4-gate-v0.1-adversarial.json -output text
