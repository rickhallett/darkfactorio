package level4gate

import "testing"

func TestEvaluatePass(t *testing.T) {
	records := []EvalRecord{
		{RunID: "r1", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 10, FirstPassSuccess: true, Retries: 1, Interventions: 2, Decision: "approved"},
		{RunID: "r2", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: true, Retries: 2, Interventions: 2, Decision: "approved"},
		{RunID: "r3", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: true, Retries: 2, Interventions: 1, Decision: "approved"},
		{RunID: "r4", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 10, FirstPassSuccess: true, Retries: 2, Interventions: 1, Decision: "approved"},
		{RunID: "r5", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: true, Retries: 2, Interventions: 1, Decision: "approved"},
		{RunID: "r6", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: true, Retries: 2, Interventions: 1, Decision: "approved"},
		{RunID: "r7", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: false, Retries: 2, Interventions: 1, Decision: "approved"},
		{RunID: "r8", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: true, Retries: 2, Interventions: 1, Decision: "approved"},
		{RunID: "r9", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: false, Retries: 2, Interventions: 0, Decision: "approved"},
		{RunID: "r10", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 9, FirstPassSuccess: true, Retries: 2, Interventions: 0, Decision: "approved"},
	}

	report := Evaluate(records, DefaultThresholds(), "w1")
	if !report.Passed {
		t.Fatalf("expected pass, got failures: %v", report.Failures)
	}
}

func TestEvaluateFail(t *testing.T) {
	records := []EvalRecord{
		{RunID: "r1", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 7, FirstPassSuccess: false, Retries: 3, Interventions: 1, Decision: "approved", CriticalIncident: true},
		{RunID: "r2", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 8, FirstPassSuccess: false, Retries: 3, Interventions: 2, Decision: "approved", DecisionReversed: true},
		{RunID: "r3", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 7, FirstPassSuccess: false, Retries: 3, Interventions: 3, Decision: "rejected"},
		{RunID: "r4", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 8, FirstPassSuccess: false, Retries: 3, Interventions: 4, Decision: "rejected"},
		{RunID: "r5", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 7, FirstPassSuccess: false, Retries: 3, Interventions: 5, Decision: "rejected"},
		{RunID: "r6", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 8, FirstPassSuccess: false, Retries: 3, Interventions: 6, Decision: "failed"},
		{RunID: "r7", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 7, FirstPassSuccess: false, Retries: 3, Interventions: 7, Decision: "failed"},
		{RunID: "r8", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 7, FirstPassSuccess: false, Retries: 3, Interventions: 8, Decision: "failed"},
		{RunID: "r9", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 8, FirstPassSuccess: false, Retries: 3, Interventions: 9, Decision: "failed"},
		{RunID: "r10", PipelineID: "p", ScenarioTotal: 10, ScenarioPassed: 7, FirstPassSuccess: false, Retries: 3, Interventions: 10, Decision: "failed"},
	}

	report := Evaluate(records, DefaultThresholds(), "w2")
	if report.Passed {
		t.Fatalf("expected fail")
	}
	if len(report.Failures) < 3 {
		t.Fatalf("expected multiple failure reasons, got: %v", report.Failures)
	}
}

