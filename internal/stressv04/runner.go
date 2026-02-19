package stressv04

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rickhallett/darkfactorio/internal/dfcorpus"
	"github.com/rickhallett/darkfactorio/internal/dfwindow"
	"github.com/rickhallett/darkfactorio/internal/factoryv04"
	"github.com/rickhallett/darkfactorio/internal/level4gate"
)

type CheckResult struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail"`
}

type Report struct {
	Passed bool          `json:"passed"`
	Checks []CheckResult `json:"checks"`
}

func Run(root string) (Report, error) {
	if root == "" {
		root = "."
	}
	report := Report{Passed: true, Checks: []CheckResult{}}

	add := func(name string, passed bool, detail string) {
		if !passed {
			report.Passed = false
		}
		report.Checks = append(report.Checks, CheckResult{Name: name, Passed: passed, Detail: detail})
	}
	add2 := func(name string, fn func(string) (bool, string)) {
		passed, detail := fn(root)
		add(name, passed, detail)
	}

	add2("data-contract-fuzz", checkDataContractFuzz)
	add2("threshold-boundary", checkThresholdBoundary)
	add2("corpus-degradation", checkCorpusDegradation)
	add2("policy-evidence-break", checkPolicyEvidenceBreak)
	add2("twin-health-chaos", checkTwinHealthChaos)
	add2("release-rollback-integrity", checkReleaseRollback)
	add2("economic-overload", checkEconomicOverload)
	add2("orchestration-cycle", checkOrchestrationCycle)
	add2("quality-guardrail", checkQualityGuardrail)
	add2("autonomy-soak", checkAutonomySoak)

	return report, nil
}

func checkDataContractFuzz(root string) (bool, string) {
	td := mustTemp()
	path := filepath.Join(td, "bad.ndjson")
	body := `{"window_id":"w","run_id":"r1","pipeline_id":"p1","pipeline_class":"low_risk_feature","scenario_total":10,"scenario_passed":10,"first_pass_success":true,"retries":0,"interventions":0,"decision":"approved","decision_reversed":false,"critical_incident":false,"timestamp":"2026-02-19T00:00:00Z","extra":"x"}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return false, err.Error()
	}
	_, err := level4gate.LoadNDJSON(path, "")
	if err == nil {
		return false, "expected unknown field validation failure"
	}
	return true, "malformed record rejected"
}

func checkThresholdBoundary(root string) (bool, string) {
	td := mustTemp()
	path := filepath.Join(td, "w.ndjson")
	recs := make([]level4gate.EvalRecord, 0, 10)
	for i := 0; i < 10; i++ {
		class := "low_risk_feature"
		if i%2 == 1 {
			class = "medium_integration"
		}
		passed := 9
		if i == 0 {
			passed = 10 // exactly 90%
		}
		recs = append(recs, level4gate.EvalRecord{
			WindowID:         "w",
			RunID:            fmt.Sprintf("run-%03d", i+1),
			PipelineID:       "p",
			PipelineClass:    class,
			ScenarioTotal:    10,
			ScenarioPassed:   passed,
			FirstPassSuccess: true,
			Retries:          1,
			Interventions:    1,
			Decision:         "approved",
			DecisionReversed: false,
			CriticalIncident: false,
			Timestamp:        time.Now().UTC().Format(time.RFC3339),
		})
	}
	if err := writeNDJSON(path, recs); err != nil {
		return false, err.Error()
	}
	loaded, err := level4gate.LoadNDJSON(path, "w")
	if err != nil {
		return false, err.Error()
	}
	rep := level4gate.EvaluateWithCriteria(loaded, level4gate.DefaultCriteria(), "w")
	if !rep.Passed {
		return false, "expected pass at exact threshold boundary"
	}
	return true, "boundary pass deterministic"
}

func checkCorpusDegradation(root string) (bool, string) {
	td := mustTemp()
	dst1 := filepath.Join(td, "w1.ndjson")
	dst2 := filepath.Join(td, "w2.ndjson")
	if err := copyFile(filepath.Join(root, "runs/w-2026-02-l4-02.ndjson"), dst1); err != nil {
		return false, err.Error()
	}
	if err := copyFile(filepath.Join(root, "runs/w-2026-02-l4-03.ndjson"), dst2); err != nil {
		return false, err.Error()
	}
	degrade := make([]level4gate.EvalRecord, 0, 8)
	for i := 0; i < 8; i++ {
		class := "low_risk_feature"
		if i%2 == 1 {
			class = "medium_integration"
		}
		degrade = append(degrade, level4gate.EvalRecord{
			WindowID:         "w-2026-02-l4-03",
			RunID:            fmt.Sprintf("run-x-%03d", i+1),
			PipelineID:       "p-x",
			PipelineClass:    class,
			ScenarioTotal:    10,
			ScenarioPassed:   0,
			FirstPassSuccess: false,
			Retries:          2,
			Interventions:    2,
			Decision:         "approved",
			DecisionReversed: false,
			CriticalIncident: false,
			Timestamp:        time.Now().UTC().Format(time.RFC3339),
		})
	}
	if err := appendNDJSON(dst2, degrade); err != nil {
		return false, err.Error()
	}
	criteria, err := loadCriteria(filepath.Join(root, "profiles/level4-gate-v0.1-adversarial.json"))
	if err != nil {
		return false, err.Error()
	}
	res, err := dfcorpus.Replay(dfcorpus.ReplayOptions{
		Inputs:   []string{dst1, dst2},
		Criteria: criteria,
	})
	if err != nil {
		return false, err.Error()
	}
	if res.Report.Passed {
		return false, "expected degraded corpus to fail"
	}
	return true, "degraded corpus fails as expected"
}

func checkPolicyEvidenceBreak(root string) (bool, string) {
	td := mustTemp()
	bundlePath, err := copyV04Example(root, td)
	if err != nil {
		return false, err.Error()
	}
	policyPath := filepath.Join(td, "factory/v0.4/examples/policy.json")
	var doc map[string]any
	if err := readJSON(policyPath, &doc); err != nil {
		return false, err.Error()
	}
	att := doc["attestations"].([]any)
	first := att[0].(map[string]any)
	first["evidence"] = []any{"factory/v0.4/examples/missing-evidence.txt"}
	if err := writeJSON(policyPath, doc); err != nil {
		return false, err.Error()
	}
	rep, err := factoryv04.ValidateBundle(td, relFrom(td, bundlePath))
	if err != nil {
		return false, err.Error()
	}
	if rep.Passed {
		return false, "expected policy evidence failure"
	}
	return true, "policy evidence break detected"
}

func checkTwinHealthChaos(root string) (bool, string) {
	td := mustTemp()
	bundlePath, err := copyV04Example(root, td)
	if err != nil {
		return false, err.Error()
	}
	path := filepath.Join(td, "factory/v0.4/examples/twins.json")
	var doc map[string]any
	if err := readJSON(path, &doc); err != nil {
		return false, err.Error()
	}
	services := doc["services"].([]any)
	s0 := services[0].(map[string]any)
	s0["healthy"] = false
	if err := writeJSON(path, doc); err != nil {
		return false, err.Error()
	}
	rep, err := factoryv04.ValidateBundle(td, relFrom(td, bundlePath))
	if err != nil {
		return false, err.Error()
	}
	if rep.Passed {
		return false, "expected twin health failure"
	}
	return true, "unhealthy twin detected"
}

func checkReleaseRollback(root string) (bool, string) {
	td := mustTemp()
	bundlePath, err := copyV04Example(root, td)
	if err != nil {
		return false, err.Error()
	}
	path := filepath.Join(td, "factory/v0.4/examples/release.json")
	var doc map[string]any
	if err := readJSON(path, &doc); err != nil {
		return false, err.Error()
	}
	doc["rollback_steps"] = []any{"one"}
	if err := writeJSON(path, doc); err != nil {
		return false, err.Error()
	}
	rep, err := factoryv04.ValidateBundle(td, relFrom(td, bundlePath))
	if err != nil {
		return false, err.Error()
	}
	if rep.Passed {
		return false, "expected rollback integrity failure"
	}
	return true, "release rollback integrity enforced"
}

func checkEconomicOverload(root string) (bool, string) {
	td := mustTemp()
	bundlePath, err := copyV04Example(root, td)
	if err != nil {
		return false, err.Error()
	}
	path := filepath.Join(td, "factory/v0.4/examples/econ.json")
	var doc map[string]any
	if err := readJSON(path, &doc); err != nil {
		return false, err.Error()
	}
	doc["cost_observed"] = 999999.0
	if err := writeJSON(path, doc); err != nil {
		return false, err.Error()
	}
	rep, err := factoryv04.ValidateBundle(td, relFrom(td, bundlePath))
	if err != nil {
		return false, err.Error()
	}
	if rep.Passed {
		return false, "expected economic overload failure"
	}
	return true, "economic budget breach detected"
}

func checkOrchestrationCycle(root string) (bool, string) {
	td := mustTemp()
	bundlePath, err := copyV04Example(root, td)
	if err != nil {
		return false, err.Error()
	}
	path := filepath.Join(td, "factory/v0.4/examples/orchestration.json")
	var doc map[string]any
	if err := readJSON(path, &doc); err != nil {
		return false, err.Error()
	}
	doc["stages"] = []any{
		map[string]any{"id": "validation", "depends_on": []any{"build"}},
		map[string]any{"id": "build", "depends_on": []any{"validation"}},
	}
	if err := writeJSON(path, doc); err != nil {
		return false, err.Error()
	}
	rep, err := factoryv04.ValidateBundle(td, relFrom(td, bundlePath))
	if err != nil {
		return false, err.Error()
	}
	if rep.Passed {
		return false, "expected orchestration cycle failure"
	}
	return true, "orchestration cycle detected"
}

func checkQualityGuardrail(root string) (bool, string) {
	td := mustTemp()
	if err := os.MkdirAll(filepath.Join(td, "profiles"), 0o755); err != nil {
		return false, err.Error()
	}
	if err := copyFile(filepath.Join(root, "profiles/level4-gate-v0.1-baseline.json"), filepath.Join(td, "profiles/level4-gate-v0.1-baseline.json")); err != nil {
		return false, err.Error()
	}
	if err := copyFile(filepath.Join(root, "profiles/level4-gate-v0.1-adversarial.json"), filepath.Join(td, "profiles/level4-gate-v0.1-adversarial.json")); err != nil {
		return false, err.Error()
	}

	_, err := dfwindow.Advance(dfwindow.AdvanceOptions{
		Root:                td,
		WindowID:            "w-guard",
		RunsPath:            "runs/w-guard.ndjson",
		AppendCount:         1,
		BaselineCriteria:    "profiles/level4-gate-v0.1-baseline.json",
		AdversarialCriteria: "profiles/level4-gate-v0.1-adversarial.json",
		LogLearning:         false,
		QualityMode:         "high",
		QualityReason:       "",
	})
	if err == nil {
		return false, "expected high-quality mode without reason to fail"
	}
	_, err = dfwindow.Advance(dfwindow.AdvanceOptions{
		Root:                td,
		WindowID:            "w-guard",
		RunsPath:            "runs/w-guard.ndjson",
		AppendCount:         1,
		BaselineCriteria:    "profiles/level4-gate-v0.1-baseline.json",
		AdversarialCriteria: "profiles/level4-gate-v0.1-adversarial.json",
		LogLearning:         false,
		QualityMode:         "high",
		QualityReason:       "stress check",
	})
	if err != nil {
		return false, fmt.Sprintf("expected quality mode with reason to pass: %v", err)
	}
	return true, "quality guardrail enforced"
}

func checkAutonomySoak(root string) (bool, string) {
	td := mustTemp()
	dst := filepath.Join(td, "w-soak.ndjson")
	if err := copyFile(filepath.Join(root, "runs/w-2026-02-l4-03.ndjson"), dst); err != nil {
		return false, err.Error()
	}
	startCount := 0
	if recs, err := level4gate.LoadNDJSON(dst, "w-2026-02-l4-03"); err == nil {
		startCount = len(recs)
	}
	for i := 0; i < 12; i++ {
		modeHigh := i%4 == 3
		recs := generateRecords("w-2026-02-l4-03", startCount+i*2+1, 2, modeHigh)
		if err := appendNDJSON(dst, recs); err != nil {
			return false, err.Error()
		}
	}
	loaded, err := level4gate.LoadNDJSON(dst, "w-2026-02-l4-03")
	if err != nil {
		return false, err.Error()
	}
	if len(loaded) < startCount+24 {
		return false, "soak did not append expected records"
	}
	return true, "soak append stability confirmed"
}

// helpers

func mustTemp() string {
	d, _ := os.MkdirTemp("", "df-stress-")
	return d
}

func copyV04Example(root, dstRoot string) (string, error) {
	src := filepath.Join(root, "factory/v0.4/examples")
	dst := filepath.Join(dstRoot, "factory/v0.4/examples")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return "", err
	}
	ents, err := os.ReadDir(src)
	if err != nil {
		return "", err
	}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		if err := copyFile(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
			return "", err
		}
	}
	return filepath.Join(dstRoot, "factory/v0.4/examples/bundle.json"), nil
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

func relFrom(root, abs string) string {
	r, err := filepath.Rel(root, abs)
	if err != nil {
		return abs
	}
	return r
}

func loadCriteria(path string) (level4gate.Criteria, error) {
	var c level4gate.Criteria
	if err := readJSON(path, &c); err != nil {
		return c, err
	}
	return c, nil
}

func readJSON(path string, out any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	return dec.Decode(out)
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func writeNDJSON(path string, recs []level4gate.EvalRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, r := range recs {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return nil
}

func appendNDJSON(path string, recs []level4gate.EvalRecord) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, r := range recs {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return nil
}

func generateRecords(window string, start int, n int, high bool) []level4gate.EvalRecord {
	out := make([]level4gate.EvalRecord, 0, n)
	for i := 0; i < n; i++ {
		idx := start + i
		class := "low_risk_feature"
		prefix := "p-low"
		if idx%2 == 0 {
			class = "medium_integration"
			prefix = "p-med"
		}
		total := 10 + (idx % 3)
		passed := total - 1
		if high {
			passed = total
		}
		out = append(out, level4gate.EvalRecord{
			WindowID:         window,
			RunID:            fmt.Sprintf("run-%03d", idx),
			PipelineID:       fmt.Sprintf("%s-soak-%03d", prefix, idx),
			PipelineClass:    class,
			ScenarioTotal:    total,
			ScenarioPassed:   passed,
			FirstPassSuccess: true,
			Retries:          1,
			Interventions:    1,
			Decision:         "approved",
			DecisionReversed: false,
			CriticalIncident: false,
			Timestamp:        time.Now().UTC().Format(time.RFC3339),
		})
	}
	return out
}
