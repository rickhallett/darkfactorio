package adaptertelemetry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSync(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "runtime-input.json"), `{"availability_percent":99.95,"error_rate_percent":0.2,"p95_latency_ms":320}`)
	mustWrite(t, filepath.Join(root, "billing-input.json"), `{"provider_cost_usd":980,"internal_cost_usd":955,"provider_tokens":740000,"internal_tokens":725000}`)
	mustWrite(t, filepath.Join(root, "cfg.json"), `{
  "runtime_input_path":"runtime-input.json",
  "billing_input_path":"billing-input.json",
  "out_runtime_path":"out/runtime-slo.json",
  "out_econ_path":"out/econ-reconcile.json",
  "min_availability_percent":99.9,
  "max_error_rate_percent":1.0,
  "max_p95_latency_ms":500,
  "max_cost_delta_percent":5,
  "max_token_delta_percent":5
}`)

	res, err := Sync(root, "cfg.json")
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if res.RuntimePath == "" || res.EconPath == "" {
		t.Fatalf("missing output paths")
	}
	if _, err := os.Stat(filepath.Join(root, res.RuntimePath)); err != nil {
		t.Fatalf("runtime output missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, res.EconPath)); err != nil {
		t.Fatalf("econ output missing: %v", err)
	}
}

func mustWrite(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
