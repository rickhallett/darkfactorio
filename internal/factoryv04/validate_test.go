package factoryv04

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateBundlePassesExample(t *testing.T) {
	rep, err := ValidateBundle(filepath.Join("..", ".."), "factory/v0.4/examples/bundle.json")
	if err != nil {
		t.Fatalf("ValidateBundle error: %v", err)
	}
	if !rep.Passed {
		t.Fatalf("expected pass, got failures: %v", rep.Failures)
	}
}

func TestValidateBundleFailsCycle(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "bundle.json"), `{"spec_path":"spec.json","holdout_path":"holdout.json","twins_path":"twins.json","release_path":"release.json","policy_path":"policy.json","econ_path":"econ.json","orchestration_path":"orch.json"}`)
	write(t, filepath.Join(root, "spec.json"), `{"title":"x","objective":"y","non_negotiables":["a","b","c"],"acceptance":["a","b","c"]}`)
	write(t, filepath.Join(root, "holdout.json"), `{"scenario_total":7,"scenario_passed":7,"hidden_from_agent":true}`)
	write(t, filepath.Join(root, "twins.json"), `{"services":[{"name":"jira","mode":"simulated","contract_version":"v1","healthy":true,"failure_policy":"fail-closed"},{"name":"okta","mode":"simulated","contract_version":"v1","healthy":true,"failure_policy":"fail-closed"}]}`)
	write(t, filepath.Join(root, "artifact.txt"), "ok")
	write(t, filepath.Join(root, "release.json"), `{"candidate_id":"c1","artifact_path":"artifact.txt","baseline_pass":true,"adversarial_pass":true,"holdout_pass":true,"policy_pass":true,"econ_pass":true,"rollback_steps":["a","b","c"]}`)
	write(t, filepath.Join(root, "policy.json"), `{"required_controls":["C1"],"attestations":[{"control_id":"C1","owner":"x","timestamp":"2026-02-19T00:00:00Z","evidence":["artifact.txt"]}]}`)
	write(t, filepath.Join(root, "econ.json"), `{"token_budget_per_day":10,"token_observed":9,"cost_budget_per_day":10,"cost_observed":9,"p95_latency_ms_max":1000,"p95_latency_ms":900}`)
	write(t, filepath.Join(root, "orch.json"), `{"agents":[{"name":"a","role":"planner"},{"name":"b","role":"executor"}],"stages":[{"id":"validation","depends_on":["build"]},{"id":"build","depends_on":["validation"]}]}`)

	rep, err := ValidateBundle(root, "bundle.json")
	if err != nil {
		t.Fatalf("ValidateBundle error: %v", err)
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
