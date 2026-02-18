package level4gate

import (
	"strings"
	"testing"
)

func TestEvaluatePass(t *testing.T) {
	records := []EvalRecord{
		{RunID: "r1", PipelineID: "p", PipelineClass: "low_risk_feature", ScenarioTotal: 10, ScenarioPassed: 10, FirstPassSuccess: true, Retries: 1, Interventions: 2, Decision: "approved"},
		{RunID: "r2", PipelineID: "p", PipelineClass: "medium_integration", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: true, Retries: 2, Interventions: 2, Decision: "approved"},
		{RunID: "r3", PipelineID: "p", PipelineClass: "low_risk_feature", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: true, Retries: 2, Interventions: 1, Decision: "approved"},
		{RunID: "r4", PipelineID: "p", PipelineClass: "medium_integration", ScenarioTotal: 10, ScenarioPassed: 10, FirstPassSuccess: true, Retries: 2, Interventions: 1, Decision: "approved"},
		{RunID: "r5", PipelineID: "p", PipelineClass: "low_risk_feature", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: true, Retries: 2, Interventions: 1, Decision: "approved"},
		{RunID: "r6", PipelineID: "p", PipelineClass: "medium_integration", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: true, Retries: 2, Interventions: 1, Decision: "approved"},
		{RunID: "r7", PipelineID: "p", PipelineClass: "low_risk_feature", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: false, Retries: 2, Interventions: 1, Decision: "approved"},
		{RunID: "r8", PipelineID: "p", PipelineClass: "medium_integration", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: true, Retries: 2, Interventions: 1, Decision: "approved"},
		{RunID: "r9", PipelineID: "p", PipelineClass: "low_risk_feature", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: false, Retries: 2, Interventions: 0, Decision: "approved"},
		{RunID: "r10", PipelineID: "p", PipelineClass: "medium_integration", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: true, Retries: 2, Interventions: 0, Decision: "approved"},
	}

	report := Evaluate(records, DefaultThresholds(), "w1")
	if !report.Passed {
		t.Fatalf("expected pass, got failures: %v", report.Failures)
	}
}

func TestEvaluateFail(t *testing.T) {
	records := []EvalRecord{
		{RunID: "r1", PipelineID: "p", PipelineClass: "low_risk_feature", ScenarioTotal: 10, ScenarioPassed: 7, FirstPassSuccess: false, Retries: 3, Interventions: 1, Decision: "approved", CriticalIncident: true},
		{RunID: "r2", PipelineID: "p", PipelineClass: "medium_integration", ScenarioTotal: 10, ScenarioPassed: 8, FirstPassSuccess: false, Retries: 3, Interventions: 2, Decision: "approved", DecisionReversed: true},
		{RunID: "r3", PipelineID: "p", PipelineClass: "low_risk_feature", ScenarioTotal: 10, ScenarioPassed: 7, FirstPassSuccess: false, Retries: 3, Interventions: 3, Decision: "rejected"},
		{RunID: "r4", PipelineID: "p", PipelineClass: "medium_integration", ScenarioTotal: 10, ScenarioPassed: 8, FirstPassSuccess: false, Retries: 3, Interventions: 4, Decision: "rejected"},
		{RunID: "r5", PipelineID: "p", PipelineClass: "low_risk_feature", ScenarioTotal: 10, ScenarioPassed: 7, FirstPassSuccess: false, Retries: 3, Interventions: 5, Decision: "rejected"},
		{RunID: "r6", PipelineID: "p", PipelineClass: "medium_integration", ScenarioTotal: 10, ScenarioPassed: 8, FirstPassSuccess: false, Retries: 3, Interventions: 6, Decision: "failed"},
		{RunID: "r7", PipelineID: "p", PipelineClass: "low_risk_feature", ScenarioTotal: 10, ScenarioPassed: 7, FirstPassSuccess: false, Retries: 3, Interventions: 7, Decision: "failed"},
		{RunID: "r8", PipelineID: "p", PipelineClass: "medium_integration", ScenarioTotal: 10, ScenarioPassed: 7, FirstPassSuccess: false, Retries: 3, Interventions: 8, Decision: "failed"},
		{RunID: "r9", PipelineID: "p", PipelineClass: "low_risk_feature", ScenarioTotal: 10, ScenarioPassed: 8, FirstPassSuccess: false, Retries: 3, Interventions: 9, Decision: "failed"},
		{RunID: "r10", PipelineID: "p", PipelineClass: "medium_integration", ScenarioTotal: 10, ScenarioPassed: 7, FirstPassSuccess: false, Retries: 3, Interventions: 10, Decision: "failed"},
	}

	report := Evaluate(records, DefaultThresholds(), "w2")
	if report.Passed {
		t.Fatalf("expected fail")
	}
	if len(report.Failures) < 3 {
		t.Fatalf("expected multiple failure reasons, got: %v", report.Failures)
	}
}

func TestDecodeNDJSONRejectsUnknownField(t *testing.T) {
	in := `{"window_id":"w","run_id":"r1","pipeline_id":"p1","pipeline_class":"low_risk_feature","scenario_total":10,"scenario_passed":10,"first_pass_success":true,"retries":0,"interventions":0,"decision":"approved","decision_reversed":false,"critical_incident":false,"timestamp":"2026-02-18T20:00:00Z","extra":"nope"}`
	_, err := DecodeNDJSON(strings.NewReader(in), "")
	if err == nil {
		t.Fatalf("expected error for unknown field")
	}
	if !strings.Contains(err.Error(), `unknown field "extra"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDecodeNDJSONRejectsMissingRequiredField(t *testing.T) {
	in := `{"window_id":"w","run_id":"r1","pipeline_id":"p1","pipeline_class":"low_risk_feature","scenario_total":10,"scenario_passed":10,"first_pass_success":true,"retries":0,"interventions":0,"decision":"approved","critical_incident":false,"timestamp":"2026-02-18T20:00:00Z"}`
	_, err := DecodeNDJSON(strings.NewReader(in), "")
	if err == nil {
		t.Fatalf("expected error for missing required field")
	}
	if !strings.Contains(err.Error(), `missing required field "decision_reversed"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDecodeNDJSONRejectsInvalidSchemaValues(t *testing.T) {
	in := strings.Join([]string{
		`{"window_id":"w","run_id":"r1","pipeline_id":"p1","pipeline_class":"unknown","scenario_total":10,"scenario_passed":10,"first_pass_success":true,"retries":0,"interventions":0,"decision":"approved","decision_reversed":false,"critical_incident":false,"timestamp":"2026-02-18T20:00:00Z"}`,
		`{"window_id":"w","run_id":"r2","pipeline_id":"p2","pipeline_class":"low_risk_feature","scenario_total":0,"scenario_passed":0,"first_pass_success":true,"retries":0,"interventions":0,"decision":"approved","decision_reversed":false,"critical_incident":false,"timestamp":"2026-02-18T20:00:00Z"}`,
		`{"window_id":"w","run_id":"r3","pipeline_id":"p3","pipeline_class":"low_risk_feature","scenario_total":10,"scenario_passed":10,"first_pass_success":true,"retries":0,"interventions":0,"decision":"rejected","decision_reversed":false,"critical_incident":true,"timestamp":"2026-02-18T20:00:00Z"}`,
		`{"window_id":"w","run_id":"r4","pipeline_id":"p4","pipeline_class":"low_risk_feature","scenario_total":10,"scenario_passed":10,"first_pass_success":true,"retries":0,"interventions":0,"decision":"approved","decision_reversed":false,"critical_incident":false,"timestamp":"not-a-time"}`,
	}, "\n")

	_, err := DecodeNDJSON(strings.NewReader(in), "")
	if err == nil {
		t.Fatalf("expected decode failure")
	}
	if !strings.Contains(err.Error(), "line 1: pipeline_class must be low_risk_feature|medium_integration") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDecodeNDJSONAcceptsValidRecord(t *testing.T) {
	in := `{"window_id":"w","run_id":"r1","pipeline_id":"p1","pipeline_class":"low_risk_feature","scenario_total":10,"scenario_passed":10,"first_pass_success":true,"retries":0,"interventions":0,"decision":"approved","decision_reversed":false,"critical_incident":false,"timestamp":"2026-02-18T20:00:00Z"}`
	recs, err := DecodeNDJSON(strings.NewReader(in), "w")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected one record, got %d", len(recs))
	}
}
