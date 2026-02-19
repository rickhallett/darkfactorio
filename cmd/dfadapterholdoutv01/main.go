package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/rickhallett/darkfactorio/internal/adapterholdout"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("dfadapterholdoutv01", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	config := fs.String("config", "", "adapter config path (required)")
	out := fs.String("out", "", "optional holdout output path")
	prov := fs.String("provenance", "", "optional provenance output path")
	validate := fs.Bool("validate", true, "run shadow-pack validation after sync")
	output := fs.String("output", "text", "output format: text|json")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	res, err := adapterholdout.Sync(adapterholdout.SyncOptions{
		Root:           ".",
		ConfigPath:     *config,
		OutPath:        *out,
		ProvenancePath: *prov,
		Validate:       *validate,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	switch *output {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
	default:
		fmt.Printf("project: %s\n", res.Project)
		fmt.Printf("synced holdout: %s\n", res.OutPath)
		fmt.Printf("provenance: %s\n", res.ProvenancePath)
		fmt.Printf("sha256: %s\n", res.ResultsSHA256)
		if res.ShadowValidation != "" {
			fmt.Printf("shadow validation: %s\n", res.ShadowValidation)
		}
	}
	return 0
}
