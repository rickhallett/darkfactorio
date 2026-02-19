package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/rickhallett/darkfactorio/internal/adaptertelemetry"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("dfadaptertelemetryv01", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	config := fs.String("config", "", "adapter telemetry config path (required)")
	output := fs.String("output", "text", "output format: text|json")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	res, err := adaptertelemetry.Sync(".", *config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	switch *output {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
	default:
		fmt.Printf("runtime output: %s\n", res.RuntimePath)
		fmt.Printf("econ output: %s\n", res.EconPath)
	}
	return 0
}
