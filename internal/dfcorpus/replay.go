package dfcorpus

import (
	"fmt"
	"os"

	"github.com/rickhallett/darkfactorio/internal/level4gate"
)

type ReplayOptions struct {
	Inputs       []string
	WindowFilter map[string]struct{}
	Criteria     level4gate.Criteria
}

type ReplayResult struct {
	Records []level4gate.EvalRecord
	Report  level4gate.GateReport
}

func Replay(opts ReplayOptions) (ReplayResult, error) {
	if len(opts.Inputs) == 0 {
		return ReplayResult{}, fmt.Errorf("at least one input file is required")
	}

	all := make([]level4gate.EvalRecord, 0, 128)
	for _, in := range opts.Inputs {
		recs, err := loadFile(in, opts.WindowFilter)
		if err != nil {
			return ReplayResult{}, fmt.Errorf("%s: %w", in, err)
		}
		all = append(all, recs...)
	}
	if len(all) == 0 {
		return ReplayResult{}, fmt.Errorf("no records matched corpus filters")
	}

	report := level4gate.EvaluateWithCriteria(all, opts.Criteria, "corpus")
	return ReplayResult{
		Records: all,
		Report:  report,
	}, nil
}

func loadFile(path string, windowFilter map[string]struct{}) ([]level4gate.EvalRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	recs, err := level4gate.DecodeNDJSON(f, "")
	if err != nil {
		return nil, err
	}
	if len(windowFilter) == 0 {
		return recs, nil
	}

	out := make([]level4gate.EvalRecord, 0, len(recs))
	for _, r := range recs {
		if _, ok := windowFilter[r.WindowID]; ok {
			out = append(out, r)
		}
	}
	return out, nil
}
