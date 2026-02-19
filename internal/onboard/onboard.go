package onboard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rickhallett/darkfactorio/internal/shadowpack"
)

type ScaffoldOptions struct {
	Root              string
	Project           string
	CandidateProducer string
	HoldoutProducer   string
}

type ValidateOptions struct {
	Root     string
	Manifest string
}

func Scaffold(opts ScaffoldOptions) (string, error) {
	if opts.Root == "" {
		opts.Root = "."
	}
	if strings.TrimSpace(opts.Project) == "" {
		return "", fmt.Errorf("project is required")
	}
	project := sanitize(opts.Project)
	if project == "" {
		return "", fmt.Errorf("project name invalid after sanitization")
	}
	candidateProducer := strings.TrimSpace(opts.CandidateProducer)
	if candidateProducer == "" {
		candidateProducer = project + "-impl"
	}
	holdoutProducer := strings.TrimSpace(opts.HoldoutProducer)
	if holdoutProducer == "" {
		holdoutProducer = project + "-qa"
	}

	dir := filepath.Join(opts.Root, "shadowpacks", project)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	candidatePath := filepath.Join(dir, "candidate.json")
	holdoutPath := filepath.Join(dir, "holdout.json")
	checklistPath := filepath.Join(dir, "promotion-checklist.md")

	manifest := map[string]any{
		"pack_id":            project + "-shadowpack-v0.1",
		"candidate_results":  filepath.ToSlash(filepath.Join("shadowpacks", project, "candidate.json")),
		"holdout_results":    filepath.ToSlash(filepath.Join("shadowpacks", project, "holdout.json")),
		"candidate_producer": candidateProducer,
		"holdout_producer":   holdoutProducer,
		"criteria": map[string]any{
			"min_overlap":                       10,
			"max_outcome_mismatch_rate_percent": 10,
			"max_p95_latency_drift_percent":     25,
		},
	}
	if err := writeJSON(manifestPath, manifest); err != nil {
		return "", err
	}

	example := []map[string]any{
		{"scenario_id": "s-001", "outcome": "pass", "latency_ms": 120.0},
		{"scenario_id": "s-002", "outcome": "pass", "latency_ms": 135.0},
	}
	if err := writeJSON(candidatePath, example); err != nil {
		return "", err
	}
	if err := writeJSON(holdoutPath, example); err != nil {
		return "", err
	}

	checklist := buildChecklist(project)
	if err := os.WriteFile(checklistPath, []byte(checklist), 0o644); err != nil {
		return "", err
	}
	return manifestPath, nil
}

func ValidateArtifacts(opts ValidateOptions) error {
	if opts.Root == "" {
		opts.Root = "."
	}
	if opts.Manifest == "" {
		return fmt.Errorf("manifest is required")
	}
	manifest, err := loadJSON[shadowpack.Manifest](filepath.Join(opts.Root, opts.Manifest))
	if err != nil {
		return err
	}
	if manifest.CandidateProducer == manifest.HoldoutProducer {
		return fmt.Errorf("candidate_producer must differ from holdout_producer")
	}

	candidate, err := loadJSON[[]shadowpack.ScenarioResult](filepath.Join(opts.Root, manifest.CandidateResults))
	if err != nil {
		return fmt.Errorf("candidate_results: %w", err)
	}
	holdout, err := loadJSON[[]shadowpack.ScenarioResult](filepath.Join(opts.Root, manifest.HoldoutResults))
	if err != nil {
		return fmt.Errorf("holdout_results: %w", err)
	}
	if len(candidate) == 0 || len(holdout) == 0 {
		return fmt.Errorf("candidate/holdout artifacts cannot be empty")
	}
	return nil
}

func sanitize(in string) string {
	in = strings.ToLower(strings.TrimSpace(in))
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ".", "-")
	in = replacer.Replace(in)
	var b strings.Builder
	for _, r := range in {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return strings.Trim(b.String(), "-")
}

func buildChecklist(project string) string {
	return fmt.Sprintf(`# Promotion Checklist: %s

1. Candidate and holdout producers are independent.
2. Holdout scenarios are not visible to implementation agents.
3. `+"`make shadow-pack`"+` passes.
4. `+"`make corpus-adversarial`"+` passes.
5. `+"`make factory-v04-validate`"+` passes.
6. Learning decision record exists with promotion rationale.
`, project)
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func loadJSON[T any](path string) (T, error) {
	var out T
	f, err := os.Open(path)
	if err != nil {
		return out, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		return out, err
	}
	return out, nil
}
