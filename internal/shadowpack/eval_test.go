package shadowpack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEvaluatePass(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "manifest.json"), `{
  "pack_id":"sp-1",
  "candidate_results":"candidate.json",
  "holdout_results":"holdout.json",
  "candidate_producer":"impl-agent",
  "holdout_producer":"qa-holdout",
  "criteria":{"min_overlap":3,"max_outcome_mismatch_rate_percent":10,"max_p95_latency_drift_percent":25}
}`)
	write(t, filepath.Join(root, "candidate.json"), `[
  {"scenario_id":"s1","outcome":"pass","latency_ms":100},
  {"scenario_id":"s2","outcome":"pass","latency_ms":120},
  {"scenario_id":"s3","outcome":"fail","latency_ms":130}
]`)
	write(t, filepath.Join(root, "holdout.json"), `[
  {"scenario_id":"s1","outcome":"pass","latency_ms":105},
  {"scenario_id":"s2","outcome":"pass","latency_ms":118},
  {"scenario_id":"s3","outcome":"fail","latency_ms":128}
]`)
	rep, err := Evaluate(root, "manifest.json")
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}
	if !rep.Passed {
		t.Fatalf("expected pass, got failures: %v", rep.Failures)
	}
}

func TestEvaluateFailsSameProducer(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "manifest.json"), `{
  "pack_id":"sp-2",
  "candidate_results":"candidate.json",
  "holdout_results":"holdout.json",
  "candidate_producer":"same",
  "holdout_producer":"same",
  "criteria":{"min_overlap":1,"max_outcome_mismatch_rate_percent":0,"max_p95_latency_drift_percent":10}
}`)
	write(t, filepath.Join(root, "candidate.json"), `[{"scenario_id":"s1","outcome":"pass","latency_ms":100}]`)
	write(t, filepath.Join(root, "holdout.json"), `[{"scenario_id":"s1","outcome":"pass","latency_ms":100}]`)
	rep, err := Evaluate(root, "manifest.json")
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}
	if rep.Passed {
		t.Fatalf("expected failure")
	}
}

func write(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
