package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rickhallett/darkfactorio/internal/promotion"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("dfpromotionv01", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	manifests := fs.String("manifests", "shadowpacks/antirez-linenoise/manifest.json,shadowpacks/davegamble-cjson/manifest.json,shadowpacks/benhoyt-inih/manifest.json", "comma-separated shadow pack manifests")
	corpusInputs := fs.String("corpus-inputs", "runs/w-2026-02-l4-02.ndjson,runs/w-2026-02-l4-03.ndjson", "comma-separated corpus input files")
	criteria := fs.String("criteria", "profiles/level4-gate-v0.1-adversarial.json", "corpus criteria profile")
	bundle := fs.String("bundle", "factory/v0.5/examples/bundle.json", "v0.5 bundle path")
	runStress := fs.Bool("run-stress", false, "run stress-v04 as final stage")
	output := fs.String("output", "text", "output format: text|json")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	rep, err := promotion.Run(promotion.Options{
		Root:         ".",
		ShadowPacks:  splitCSV(*manifests),
		CorpusInputs: splitCSV(*corpusInputs),
		CriteriaPath: *criteria,
		V05Bundle:    *bundle,
		RunStress:    *runStress,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	switch *output {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(rep)
	default:
		fmt.Printf("promotion gate passed: %v\n", rep.Passed)
		for _, s := range rep.Stages {
			state := "PASS"
			if !s.Passed {
				state = "FAIL"
			}
			fmt.Printf("- [%s] %s :: %s\n", state, s.Name, s.Details)
		}
	}

	if !rep.Passed {
		return 2
	}
	return 0
}

func splitCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
