package dfwindow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rickhallett/darkfactorio/internal/level4gate"
)

func TestAdvanceAppendsRecordsAndEvaluates(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "profiles/level4-gate-v0.1-baseline.json"), `{"version":"b","min_runs":2,"thresholds":{"min_scenario_pass_rate_percent":90,"min_first_pass_rate_percent":70,"max_mean_retries":2,"max_decision_reversal_percent":5,"max_approved_incidents":0},"required_class_minimum":{"low_risk_feature":1,"medium_integration":1}}`)
	mustWrite(t, filepath.Join(root, "profiles/level4-gate-v0.1-adversarial.json"), `{"version":"a","min_runs":4,"thresholds":{"min_scenario_pass_rate_percent":95,"min_first_pass_rate_percent":85,"max_mean_retries":1,"max_decision_reversal_percent":2,"max_approved_incidents":0},"required_class_minimum":{"low_risk_feature":2,"medium_integration":2}}`)

	res, err := Advance(AdvanceOptions{
		Root:        root,
		WindowID:    "w-test",
		AppendCount: 2,
		RunsPath:    "runs/w-test.ndjson",
		LogLearning: true,
		QualityMode: "high",
	})
	if err != nil {
		t.Fatalf("Advance failed: %v", err)
	}
	if len(res.Added) != 2 {
		t.Fatalf("expected 2 added records, got %d", len(res.Added))
	}
	if !res.BaselineReport.Passed {
		t.Fatalf("expected baseline pass")
	}
	if res.AdversarialReport.Passed {
		t.Fatalf("expected adversarial fail with only 2 runs")
	}
	if res.LearningEntryPath == "" {
		t.Fatalf("expected learning entry path")
	}

	recs, err := level4gate.LoadNDJSON(filepath.Join(root, "runs/w-test.ndjson"), "w-test")
	if err != nil {
		t.Fatalf("LoadNDJSON failed: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 records in file, got %d", len(recs))
	}
}

func TestAdvanceRejectsInvalidQualityMode(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "profiles/level4-gate-v0.1-baseline.json"), `{"version":"b","min_runs":1,"thresholds":{"min_scenario_pass_rate_percent":90,"min_first_pass_rate_percent":70,"max_mean_retries":2,"max_decision_reversal_percent":5,"max_approved_incidents":0},"required_class_minimum":{"low_risk_feature":0,"medium_integration":0}}`)
	mustWrite(t, filepath.Join(root, "profiles/level4-gate-v0.1-adversarial.json"), `{"version":"a","min_runs":1,"thresholds":{"min_scenario_pass_rate_percent":90,"min_first_pass_rate_percent":70,"max_mean_retries":2,"max_decision_reversal_percent":5,"max_approved_incidents":0},"required_class_minimum":{"low_risk_feature":0,"medium_integration":0}}`)

	_, err := Advance(AdvanceOptions{
		Root:        root,
		WindowID:    "w-test",
		AppendCount: 1,
		QualityMode: "bad",
	})
	if err == nil {
		t.Fatalf("expected error for invalid quality mode")
	}
}

func mustWrite(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
