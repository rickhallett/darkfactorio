package dfcorpus

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rickhallett/darkfactorio/internal/level4gate"
)

func TestReplayAcrossFiles(t *testing.T) {
	root := t.TempDir()
	f1 := filepath.Join(root, "w1.ndjson")
	f2 := filepath.Join(root, "w2.ndjson")
	mustWrite(t, f1, `{"window_id":"w1","run_id":"run-001","pipeline_id":"p-low-001","pipeline_class":"low_risk_feature","scenario_total":10,"scenario_passed":10,"first_pass_success":true,"retries":1,"interventions":1,"decision":"approved","decision_reversed":false,"critical_incident":false,"timestamp":"2026-02-19T00:00:00Z"}`+"\n")
	mustWrite(t, f2, `{"window_id":"w2","run_id":"run-001","pipeline_id":"p-med-001","pipeline_class":"medium_integration","scenario_total":10,"scenario_passed":9,"first_pass_success":true,"retries":1,"interventions":1,"decision":"approved","decision_reversed":false,"critical_incident":false,"timestamp":"2026-02-19T00:05:00Z"}`+"\n")

	c := level4gate.DefaultCriteria()
	c.MinRuns = 2
	c.RequiredClassMinimum = map[string]int{
		"low_risk_feature":   1,
		"medium_integration": 1,
	}

	res, err := Replay(ReplayOptions{
		Inputs:   []string{f1, f2},
		Criteria: c,
	})
	if err != nil {
		t.Fatalf("Replay failed: %v", err)
	}
	if len(res.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(res.Records))
	}
	if !res.Report.Passed {
		t.Fatalf("expected pass, got failures: %v", res.Report.Failures)
	}
}

func TestReplayWithWindowFilter(t *testing.T) {
	root := t.TempDir()
	f := filepath.Join(root, "mix.ndjson")
	mustWrite(t, f, `{"window_id":"w1","run_id":"run-001","pipeline_id":"p-low-001","pipeline_class":"low_risk_feature","scenario_total":10,"scenario_passed":10,"first_pass_success":true,"retries":1,"interventions":1,"decision":"approved","decision_reversed":false,"critical_incident":false,"timestamp":"2026-02-19T00:00:00Z"}`+"\n"+`{"window_id":"w2","run_id":"run-001","pipeline_id":"p-med-001","pipeline_class":"medium_integration","scenario_total":10,"scenario_passed":10,"first_pass_success":true,"retries":1,"interventions":1,"decision":"approved","decision_reversed":false,"critical_incident":false,"timestamp":"2026-02-19T00:05:00Z"}`+"\n")

	c := level4gate.DefaultCriteria()
	c.MinRuns = 1
	c.RequiredClassMinimum = map[string]int{
		"low_risk_feature": 1,
	}

	res, err := Replay(ReplayOptions{
		Inputs: []string{f},
		WindowFilter: map[string]struct{}{
			"w1": {},
		},
		Criteria: c,
	})
	if err != nil {
		t.Fatalf("Replay failed: %v", err)
	}
	if len(res.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(res.Records))
	}
}

func mustWrite(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
