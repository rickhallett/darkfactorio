package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/rickhallett/darkfactorio/internal/level4gate"
)

func main() {
	var input string
	var windowID string
	var output string

	flag.StringVar(&input, "input", "", "path to NDJSON metrics file (required)")
	flag.StringVar(&windowID, "window", "", "optional window_id filter")
	flag.StringVar(&output, "output", "text", "output format: text|json")
	flag.Parse()

	if input == "" {
		fmt.Fprintln(os.Stderr, "error: -input is required")
		os.Exit(1)
	}

	records, err := level4gate.LoadNDJSON(input, windowID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	report := level4gate.Evaluate(records, level4gate.DefaultThresholds(), windowID)

	switch output {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		printText(report)
	}

	if !report.Passed {
		os.Exit(2)
	}
}

func printText(report level4gate.GateReport) {
	fmt.Printf("darkfactorio level4 gate window=%q\n", report.WindowID)
	fmt.Printf("passed: %v\n", report.Passed)
	fmt.Printf("run_count: %d\n", report.Metrics.RunCount)
	fmt.Printf("scenario_pass_rate: %.2f%%\n", report.Metrics.ScenarioPassRatePercent)
	fmt.Printf("first_pass_rate: %.2f%%\n", report.Metrics.FirstPassRatePercent)
	fmt.Printf("mean_retries: %.2f\n", report.Metrics.MeanRetries)
	fmt.Printf(
		"intervention_trend: first_half=%.2f second_half=%.2f stable_or_decreasing=%v\n",
		report.Metrics.InterventionAvgFirstHalf,
		report.Metrics.InterventionAvgSecondHalf,
		report.Metrics.InterventionStableOrDecreasing,
	)
	fmt.Printf("decision_reversal_rate: %.2f%%\n", report.Metrics.DecisionReversalPercent)
	fmt.Printf("approved_run_critical_incidents: %d\n", report.Metrics.ApprovedRunCriticalIncidents)

	if len(report.Failures) > 0 {
		fmt.Println("failures:")
		for _, f := range report.Failures {
			fmt.Printf("- %s\n", f)
		}
	}
}

