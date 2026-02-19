package adapterholdout

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncFromFileSource(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "shadowpacks/demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "shadowpacks/demo/manifest.json"), []byte(`{
  "pack_id":"demo-pack",
  "candidate_results":"shadowpacks/demo/candidate.json",
  "holdout_results":"shadowpacks/demo/holdout.json",
  "candidate_producer":"impl",
  "holdout_producer":"qa",
  "criteria":{"min_overlap":1,"max_outcome_mismatch_rate_percent":100,"max_p95_latency_drift_percent":100}
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "shadowpacks/demo/candidate.json"), []byte(`[{"scenario_id":"s-1","outcome":"pass","latency_ms":10}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(root, "source-holdout.json")
	if err := os.WriteFile(src, []byte(`[{"scenario_id":"s-1","outcome":"pass","latency_ms":11}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(root, "cfg.json")
	if err := os.WriteFile(cfg, []byte(`{
  "project":"demo",
  "holdout_producer":"qa",
  "holdout_repo":"example/qa",
  "holdout_sha":"abc",
  "holdout_results_source":"`+src+`"
}`), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Sync(SyncOptions{
		Root:       root,
		ConfigPath: "cfg.json",
		Validate:   true,
	})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if res.ShadowValidation != "pass" {
		t.Fatalf("expected shadow pass")
	}
	if _, err := os.Stat(filepath.Join(root, "shadowpacks/demo/holdout-provenance.json")); err != nil {
		t.Fatalf("missing provenance: %v", err)
	}
}
