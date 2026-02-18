package level4gate

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

type EvalRecord struct {
	WindowID         string `json:"window_id"`
	RunID            string `json:"run_id"`
	PipelineID       string `json:"pipeline_id"`
	PipelineClass    string `json:"pipeline_class"`
	ScenarioTotal    int    `json:"scenario_total"`
	ScenarioPassed   int    `json:"scenario_passed"`
	FirstPassSuccess bool   `json:"first_pass_success"`
	Retries          int    `json:"retries"`
	Interventions    int    `json:"interventions"`
	Decision         string `json:"decision"`
	DecisionReversed bool   `json:"decision_reversed"`
	CriticalIncident bool   `json:"critical_incident"`
	Timestamp        string `json:"timestamp"`
}

type Thresholds struct {
	MinScenarioPassRatePercent float64 `json:"min_scenario_pass_rate_percent"`
	MinFirstPassRatePercent    float64 `json:"min_first_pass_rate_percent"`
	MaxMeanRetries             float64 `json:"max_mean_retries"`
	MaxDecisionReversalPercent float64 `json:"max_decision_reversal_percent"`
	MaxApprovedIncidents       int     `json:"max_approved_incidents"`
}

type Criteria struct {
	Version              string         `json:"version"`
	MinRuns              int            `json:"min_runs"`
	Thresholds           Thresholds     `json:"thresholds"`
	RequiredClassMinimum map[string]int `json:"required_class_minimum"`
}

type Metrics struct {
	RunCount                       int            `json:"run_count"`
	RunCountByClass                map[string]int `json:"run_count_by_class"`
	ScenarioPassRatePercent        float64        `json:"scenario_pass_rate_percent"`
	FirstPassRatePercent           float64        `json:"first_pass_rate_percent"`
	MeanRetries                    float64        `json:"mean_retries"`
	InterventionAvgFirstHalf       float64        `json:"intervention_avg_first_half"`
	InterventionAvgSecondHalf      float64        `json:"intervention_avg_second_half"`
	InterventionStableOrDecreasing bool           `json:"intervention_stable_or_decreasing"`
	DecisionReversalPercent        float64        `json:"decision_reversal_percent"`
	ApprovedRunCriticalIncidents   int            `json:"approved_run_critical_incidents"`
}

type GateReport struct {
	WindowID   string     `json:"window_id"`
	Thresholds Thresholds `json:"thresholds"`
	Metrics    Metrics    `json:"metrics"`
	Passed     bool       `json:"passed"`
	Failures   []string   `json:"failures"`
}

func DefaultThresholds() Thresholds {
	return Thresholds{
		MinScenarioPassRatePercent: 90.0,
		MinFirstPassRatePercent:    70.0,
		MaxMeanRetries:             2.0,
		MaxDecisionReversalPercent: 5.0,
		MaxApprovedIncidents:       0,
	}
}

func DefaultCriteria() Criteria {
	return Criteria{
		Version:    "level4-gate-v0.1",
		MinRuns:    10,
		Thresholds: DefaultThresholds(),
		RequiredClassMinimum: map[string]int{
			"low_risk_feature":   4,
			"medium_integration": 4,
		},
	}
}

func LoadNDJSON(path string, windowID string) ([]EvalRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return DecodeNDJSON(f, windowID)
}

func DecodeNDJSON(r io.Reader, windowID string) ([]EvalRecord, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	var out []EvalRecord
	line := 0
	for sc.Scan() {
		line++
		raw := sc.Bytes()
		if len(bytes.TrimSpace(raw)) == 0 {
			continue
		}

		rec, err := decodeLine(raw)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		if windowID != "" && rec.WindowID != windowID {
			continue
		}
		if err := validateRecord(rec); err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		out = append(out, rec)
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, errors.New("no records matched filter")
	}
	return out, nil
}

func decodeLine(raw []byte) (EvalRecord, error) {
	var keySet map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keySet); err != nil {
		return EvalRecord{}, err
	}

	requiredKeys := map[string]struct{}{
		"window_id":          {},
		"run_id":             {},
		"pipeline_id":        {},
		"pipeline_class":     {},
		"scenario_total":     {},
		"scenario_passed":    {},
		"first_pass_success": {},
		"retries":            {},
		"interventions":      {},
		"decision":           {},
		"decision_reversed":  {},
		"critical_incident":  {},
		"timestamp":          {},
	}

	for req := range requiredKeys {
		if _, ok := keySet[req]; !ok {
			return EvalRecord{}, fmt.Errorf("missing required field %q", req)
		}
	}
	for k := range keySet {
		if _, ok := requiredKeys[k]; !ok {
			return EvalRecord{}, fmt.Errorf("unknown field %q", k)
		}
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var rec EvalRecord
	if err := dec.Decode(&rec); err != nil {
		return EvalRecord{}, err
	}
	return rec, nil
}

func Evaluate(records []EvalRecord, thresholds Thresholds, windowID string) GateReport {
	criteria := DefaultCriteria()
	criteria.Thresholds = thresholds
	return EvaluateWithCriteria(records, criteria, windowID)
}

func EvaluateWithCriteria(records []EvalRecord, criteria Criteria, windowID string) GateReport {
	m := computeMetrics(records)
	failures := evaluateFailures(m, criteria)
	return GateReport{
		WindowID:   windowID,
		Thresholds: criteria.Thresholds,
		Metrics:    m,
		Passed:     len(failures) == 0,
		Failures:   failures,
	}
}

func computeMetrics(records []EvalRecord) Metrics {
	var scenarioTotal, scenarioPassed int
	var firstPassCount int
	var retriesTotal int
	var interventionsTotal int
	var reversals int
	var approvedCount int
	var approvedIncidents int
	classCounts := map[string]int{}

	firstHalfInterventions := 0
	secondHalfInterventions := 0
	split := len(records) / 2
	if split == 0 {
		split = 1
	}

	for i, r := range records {
		classCounts[r.PipelineClass]++
		scenarioTotal += r.ScenarioTotal
		scenarioPassed += r.ScenarioPassed
		if r.FirstPassSuccess {
			firstPassCount++
		}
		retriesTotal += r.Retries
		interventionsTotal += r.Interventions
		if r.DecisionReversed {
			reversals++
		}
		if r.Decision == "approved" {
			approvedCount++
			if r.CriticalIncident {
				approvedIncidents++
			}
		}

		if i < split {
			firstHalfInterventions += r.Interventions
		} else {
			secondHalfInterventions += r.Interventions
		}
	}

	runCount := len(records)
	secondHalfCount := runCount - split
	if secondHalfCount == 0 {
		secondHalfCount = 1
	}
	approvedDenominator := approvedCount
	if approvedDenominator == 0 {
		approvedDenominator = 1
	}

	scenarioPassRate := 0.0
	if scenarioTotal > 0 {
		scenarioPassRate = pct(float64(scenarioPassed), float64(scenarioTotal))
	}

	return Metrics{
		RunCount:                       runCount,
		RunCountByClass:                classCounts,
		ScenarioPassRatePercent:        scenarioPassRate,
		FirstPassRatePercent:           pct(float64(firstPassCount), float64(runCount)),
		MeanRetries:                    float64(retriesTotal) / float64(runCount),
		InterventionAvgFirstHalf:       float64(firstHalfInterventions) / float64(split),
		InterventionAvgSecondHalf:      float64(secondHalfInterventions) / float64(secondHalfCount),
		InterventionStableOrDecreasing: float64(secondHalfInterventions)/float64(secondHalfCount) <= float64(firstHalfInterventions)/float64(split),
		DecisionReversalPercent:        pct(float64(reversals), float64(approvedDenominator)),
		ApprovedRunCriticalIncidents:   approvedIncidents,
	}
}

func evaluateFailures(m Metrics, c Criteria) []string {
	t := c.Thresholds
	var failures []string
	if m.RunCount < c.MinRuns {
		failures = append(failures, fmt.Sprintf("run_count below minimum window (need >= %d)", c.MinRuns))
	}
	for className, min := range c.RequiredClassMinimum {
		got := m.RunCountByClass[className]
		if got < min {
			failures = append(failures, fmt.Sprintf("run_count_by_class[%s] %d < %d", className, got, min))
		}
	}
	if m.ScenarioPassRatePercent < t.MinScenarioPassRatePercent {
		failures = append(failures, fmt.Sprintf("scenario_pass_rate %.2f < %.2f", m.ScenarioPassRatePercent, t.MinScenarioPassRatePercent))
	}
	if m.FirstPassRatePercent < t.MinFirstPassRatePercent {
		failures = append(failures, fmt.Sprintf("first_pass_rate %.2f < %.2f", m.FirstPassRatePercent, t.MinFirstPassRatePercent))
	}
	if m.MeanRetries > t.MaxMeanRetries {
		failures = append(failures, fmt.Sprintf("mean_retries %.2f > %.2f", m.MeanRetries, t.MaxMeanRetries))
	}
	if !m.InterventionStableOrDecreasing {
		failures = append(failures, "intervention trend increased in second half")
	}
	if m.DecisionReversalPercent > t.MaxDecisionReversalPercent {
		failures = append(failures, fmt.Sprintf("decision_reversal_rate %.2f > %.2f", m.DecisionReversalPercent, t.MaxDecisionReversalPercent))
	}
	if m.ApprovedRunCriticalIncidents > t.MaxApprovedIncidents {
		failures = append(failures, fmt.Sprintf("approved_run_critical_incidents %d > %d", m.ApprovedRunCriticalIncidents, t.MaxApprovedIncidents))
	}
	return failures
}

func validateRecord(r EvalRecord) error {
	if r.WindowID == "" {
		return errors.New("window_id is required")
	}
	if r.RunID == "" {
		return errors.New("run_id is required")
	}
	if r.PipelineID == "" {
		return errors.New("pipeline_id is required")
	}
	switch r.PipelineClass {
	case "low_risk_feature", "medium_integration":
	default:
		return errors.New("pipeline_class must be low_risk_feature|medium_integration")
	}
	if r.ScenarioTotal < 1 || r.ScenarioPassed < 0 || r.ScenarioPassed > r.ScenarioTotal {
		return errors.New("invalid scenario counts")
	}
	if r.Retries < 0 || r.Interventions < 0 {
		return errors.New("retries/interventions cannot be negative")
	}
	switch r.Decision {
	case "approved", "rejected", "failed":
	default:
		return errors.New("decision must be approved|rejected|failed")
	}
	if r.CriticalIncident && r.Decision != "approved" {
		return errors.New("critical_incident=true requires decision=approved")
	}
	if _, err := time.Parse(time.RFC3339, r.Timestamp); err != nil {
		return fmt.Errorf("timestamp must be RFC3339: %w", err)
	}
	return nil
}

func pct(n, d float64) float64 {
	if d == 0 {
		return 0
	}
	return (n / d) * 100.0
}
