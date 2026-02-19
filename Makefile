.PHONY: test gate-sample gate-sample-adversarial build-dfgate build-dfgatev01 build-dflearn build-dfwindowv01 build-dfcorpusv01 build-dffactoryv04 build-dffactoryv05 build-dfstressv04 build-dfshadowv01 build-dfonboardv01 build-dfadapterholdoutv01 learning-touch learning-check window-advance window-advance-high corpus-adversarial factory-v04-validate factory-v05-validate stress-v04 shadow-pack onboard-project onboard-validate holdout-sync

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

build-dffactoryv04:
	$(GO) build -o ./bin/dffactoryv04 ./cmd/dffactoryv04

build-dffactoryv05:
	$(GO) build -o ./bin/dffactoryv05 ./cmd/dffactoryv05

build-dfstressv04:
	$(GO) build -o ./bin/dfstressv04 ./cmd/dfstressv04

build-dfshadowv01:
	$(GO) build -o ./bin/dfshadowv01 ./cmd/dfshadowv01

build-dfonboardv01:
	$(GO) build -o ./bin/dfonboardv01 ./cmd/dfonboardv01

build-dfadapterholdoutv01:
	$(GO) build -o ./bin/dfadapterholdoutv01 ./cmd/dfadapterholdoutv01

learning-touch:
	$(GO) run ./cmd/dflearn touch --source-project darkfactorio --summary "manual learning checkpoint"

learning-check:
	$(GO) run ./cmd/dflearn check --base HEAD~1 --head HEAD

window-advance:
	$(GO) run ./cmd/dfwindowv01 --window $(WINDOW) --append $(or $(APPEND),2)

window-advance-high:
	$(GO) run ./cmd/dfwindowv01 --window $(WINDOW) --append $(or $(APPEND),2) --quality high --quality-reason "$(QUALITY_REASON)"

corpus-adversarial:
	$(GO) run ./cmd/dfcorpusv01 --inputs runs/w-2026-02-l4-02.ndjson,runs/w-2026-02-l4-03.ndjson --criteria profiles/level4-gate-v0.1-adversarial.json --output text

factory-v04-validate:
	$(GO) run ./cmd/dffactoryv04 --bundle factory/v0.4/examples/bundle.json --output text

factory-v05-validate:
	$(GO) run ./cmd/dffactoryv05 --bundle factory/v0.5/examples/bundle.json --output text

stress-v04:
	$(GO) run ./cmd/dfstressv04 --output text

shadow-pack:
	$(GO) run ./cmd/dfshadowv01 --manifest shadowpacks/examples/manifest.json --output text

onboard-project:
	$(GO) run ./cmd/dfonboardv01 scaffold --project $(PROJECT) --candidate-producer $(or $(CANDIDATE_PRODUCER),$(PROJECT)-impl) --holdout-producer $(or $(HOLDOUT_PRODUCER),$(PROJECT)-qa)

onboard-validate:
	$(GO) run ./cmd/dfonboardv01 validate-artifacts --manifest shadowpacks/$(PROJECT)/manifest.json

holdout-sync:
	$(GO) run ./cmd/dfadapterholdoutv01 --config $(CONFIG) --validate true --output text

gate-sample:
	$(GO) run ./cmd/dfgatev01 -input runs/examples/window-sample.ndjson -window w-2026-02-l4-01 -criteria profiles/level4-gate-v0.1-baseline.json -output text

gate-sample-adversarial:
	$(GO) run ./cmd/dfgatev01 -input runs/examples/window-sample.ndjson -window w-2026-02-l4-01 -criteria profiles/level4-gate-v0.1-adversarial.json -output text
