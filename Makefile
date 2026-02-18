.PHONY: test gate-sample gate-sample-adversarial build-dfgate build-dfgatev01

test:
	go test ./...

build-dfgate:
	go build -o ./bin/dfgate ./cmd/dfgate

build-dfgatev01:
	go build -o ./bin/dfgatev01 ./cmd/dfgatev01

gate-sample:
	go run ./cmd/dfgatev01 -input runs/examples/window-sample.ndjson -window w-2026-02-l4-01 -criteria profiles/level4-gate-v0.1-baseline.json -output text

gate-sample-adversarial:
	go run ./cmd/dfgatev01 -input runs/examples/window-sample.ndjson -window w-2026-02-l4-01 -criteria profiles/level4-gate-v0.1-adversarial.json -output text
