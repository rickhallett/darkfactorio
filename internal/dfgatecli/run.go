package dfgatecli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/rickhallett/darkfactorio/internal/level4gate"
)

func Run(args []string) int {
	fs := flag.NewFlagSet("dfgate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var input string
	var windowID string
	var output string
	var criteriaPath string

	fs.StringVar(&input, "input", "", "path to NDJSON metrics file (required)")
	fs.StringVar(&windowID, "window", "", "optional window_id filter")
	fs.StringVar(&output, "output", "text", "output format: text|json")
	fs.StringVar(&criteriaPath, "criteria", "", "optional criteria profile JSON path")

	if err := fs.Parse(args); err != nil {
		return 1
	}
	if input == "" {
		fmt.Fprintln(os.Stderr, "error: -input is required")
		return 1
	}

	records, err := level4gate.LoadNDJSON(input, windowID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	criteria := level4gate.DefaultCriteria()
	if criteriaPath != "" {
		loaded, err := loadCriteria(criteriaPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: loading criteria: %v\n", err)
			return 1
		}
		criteria = loaded
	}
	report := level4gate.EvaluateWithCriteria(records, criteria, windowID)

	switch output {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
	default:
		printText(report)
	}

	if !report.Passed {
		return 2
	}
	return 0
}

func loadCriteria(path string) (level4gate.Criteria, error) {
	f, err := os.Open(path)
	if err != nil {
		return level4gate.Criteria{}, err
	}
	defer f.Close()
	return decodeCriteria(f)
}

func decodeCriteria(r io.Reader) (level4gate.Criteria, error) {
	var c level4gate.Criteria
	dec := json.NewDecoder(r)
	if err := dec.Decode(&c); err != nil {
		return level4gate.Criteria{}, err
	}
	if c.MinRuns <= 0 {
		return level4gate.Criteria{}, fmt.Errorf("min_runs must be > 0")
	}
	return c, nil
}

func printText(report level4gate.GateReport) {
	fmt.Printf("darkfactorio level4 gate window=%q\n", report.WindowID)
	fmt.Printf("passed: %v\n", report.Passed)
	fmt.Printf("run_count: %d\n", report.Metrics.RunCount)
	fmt.Printf("run_count_by_class: %v\n", report.Metrics.RunCountByClass)
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

