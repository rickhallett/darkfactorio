package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rickhallett/darkfactorio/internal/dfcorpus"
	"github.com/rickhallett/darkfactorio/internal/level4gate"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("dfcorpusv01", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var inputs string
	var criteriaPath string
	var output string
	var windows string

	fs.StringVar(&inputs, "inputs", "", "comma-separated NDJSON files (required)")
	fs.StringVar(&criteriaPath, "criteria", "profiles/level4-gate-v0.1-adversarial.json", "criteria profile JSON path")
	fs.StringVar(&output, "output", "text", "output format: text|json")
	fs.StringVar(&windows, "windows", "", "optional comma-separated window_id filter")

	if err := fs.Parse(args); err != nil {
		return 1
	}
	if strings.TrimSpace(inputs) == "" {
		fmt.Fprintln(os.Stderr, "error: --inputs is required")
		return 1
	}

	criteria, err := loadCriteria(criteriaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: loading criteria: %v\n", err)
		return 1
	}

	res, err := dfcorpus.Replay(dfcorpus.ReplayOptions{
		Inputs:       splitCSV(inputs),
		WindowFilter: asSet(splitCSV(windows)),
		Criteria:     criteria,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	switch output {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(res.Report); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
	default:
		printText(res.Report, len(res.Records))
	}

	if !res.Report.Passed {
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
	var c level4gate.Criteria
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return level4gate.Criteria{}, err
	}
	if c.MinRuns <= 0 {
		return level4gate.Criteria{}, fmt.Errorf("min_runs must be > 0")
	}
	return c, nil
}

func splitCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func asSet(vals []string) map[string]struct{} {
	if len(vals) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(vals))
	for _, v := range vals {
		out[v] = struct{}{}
	}
	return out
}

func printText(report level4gate.GateReport, corpusSize int) {
	fmt.Printf("darkfactorio corpus gate\n")
	fmt.Printf("records: %d\n", corpusSize)
	fmt.Printf("passed: %v\n", report.Passed)
	fmt.Printf("run_count: %d\n", report.Metrics.RunCount)
	fmt.Printf("run_count_by_class: %v\n", report.Metrics.RunCountByClass)
	fmt.Printf("scenario_pass_rate: %.2f%%\n", report.Metrics.ScenarioPassRatePercent)
	fmt.Printf("first_pass_rate: %.2f%%\n", report.Metrics.FirstPassRatePercent)
	fmt.Printf("mean_retries: %.2f\n", report.Metrics.MeanRetries)
	fmt.Printf("decision_reversal_rate: %.2f%%\n", report.Metrics.DecisionReversalPercent)
	fmt.Printf("approved_run_critical_incidents: %d\n", report.Metrics.ApprovedRunCriticalIncidents)
	if len(report.Failures) > 0 {
		fmt.Println("failures:")
		for _, f := range report.Failures {
			fmt.Printf("- %s\n", f)
		}
	}
}
