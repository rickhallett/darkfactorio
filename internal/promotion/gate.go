package promotion

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rickhallett/darkfactorio/internal/dfcorpus"
	"github.com/rickhallett/darkfactorio/internal/factoryv05"
	"github.com/rickhallett/darkfactorio/internal/level4gate"
	"github.com/rickhallett/darkfactorio/internal/shadowpack"
	"github.com/rickhallett/darkfactorio/internal/stressv04"
)

type Options struct {
	Root         string
	ShadowPacks  []string
	CorpusInputs []string
	CriteriaPath string
	V05Bundle    string
	RunStress    bool
}

type StageResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details"`
}

type Report struct {
	Passed bool          `json:"passed"`
	Stages []StageResult `json:"stages"`
}

func Run(opts Options) (Report, error) {
	if opts.Root == "" {
		opts.Root = "."
	}
	if len(opts.ShadowPacks) == 0 {
		return Report{}, fmt.Errorf("shadow packs required")
	}
	if len(opts.CorpusInputs) == 0 {
		return Report{}, fmt.Errorf("corpus inputs required")
	}
	if opts.CriteriaPath == "" {
		opts.CriteriaPath = "profiles/level4-gate-v0.1-adversarial.json"
	}
	if opts.V05Bundle == "" {
		opts.V05Bundle = "factory/v0.5/examples/bundle.json"
	}

	rep := Report{Passed: true, Stages: []StageResult{}}
	add := func(name string, passed bool, details string) {
		if !passed {
			rep.Passed = false
		}
		rep.Stages = append(rep.Stages, StageResult{Name: name, Passed: passed, Details: details})
	}

	// Stage 1: shadow packs
	for _, m := range opts.ShadowPacks {
		r, err := shadowpack.Evaluate(opts.Root, m)
		if err != nil {
			add("shadow:"+m, false, err.Error())
			continue
		}
		d := fmt.Sprintf("overlap=%d mismatch=%.2f drift=%.2f", r.OverlapCount, r.OutcomeMismatchRatePercent, r.P95LatencyDriftPercent)
		if !r.Passed {
			d = d + " failures=" + strings.Join(r.Failures, "; ")
		}
		add("shadow:"+m, r.Passed, d)
	}

	// Stage 2: corpus adversarial
	criteria, err := loadCriteria(opts.Root, opts.CriteriaPath)
	if err != nil {
		add("corpus-adversarial", false, err.Error())
	} else {
		c, err := dfcorpus.Replay(dfcorpus.ReplayOptions{
			Inputs:   opts.CorpusInputs,
			Criteria: criteria,
		})
		if err != nil {
			add("corpus-adversarial", false, err.Error())
		} else {
			d := fmt.Sprintf("records=%d pass_rate=%.2f", len(c.Records), c.Report.Metrics.ScenarioPassRatePercent)
			if !c.Report.Passed {
				d += " failures=" + strings.Join(c.Report.Failures, "; ")
			}
			add("corpus-adversarial", c.Report.Passed, d)
		}
	}

	// Stage 3: v0.5 bundle
	v5, err := factoryv05.ValidateBundle(opts.Root, opts.V05Bundle)
	if err != nil {
		add("factory-v05", false, err.Error())
	} else {
		d := "checks=" + strings.Join(v5.Checks, ",")
		if !v5.Passed {
			d += " failures=" + strings.Join(v5.Failures, "; ")
		}
		add("factory-v05", v5.Passed, d)
	}

	// Stage 4: optional stress
	if opts.RunStress {
		s, err := stressv04.Run(opts.Root)
		if err != nil {
			add("stress-v04", false, err.Error())
		} else {
			passCount := 0
			for _, c := range s.Checks {
				if c.Passed {
					passCount++
				}
			}
			add("stress-v04", s.Passed, fmt.Sprintf("checks=%d/%d", passCount, len(s.Checks)))
		}
	}

	return rep, nil
}

func loadCriteria(root, path string) (level4gate.Criteria, error) {
	return loadJSON[level4gate.Criteria](root, path)
}

func loadJSON[T any](root, path string) (T, error) {
	var out T
	f, err := os.Open(filepath.Join(root, path))
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
