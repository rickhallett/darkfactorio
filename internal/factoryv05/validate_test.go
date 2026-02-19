package factoryv05

import (
	"path/filepath"
	"testing"
)

func TestValidateBundlePassesExample(t *testing.T) {
	rep, err := ValidateBundle(filepath.Join("..", ".."), "factory/v0.5/examples/bundle.json")
	if err != nil {
		t.Fatalf("ValidateBundle error: %v", err)
	}
	if !rep.Passed {
		t.Fatalf("expected pass, got failures: %v", rep.Failures)
	}
}
