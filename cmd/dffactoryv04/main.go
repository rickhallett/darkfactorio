package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/rickhallett/darkfactorio/internal/factoryv04"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("dffactoryv04", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	bundle := fs.String("bundle", "factory/v0.4/examples/bundle.json", "bundle JSON path")
	output := fs.String("output", "text", "output format: text|json")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	rep, err := factoryv04.ValidateBundle(".", *bundle)
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
		fmt.Printf("factory bundle: %s\n", rep.BundleRef)
		fmt.Printf("passed: %v\n", rep.Passed)
		fmt.Printf("checks_passed: %v\n", rep.Checks)
		if len(rep.Failures) > 0 {
			fmt.Println("failures:")
			for _, f := range rep.Failures {
				fmt.Printf("- %s\n", f)
			}
		}
	}

	if !rep.Passed {
		return 2
	}
	return 0
}
