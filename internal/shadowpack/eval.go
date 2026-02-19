package shadowpack

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
)

type Criteria struct {
	MinOverlap                    int     `json:"min_overlap"`
	MaxOutcomeMismatchRatePercent float64 `json:"max_outcome_mismatch_rate_percent"`
	MaxP95LatencyDriftPercent     float64 `json:"max_p95_latency_drift_percent"`
}

type Manifest struct {
	PackID            string   `json:"pack_id"`
	CandidateResults  string   `json:"candidate_results"`
	HoldoutResults    string   `json:"holdout_results"`
	CandidateProducer string   `json:"candidate_producer"`
	HoldoutProducer   string   `json:"holdout_producer"`
	Criteria          Criteria `json:"criteria"`
}

type ScenarioResult struct {
	ScenarioID string  `json:"scenario_id"`
	Outcome    string  `json:"outcome"` // pass|fail|error
	LatencyMs  float64 `json:"latency_ms"`
}

type Report struct {
	PackID                     string   `json:"pack_id"`
	Passed                     bool     `json:"passed"`
	OverlapCount               int      `json:"overlap_count"`
	CandidateOnlyCount         int      `json:"candidate_only_count"`
	HoldoutOnlyCount           int      `json:"holdout_only_count"`
	OutcomeMismatchCount       int      `json:"outcome_mismatch_count"`
	OutcomeMismatchRatePercent float64  `json:"outcome_mismatch_rate_percent"`
	CandidateP95LatencyMs      float64  `json:"candidate_p95_latency_ms"`
	HoldoutP95LatencyMs        float64  `json:"holdout_p95_latency_ms"`
	P95LatencyDriftPercent     float64  `json:"p95_latency_drift_percent"`
	Failures                   []string `json:"failures"`
}

func Evaluate(root, manifestPath string) (Report, error) {
	if root == "" {
		root = "."
	}
	m, err := loadJSON[Manifest](filepath.Join(root, manifestPath))
	if err != nil {
		return Report{}, err
	}
	r := Report{
		PackID:   m.PackID,
		Passed:   true,
		Failures: []string{},
	}

	if m.CandidateProducer == "" || m.HoldoutProducer == "" {
		return Report{}, fmt.Errorf("candidate_producer and holdout_producer are required")
	}
	if m.CandidateProducer == m.HoldoutProducer {
		r.Passed = false
		r.Failures = append(r.Failures, "candidate_producer must differ from holdout_producer")
	}

	cand, err := loadJSON[[]ScenarioResult](filepath.Join(root, m.CandidateResults))
	if err != nil {
		return Report{}, err
	}
	hold, err := loadJSON[[]ScenarioResult](filepath.Join(root, m.HoldoutResults))
	if err != nil {
		return Report{}, err
	}

	cMap, err := indexByID(cand)
	if err != nil {
		return Report{}, err
	}
	hMap, err := indexByID(hold)
	if err != nil {
		return Report{}, err
	}

	overlapIDs := make([]string, 0)
	for id := range cMap {
		if _, ok := hMap[id]; ok {
			overlapIDs = append(overlapIDs, id)
		}
	}
	r.OverlapCount = len(overlapIDs)
	r.CandidateOnlyCount = len(cMap) - r.OverlapCount
	r.HoldoutOnlyCount = len(hMap) - r.OverlapCount

	mismatches := 0
	cLat := make([]float64, 0, len(overlapIDs))
	hLat := make([]float64, 0, len(overlapIDs))
	for _, id := range overlapIDs {
		cr := cMap[id]
		hr := hMap[id]
		if !validOutcome(cr.Outcome) || !validOutcome(hr.Outcome) {
			return Report{}, fmt.Errorf("invalid outcome on scenario_id=%s", id)
		}
		if cr.Outcome != hr.Outcome {
			mismatches++
		}
		cLat = append(cLat, cr.LatencyMs)
		hLat = append(hLat, hr.LatencyMs)
	}

	r.OutcomeMismatchCount = mismatches
	if r.OverlapCount > 0 {
		r.OutcomeMismatchRatePercent = float64(mismatches) / float64(r.OverlapCount) * 100
	}
	r.CandidateP95LatencyMs = p95(cLat)
	r.HoldoutP95LatencyMs = p95(hLat)
	if r.HoldoutP95LatencyMs > 0 {
		r.P95LatencyDriftPercent = math.Abs(r.CandidateP95LatencyMs-r.HoldoutP95LatencyMs) / r.HoldoutP95LatencyMs * 100
	}

	if r.OverlapCount < m.Criteria.MinOverlap {
		r.Passed = false
		r.Failures = append(r.Failures, fmt.Sprintf("overlap_count %d < %d", r.OverlapCount, m.Criteria.MinOverlap))
	}
	if r.OutcomeMismatchRatePercent > m.Criteria.MaxOutcomeMismatchRatePercent {
		r.Passed = false
		r.Failures = append(r.Failures, fmt.Sprintf("outcome_mismatch_rate %.2f > %.2f", r.OutcomeMismatchRatePercent, m.Criteria.MaxOutcomeMismatchRatePercent))
	}
	if r.P95LatencyDriftPercent > m.Criteria.MaxP95LatencyDriftPercent {
		r.Passed = false
		r.Failures = append(r.Failures, fmt.Sprintf("p95_latency_drift %.2f > %.2f", r.P95LatencyDriftPercent, m.Criteria.MaxP95LatencyDriftPercent))
	}
	return r, nil
}

func indexByID(in []ScenarioResult) (map[string]ScenarioResult, error) {
	out := make(map[string]ScenarioResult, len(in))
	for _, r := range in {
		if r.ScenarioID == "" {
			return nil, fmt.Errorf("scenario_id is required")
		}
		if _, ok := out[r.ScenarioID]; ok {
			return nil, fmt.Errorf("duplicate scenario_id %q", r.ScenarioID)
		}
		if r.LatencyMs < 0 {
			return nil, fmt.Errorf("latency_ms cannot be negative")
		}
		out[r.ScenarioID] = r
	}
	return out, nil
}

func validOutcome(v string) bool {
	return v == "pass" || v == "fail" || v == "error"
}

func p95(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	cp := append([]float64{}, vals...)
	sort.Float64s(cp)
	idx := int(math.Ceil(0.95*float64(len(cp)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cp) {
		idx = len(cp) - 1
	}
	return cp[idx]
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
