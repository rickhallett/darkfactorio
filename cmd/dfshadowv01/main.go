package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/rickhallett/darkfactorio/internal/shadowpack"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("dfshadowv01", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	manifest := fs.String("manifest", "shadowpacks/examples/manifest.json", "shadow pack manifest path")
	output := fs.String("output", "text", "output format: text|json")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	rep, err := shadowpack.Evaluate(".", *manifest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	switch *output {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(rep)
	default:
		fmt.Printf("shadow pack: %s\n", rep.PackID)
		fmt.Printf("passed: %v\n", rep.Passed)
		fmt.Printf("overlap: %d\n", rep.OverlapCount)
		fmt.Printf("candidate_only: %d\n", rep.CandidateOnlyCount)
		fmt.Printf("holdout_only: %d\n", rep.HoldoutOnlyCount)
		fmt.Printf("mismatch_rate: %.2f%%\n", rep.OutcomeMismatchRatePercent)
		fmt.Printf("p95_latency_drift: %.2f%%\n", rep.P95LatencyDriftPercent)
		if len(rep.Failures) > 0 {
			fmt.Println("failures:")
			for _, f := range rep.Failures {
				fmt.Printf("- %s\n", f)
			}
		}
	}
	if !rep.Passed {
		return 2
	}
	return 0
}
