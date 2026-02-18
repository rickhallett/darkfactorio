.PHONY: test gate-sample build-dfgate

test:
	go test ./...

build-dfgate:
	go build -o ./bin/dfgate ./cmd/dfgate

gate-sample:
	go run ./cmd/dfgate -input runs/examples/window-sample.ndjson -window w-2026-02-l4-01 -output text

