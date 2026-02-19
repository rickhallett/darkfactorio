package factoryv05

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"
)

type Bundle struct {
	SpecExecPath          string `json:"spec_exec_path"`
	HoldoutProvenancePath string `json:"holdout_provenance_path"`
	TwinDriftPath         string `json:"twin_drift_path"`
	DeployEvidencePath    string `json:"deploy_evidence_path"`
	RuntimeSLOPath        string `json:"runtime_slo_path"`
	EconReconcilePath     string `json:"econ_reconcile_path"`
	RedteamPath           string `json:"redteam_path"`
	PolicyChainPath       string `json:"policy_chain_path"`
	PortfolioPath         string `json:"portfolio_path"`
}

type Report struct {
	Passed    bool     `json:"passed"`
	Checks    []string `json:"checks"`
	Failures  []string `json:"failures"`
	BundleRef string   `json:"bundle_ref"`
}

type specExecDoc struct {
	SpecID             string `json:"spec_id"`
	ImplementationRepo string `json:"implementation_repo"`
	ImplementationSHA  string `json:"implementation_sha"`
	Command            string `json:"command"`
	ExitCode           int    `json:"exit_code"`
	ArtifactPath       string `json:"artifact_path"`
}

type holdoutDoc struct {
	HoldoutProducer string `json:"holdout_producer"`
	HoldoutRepo     string `json:"holdout_repo"`
	HoldoutSHA      string `json:"holdout_sha"`
	ResultsPath     string `json:"results_path"`
	ResultsSHA256   string `json:"results_sha256"`
}

type twinService struct {
	Name             string  `json:"name"`
	RealP95LatencyMs float64 `json:"real_p95_latency_ms"`
	TwinP95LatencyMs float64 `json:"twin_p95_latency_ms"`
	MaxDriftPercent  float64 `json:"max_drift_percent"`
}
type twinDriftDoc struct {
	Services []twinService `json:"services"`
}

type deployEvidence struct {
	Environment        string   `json:"environment"`
	CanaryPercent      int      `json:"canary_percent"`
	Promoted           bool     `json:"promoted"`
	RollbackReady      bool     `json:"rollback_ready"`
	RollbackSteps      []string `json:"rollback_steps"`
	RollbackTriggerSLO string   `json:"rollback_trigger_slo"`
}

type runtimeSLODoc struct {
	AvailabilityPercent    float64 `json:"availability_percent"`
	MinAvailabilityPercent float64 `json:"min_availability_percent"`
	ErrorRatePercent       float64 `json:"error_rate_percent"`
	MaxErrorRatePercent    float64 `json:"max_error_rate_percent"`
	P95LatencyMs           float64 `json:"p95_latency_ms"`
	MaxP95LatencyMs        float64 `json:"max_p95_latency_ms"`
}

type econReconcileDoc struct {
	ProviderCostUSD  float64 `json:"provider_cost_usd"`
	InternalCostUSD  float64 `json:"internal_cost_usd"`
	MaxDeltaPercent  float64 `json:"max_delta_percent"`
	ProviderTokens   int64   `json:"provider_tokens"`
	InternalTokens   int64   `json:"internal_tokens"`
	MaxTokenDeltaPct float64 `json:"max_token_delta_percent"`
}

type redteamCase struct {
	ID                string `json:"id"`
	ExpectedDetection bool   `json:"expected_detection"`
	Detected          bool   `json:"detected"`
}
type redteamDoc struct {
	Cases                   []redteamCase `json:"cases"`
	MinDetectionRatePercent float64       `json:"min_detection_rate_percent"`
}

type policyEntry struct {
	Index     int    `json:"index"`
	Timestamp string `json:"timestamp"`
	Actor     string `json:"actor"`
	Payload   string `json:"payload"`
	PrevHash  string `json:"prev_hash"`
	Hash      string `json:"hash"`
}
type policyChainDoc struct {
	Entries []policyEntry `json:"entries"`
}

type portfolioProject struct {
	Name           string  `json:"name"`
	ValueScore     float64 `json:"value_score"`
	RiskScore      float64 `json:"risk_score"`
	ReadinessScore float64 `json:"readiness_score"`
	ExpectedHours  float64 `json:"expected_hours"`
	PriorityScore  float64 `json:"priority_score"`
}
type portfolioDoc struct {
	Projects []portfolioProject `json:"projects"`
}

func ValidateBundle(root string, bundlePath string) (Report, error) {
	if root == "" {
		root = "."
	}
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

	check("spec-exec", func() error { return validateSpecExec(root, b.SpecExecPath) })
	check("holdout-provenance", func() error { return validateHoldout(root, b.HoldoutProvenancePath) })
	check("twin-drift", func() error { return validateTwinDrift(root, b.TwinDriftPath) })
	check("deploy-evidence", func() error { return validateDeploy(root, b.DeployEvidencePath) })
	check("runtime-slo", func() error { return validateRuntimeSLO(root, b.RuntimeSLOPath) })
	check("econ-reconcile", func() error { return validateEcon(root, b.EconReconcilePath) })
	check("redteam", func() error { return validateRedteam(root, b.RedteamPath) })
	check("policy-chain", func() error { return validatePolicyChain(root, b.PolicyChainPath) })
	check("portfolio", func() error { return validatePortfolio(root, b.PortfolioPath) })

	return r, nil
}

func validateSpecExec(root, p string) error {
	d, err := loadJSON[specExecDoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if d.SpecID == "" || d.ImplementationRepo == "" || d.ImplementationSHA == "" || d.Command == "" {
		return fmt.Errorf("missing required spec execution fields")
	}
	if d.ExitCode != 0 {
		return fmt.Errorf("implementation command exit code must be 0")
	}
	if _, err := os.Stat(filepath.Join(root, d.ArtifactPath)); err != nil {
		return fmt.Errorf("artifact missing: %w", err)
	}
	return nil
}

func validateHoldout(root, p string) error {
	d, err := loadJSON[holdoutDoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if d.HoldoutProducer == "" || d.HoldoutRepo == "" || d.HoldoutSHA == "" {
		return fmt.Errorf("missing holdout provenance fields")
	}
	b, err := os.ReadFile(filepath.Join(root, d.ResultsPath))
	if err != nil {
		return err
	}
	sum := sha256.Sum256(b)
	got := hex.EncodeToString(sum[:])
	if got != d.ResultsSHA256 {
		return fmt.Errorf("results sha mismatch")
	}
	return nil
}

func validateTwinDrift(root, p string) error {
	d, err := loadJSON[twinDriftDoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if len(d.Services) == 0 {
		return fmt.Errorf("no twin services")
	}
	for _, s := range d.Services {
		if s.Name == "" || s.RealP95LatencyMs <= 0 || s.TwinP95LatencyMs <= 0 {
			return fmt.Errorf("invalid twin latency entry")
		}
		drift := abs(s.RealP95LatencyMs-s.TwinP95LatencyMs) / s.RealP95LatencyMs * 100
		if drift > s.MaxDriftPercent {
			return fmt.Errorf("service %s drift %.2f > %.2f", s.Name, drift, s.MaxDriftPercent)
		}
	}
	return nil
}

func validateDeploy(root, p string) error {
	d, err := loadJSON[deployEvidence](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if d.Environment == "" || d.RollbackTriggerSLO == "" {
		return fmt.Errorf("missing deploy fields")
	}
	if d.CanaryPercent <= 0 || d.CanaryPercent > 100 {
		return fmt.Errorf("invalid canary_percent")
	}
	if !d.Promoted || !d.RollbackReady || len(d.RollbackSteps) < 3 {
		return fmt.Errorf("deploy evidence incomplete")
	}
	return nil
}

func validateRuntimeSLO(root, p string) error {
	d, err := loadJSON[runtimeSLODoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if d.AvailabilityPercent < d.MinAvailabilityPercent {
		return fmt.Errorf("availability below threshold")
	}
	if d.ErrorRatePercent > d.MaxErrorRatePercent {
		return fmt.Errorf("error rate above threshold")
	}
	if d.P95LatencyMs > d.MaxP95LatencyMs {
		return fmt.Errorf("latency above threshold")
	}
	return nil
}

func validateEcon(root, p string) error {
	d, err := loadJSON[econReconcileDoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if d.ProviderCostUSD <= 0 || d.ProviderTokens <= 0 {
		return fmt.Errorf("provider values must be > 0")
	}
	costDelta := abs(d.ProviderCostUSD-d.InternalCostUSD) / d.ProviderCostUSD * 100
	if costDelta > d.MaxDeltaPercent {
		return fmt.Errorf("cost delta %.2f > %.2f", costDelta, d.MaxDeltaPercent)
	}
	tokenDelta := abs(float64(d.ProviderTokens-d.InternalTokens)) / float64(d.ProviderTokens) * 100
	if tokenDelta > d.MaxTokenDeltaPct {
		return fmt.Errorf("token delta %.2f > %.2f", tokenDelta, d.MaxTokenDeltaPct)
	}
	return nil
}

func validateRedteam(root, p string) error {
	d, err := loadJSON[redteamDoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if len(d.Cases) == 0 {
		return fmt.Errorf("no redteam cases")
	}
	exp := 0
	hit := 0
	for _, c := range d.Cases {
		if c.ExpectedDetection {
			exp++
			if c.Detected {
				hit++
			}
		}
	}
	if exp == 0 {
		return fmt.Errorf("no expected_detection=true cases")
	}
	rate := float64(hit) / float64(exp) * 100
	if rate < d.MinDetectionRatePercent {
		return fmt.Errorf("redteam detection %.2f < %.2f", rate, d.MinDetectionRatePercent)
	}
	return nil
}

func validatePolicyChain(root, p string) error {
	d, err := loadJSON[policyChainDoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if len(d.Entries) < 2 {
		return fmt.Errorf("need >=2 chain entries")
	}
	prev := "GENESIS"
	for i, e := range d.Entries {
		if e.Index != i {
			return fmt.Errorf("non-sequential policy index")
		}
		if _, err := time.Parse(time.RFC3339, e.Timestamp); err != nil {
			return fmt.Errorf("invalid timestamp")
		}
		if e.Actor == "" || e.Payload == "" {
			return fmt.Errorf("actor/payload required")
		}
		if e.PrevHash != prev {
			return fmt.Errorf("prev hash mismatch at index %d", i)
		}
		sum := sha256.Sum256([]byte(fmt.Sprintf("%d|%s|%s|%s|%s", e.Index, e.Timestamp, e.Actor, e.Payload, e.PrevHash)))
		want := hex.EncodeToString(sum[:])
		if e.Hash != want {
			return fmt.Errorf("hash mismatch at index %d", i)
		}
		prev = e.Hash
	}
	return nil
}

func validatePortfolio(root, p string) error {
	d, err := loadJSON[portfolioDoc](filepath.Join(root, p))
	if err != nil {
		return err
	}
	if len(d.Projects) < 2 {
		return fmt.Errorf("need >=2 projects")
	}
	prev := 1e18
	for _, pr := range d.Projects {
		if pr.Name == "" || pr.ExpectedHours <= 0 {
			return fmt.Errorf("invalid project entry")
		}
		expected := (pr.ValueScore * pr.ReadinessScore) / (pr.RiskScore * pr.ExpectedHours)
		if abs(expected-pr.PriorityScore) > 0.0001 {
			return fmt.Errorf("priority score mismatch for %s", pr.Name)
		}
		if pr.PriorityScore > prev {
			return fmt.Errorf("projects not sorted by priority_score desc")
		}
		prev = pr.PriorityScore
	}
	return nil
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
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

func BuildPolicyChainEntries(entries []policyEntry) []policyEntry {
	prev := "GENESIS"
	out := make([]policyEntry, 0, len(entries))
	for i, e := range entries {
		e.Index = i
		e.PrevHash = prev
		sum := sha256.Sum256([]byte(fmt.Sprintf("%d|%s|%s|%s|%s", e.Index, e.Timestamp, e.Actor, e.Payload, e.PrevHash)))
		e.Hash = hex.EncodeToString(sum[:])
		prev = e.Hash
		out = append(out, e)
	}
	return out
}

func SortedProjectsByPriority(in []portfolioProject) []portfolioProject {
	out := append([]portfolioProject{}, in...)
	for i := range out {
		out[i].PriorityScore = (out[i].ValueScore * out[i].ReadinessScore) / (out[i].RiskScore * out[i].ExpectedHours)
	}
	slices.SortFunc(out, func(a, b portfolioProject) int {
		if a.PriorityScore > b.PriorityScore {
			return -1
		}
		if a.PriorityScore < b.PriorityScore {
			return 1
		}
		return 0
	})
	return out
}
