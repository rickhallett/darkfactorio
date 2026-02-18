package level4gate

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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

type Metrics struct {
	RunCount                     int     `json:"run_count"`
	ScenarioPassRatePercent      float64 `json:"scenario_pass_rate_percent"`
	FirstPassRatePercent         float64 `json:"first_pass_rate_percent"`
	MeanRetries                  float64 `json:"mean_retries"`
	InterventionAvgFirstHalf     float64 `json:"intervention_avg_first_half"`
	InterventionAvgSecondHalf    float64 `json:"intervention_avg_second_half"`
	InterventionStableOrDecreasing bool  `json:"intervention_stable_or_decreasing"`
	DecisionReversalPercent      float64 `json:"decision_reversal_percent"`
	ApprovedRunCriticalIncidents int     `json:"approved_run_critical_incidents"`
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
		if len(raw) == 0 {
			continue
		}
		var rec EvalRecord
		if err := json.Unmarshal(raw, &rec); err != nil {
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

func Evaluate(records []EvalRecord, thresholds Thresholds, windowID string) GateReport {
	m := computeMetrics(records)
	failures := evaluateFailures(m, thresholds)
	return GateReport{
		WindowID:   windowID,
		Thresholds: thresholds,
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

	firstHalfInterventions := 0
	secondHalfInterventions := 0
	split := len(records) / 2
	if split == 0 {
		split = 1
	}

	for i, r := range records {
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

func evaluateFailures(m Metrics, t Thresholds) []string {
	var failures []string
	if m.RunCount < 10 {
		failures = append(failures, "run_count below minimum window (need >= 10)")
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
	if r.RunID == "" {
		return errors.New("run_id is required")
	}
	if r.PipelineID == "" {
		return errors.New("pipeline_id is required")
	}
	if r.ScenarioTotal < 0 || r.ScenarioPassed < 0 || r.ScenarioPassed > r.ScenarioTotal {
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
	return nil
}

func pct(n, d float64) float64 {
	if d == 0 {
		return 0
	}
	return (n / d) * 100.0
}

