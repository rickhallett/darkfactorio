package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rickhallett/darkfactorio/internal/dfwindow"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("dfwindowv01", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var windowID string
	var runsPath string
	var appendCount int
	var logLearning bool
	var start string
	var interval string
	var baseline string
	var adversarial string
	var quality string

	fs.StringVar(&windowID, "window", "", "window id (required)")
	fs.StringVar(&runsPath, "runs", "", "runs NDJSON path (default runs/<window>.ndjson)")
	fs.IntVar(&appendCount, "append", 2, "number of runs to append")
	fs.BoolVar(&logLearning, "log-learning", true, "append learning journal entry")
	fs.StringVar(&start, "start", "", "RFC3339 start time for first appended run (default now UTC)")
	fs.StringVar(&interval, "interval", "15m", "duration between appended runs")
	fs.StringVar(&baseline, "baseline", "profiles/level4-gate-v0.1-baseline.json", "baseline criteria path")
	fs.StringVar(&adversarial, "adversarial", "profiles/level4-gate-v0.1-adversarial.json", "adversarial criteria path")
	fs.StringVar(&quality, "quality", "standard", "run quality mode: standard|high")

	if err := fs.Parse(args); err != nil {
		return 1
	}
	if windowID == "" {
		fmt.Fprintln(os.Stderr, "error: --window is required")
		return 1
	}

	var t time.Time
	if start != "" {
		parsed, err := time.Parse(time.RFC3339, start)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: invalid --start: %v\n", err)
			return 1
		}
		t = parsed
	}
	d, err := time.ParseDuration(interval)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid --interval: %v\n", err)
		return 1
	}

	res, err := dfwindow.Advance(dfwindow.AdvanceOptions{
		Root:                ".",
		WindowID:            windowID,
		RunsPath:            runsPath,
		AppendCount:         appendCount,
		StartTime:           t,
		Interval:            d,
		BaselineCriteria:    baseline,
		AdversarialCriteria: adversarial,
		LogLearning:         logLearning,
		QualityMode:         quality,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	first := res.Added[0].RunID
	last := res.Added[len(res.Added)-1].RunID
	fmt.Printf("window advance complete: %s\n", windowID)
	fmt.Printf("runs appended: %d (%s..%s)\n", len(res.Added), first, last)
	fmt.Printf("runs path: %s\n", res.RunsPath)
	fmt.Printf("baseline: passed=%v run_count=%d scenario_pass=%.2f%%\n", res.BaselineReport.Passed, res.BaselineReport.Metrics.RunCount, res.BaselineReport.Metrics.ScenarioPassRatePercent)
	fmt.Printf("adversarial: passed=%v run_count=%d scenario_pass=%.2f%%\n", res.AdversarialReport.Passed, res.AdversarialReport.Metrics.RunCount, res.AdversarialReport.Metrics.ScenarioPassRatePercent)
	if res.LearningEntryPath != "" {
		fmt.Printf("learning entry: %s\n", res.LearningEntryPath)
	}

	if !res.BaselineReport.Passed {
		return 2
	}
	return 0
}
