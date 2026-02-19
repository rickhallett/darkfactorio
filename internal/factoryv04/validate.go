package factoryv04

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"
)

type Bundle struct {
	SpecPath          string `json:"spec_path"`
	HoldoutPath       string `json:"holdout_path"`
	TwinsPath         string `json:"twins_path"`
	ReleasePath       string `json:"release_path"`
	PolicyPath        string `json:"policy_path"`
	EconPath          string `json:"econ_path"`
	OrchestrationPath string `json:"orchestration_path"`
}

type Report struct {
	Passed    bool     `json:"passed"`
	Checks    []string `json:"checks"`
	Failures  []string `json:"failures"`
	BundleRef string   `json:"bundle_ref"`
}

type specDoc struct {
	Title          string   `json:"title"`
	Objective      string   `json:"objective"`
	NonNegotiables []string `json:"non_negotiables"`
	Acceptance     []string `json:"acceptance"`
}

type holdoutDoc struct {
	ScenarioTotal   int  `json:"scenario_total"`
	ScenarioPassed  int  `json:"scenario_passed"`
	HiddenFromAgent bool `json:"hidden_from_agent"`
}

type twinService struct {
	Name            string `json:"name"`
	Mode            string `json:"mode"`
	ContractVersion string `json:"contract_version"`
	Healthy         bool   `json:"healthy"`
	FailurePolicy   string `json:"failure_policy"`
}
type twinsDoc struct {
	Services []twinService `json:"services"`
}

type releaseDoc struct {
	CandidateID     string   `json:"candidate_id"`
	ArtifactPath    string   `json:"artifact_path"`
	BaselinePass    bool     `json:"baseline_pass"`
	AdversarialPass bool     `json:"adversarial_pass"`
	HoldoutPass     bool     `json:"holdout_pass"`
	PolicyPass      bool     `json:"policy_pass"`
	EconPass        bool     `json:"econ_pass"`
	RollbackSteps   []string `json:"rollback_steps"`
}

type attestation struct {
	ControlID string   `json:"control_id"`
	Owner     string   `json:"owner"`
	Timestamp string   `json:"timestamp"`
	Evidence  []string `json:"evidence"`
}
type policyDoc struct {
	RequiredControls []string      `json:"required_controls"`
	Attestations     []attestation `json:"attestations"`
}

type econDoc struct {
	TokenBudgetPerDay float64 `json:"token_budget_per_day"`
	TokenObserved     float64 `json:"token_observed"`
	CostBudgetPerDay  float64 `json:"cost_budget_per_day"`
	CostObserved      float64 `json:"cost_observed"`
	P95LatencyMsMax   float64 `json:"p95_latency_ms_max"`
	P95LatencyMs      float64 `json:"p95_latency_ms"`
}

type agentDoc struct {
	Name string `json:"name"`
	Role string `json:"role"`
}
type stageDoc struct {
	ID        string   `json:"id"`
	DependsOn []string `json:"depends_on"`
}
type orchestrationDoc struct {
	Agents []agentDoc `json:"agents"`
	Stages []stageDoc `json:"stages"`
}

func ValidateBundle(root string, bundlePath string) (Report, error) {
	b, err := loadJSON[Bundle](filepath.Join(root, bundlePath))
	if err != nil {
		return Report{}, err
	}
	r := Report{Passed: true, Checks: []string{}, Failures: []string{}, BundleRef: bundlePath}

	check := func(name string, fn func() error) {
		if err := fn(); err != nil {
			r.Passed = false
			r.Failures = append(r.Failures, fmt.Sprintf("%s: %v", name, err))
			return
		}
		r.Checks = append(r.Checks, name)
	}

	check("spec", func() error { return validateSpec(root, b.SpecPath) })
	check("holdout", func() error { return validateHoldout(root, b.HoldoutPath) })
	check("twins", func() error { return validateTwins(root, b.TwinsPath) })
	check("release", func() error { return validateRelease(root, b.ReleasePath) })
	check("policy", func() error { return validatePolicy(root, b.PolicyPath) })
	check("economics", func() error { return validateEcon(root, b.EconPath) })
	check("orchestration", func() error { return validateOrchestration(root, b.OrchestrationPath) })

	return r, nil
}

func validateSpec(root string, p string) error {
	d, err := loadJSON[specDoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if d.Title == "" || d.Objective == "" {
		return fmt.Errorf("title/objective required")
	}
	if len(d.NonNegotiables) < 3 {
		return fmt.Errorf("need >=3 non_negotiables")
	}
	if len(d.Acceptance) < 3 {
		return fmt.Errorf("need >=3 acceptance statements")
	}
	return nil
}

func validateHoldout(root string, p string) error {
	d, err := loadJSON[holdoutDoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if !d.HiddenFromAgent {
		return fmt.Errorf("holdout must be hidden_from_agent=true")
	}
	if d.ScenarioTotal < 7 || d.ScenarioPassed > d.ScenarioTotal {
		return fmt.Errorf("invalid scenario totals")
	}
	pass := float64(d.ScenarioPassed) / float64(d.ScenarioTotal) * 100
	if pass < 90 {
		return fmt.Errorf("scenario pass rate %.2f < 90", pass)
	}
	return nil
}

func validateTwins(root string, p string) error {
	d, err := loadJSON[twinsDoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if len(d.Services) < 2 {
		return fmt.Errorf("need >=2 twin services")
	}
	for _, s := range d.Services {
		if s.Name == "" || s.ContractVersion == "" || s.FailurePolicy == "" {
			return fmt.Errorf("twin service fields required")
		}
		if s.Mode != "simulated" && s.Mode != "hybrid" {
			return fmt.Errorf("invalid twin mode %q", s.Mode)
		}
		if !s.Healthy {
			return fmt.Errorf("twin %q not healthy", s.Name)
		}
	}
	return nil
}

func validateRelease(root string, p string) error {
	d, err := loadJSON[releaseDoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if d.CandidateID == "" {
		return fmt.Errorf("candidate_id required")
	}
	if len(d.RollbackSteps) < 3 {
		return fmt.Errorf("need >=3 rollback steps")
	}
	if !d.BaselinePass || !d.AdversarialPass || !d.HoldoutPass || !d.PolicyPass || !d.EconPass {
		return fmt.Errorf("all release gates must pass")
	}
	if _, err := os.Stat(filepath.Join(root, d.ArtifactPath)); err != nil {
		return fmt.Errorf("artifact_path missing: %w", err)
	}
	return nil
}

func validatePolicy(root string, p string) error {
	d, err := loadJSON[policyDoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if len(d.RequiredControls) == 0 {
		return fmt.Errorf("required_controls cannot be empty")
	}
	got := map[string]bool{}
	for _, a := range d.Attestations {
		if a.ControlID == "" || a.Owner == "" || len(a.Evidence) == 0 {
			return fmt.Errorf("invalid attestation entry")
		}
		if _, err := time.Parse(time.RFC3339, a.Timestamp); err != nil {
			return fmt.Errorf("invalid attestation timestamp")
		}
		for _, ev := range a.Evidence {
			if _, err := os.Stat(filepath.Join(root, ev)); err != nil {
				return fmt.Errorf("missing policy evidence %q", ev)
			}
		}
		got[a.ControlID] = true
	}
	for _, c := range d.RequiredControls {
		if !got[c] {
			return fmt.Errorf("missing attestation for control %q", c)
		}
	}
	return nil
}

func validateEcon(root string, p string) error {
	d, err := loadJSON[econDoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if d.TokenObserved > d.TokenBudgetPerDay {
		return fmt.Errorf("token budget exceeded")
	}
	if d.CostObserved > d.CostBudgetPerDay {
		return fmt.Errorf("cost budget exceeded")
	}
	if d.P95LatencyMs > d.P95LatencyMsMax {
		return fmt.Errorf("latency budget exceeded")
	}
	return nil
}

func validateOrchestration(root string, p string) error {
	d, err := loadJSON[orchestrationDoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if len(d.Agents) < 2 {
		return fmt.Errorf("need >=2 agents")
	}
	roles := map[string]bool{}
	for _, a := range d.Agents {
		if a.Name == "" || a.Role == "" {
			return fmt.Errorf("agent name/role required")
		}
		if roles[a.Role] {
			return fmt.Errorf("duplicate agent role %q", a.Role)
		}
		roles[a.Role] = true
	}
	stageIDs := map[string]bool{}
	hasValidation := false
	for _, s := range d.Stages {
		if s.ID == "" {
			return fmt.Errorf("stage id required")
		}
		stageIDs[s.ID] = true
		if s.ID == "validation" {
			hasValidation = true
		}
	}
	if !hasValidation {
		return fmt.Errorf("missing required stage 'validation'")
	}
	for _, s := range d.Stages {
		for _, dep := range s.DependsOn {
			if !stageIDs[dep] {
				return fmt.Errorf("stage %q depends on unknown %q", s.ID, dep)
			}
		}
	}
	if hasCycle(d.Stages) {
		return fmt.Errorf("stage graph has cycle")
	}
	return nil
}

func hasCycle(stages []stageDoc) bool {
	graph := map[string][]string{}
	for _, s := range stages {
		graph[s.ID] = append([]string{}, s.DependsOn...)
	}
	visited := map[string]int{} // 0 unvisited, 1 visiting, 2 done
	var visit func(string) bool
	visit = func(n string) bool {
		if visited[n] == 1 {
			return true
		}
		if visited[n] == 2 {
			return false
		}
		visited[n] = 1
		for _, dep := range graph[n] {
			if visit(dep) {
				return true
			}
		}
		visited[n] = 2
		return false
	}
	keys := make([]string, 0, len(graph))
	for k := range graph {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		if visit(k) {
			return true
		}
	}
	return false
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
