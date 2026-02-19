package onboard

import (
	"path/filepath"
	"testing"
)

func TestScaffoldAndValidate(t *testing.T) {
	root := t.TempDir()
	manifestPath, err := Scaffold(ScaffoldOptions{
		Root:              root,
		Project:           "tspit",
		CandidateProducer: "tspit-impl",
		HoldoutProducer:   "tspit-qa",
	})
	if err != nil {
		t.Fatalf("Scaffold failed: %v", err)
	}
	if filepath.Base(manifestPath) != "manifest.json" {
		t.Fatalf("unexpected manifest path: %s", manifestPath)
	}
	rel := filepath.ToSlash(filepath.Join("shadowpacks", "tspit", "manifest.json"))
	if err := ValidateArtifacts(ValidateOptions{Root: root, Manifest: rel}); err != nil {
		t.Fatalf("ValidateArtifacts failed: %v", err)
	}
}
