package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/rickhallett/darkfactorio/internal/stressv04"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("dfstressv04", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	output := fs.String("output", "text", "output format: text|json")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	rep, err := stressv04.Run(".")
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
		fmt.Printf("stress-v04 passed: %v\n", rep.Passed)
		for _, c := range rep.Checks {
			status := "PASS"
			if !c.Passed {
				status = "FAIL"
			}
			fmt.Printf("- [%s] %s: %s\n", status, c.Name, c.Detail)
		}
	}

	if !rep.Passed {
		return 2
	}
	return 0
}
