package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rickhallett/darkfactorio/internal/learning"
)

type listFlag []string

func (l *listFlag) String() string {
	return strings.Join(*l, ",")
}

func (l *listFlag) Set(v string) error {
	for _, part := range strings.Split(v, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			*l = append(*l, part)
		}
	}
	return nil
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}

	switch args[0] {
	case "touch":
		return runTouch(args[1:])
	case "check":
		return runCheck(args[1:])
	case "-h", "--help", "help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n\n", args[0])
		usage()
		return 2
	}
}

func runTouch(args []string) int {
	fs := flag.NewFlagSet("touch", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	root := fs.String("root", ".", "repo root")
	sourceProject := fs.String("source-project", "unknown", "source project name")
	summary := fs.String("summary", "", "short summary for this learning entry")
	when := fs.String("when", "", "timestamp in RFC3339; defaults to now UTC")

	var refs listFlag
	var decisions listFlag
	var evidence listFlag
	var nextActions listFlag

	fs.Var(&refs, "source-ref", "source reference (repeatable or comma-separated)")
	fs.Var(&decisions, "decision", "key decision (repeatable or comma-separated)")
	fs.Var(&evidence, "evidence", "evidence item (repeatable or comma-separated)")
	fs.Var(&nextActions, "next-action", "next action (repeatable or comma-separated)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	var t time.Time
	if strings.TrimSpace(*when) != "" {
		parsed, err := time.Parse(time.RFC3339, *when)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid --when value: %v\n", err)
			return 2
		}
		t = parsed
	}

	path, err := learning.Touch(learning.TouchOptions{
		Root:          *root,
		When:          t,
		SourceProject: strings.TrimSpace(*sourceProject),
		SourceRefs:    refs,
		Summary:       strings.TrimSpace(*summary),
		Decisions:     decisions,
		Evidence:      evidence,
		NextActions:   nextActions,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "touch failed: %v\n", err)
		return 1
	}

	fmt.Printf("learning entry appended: %s\n", path)
	return 0
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	root := fs.String("root", ".", "repo root")
	base := fs.String("base", "HEAD~1", "base ref for comparison")
	head := fs.String("head", "HEAD", "head ref for comparison")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	result, err := learning.Check(learning.CheckOptions{
		Root: *root,
		Base: *base,
		Head: *head,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "check failed: %v\n", err)
		return 1
	}

	if result.Passed {
		fmt.Println("learning gate: PASS")
		if len(result.SubstantiveChanged) > 0 {
			fmt.Printf("substantive changes: %s\n", strings.Join(result.SubstantiveChanged, ", "))
			fmt.Printf("learning updates: %s\n", strings.Join(result.LearningChanged, ", "))
		} else {
			fmt.Println("no substantive changes detected")
		}
		return 0
	}

	fmt.Println("learning gate: FAIL")
	fmt.Printf("substantive changes without learning log updates: %s\n", strings.Join(result.SubstantiveChanged, ", "))
	if len(result.LearningChanged) == 0 {
		fmt.Println("required: update learning/journal/* or learning/decisions/* in same change set")
	} else {
		fmt.Printf("learning files touched (insufficient): %s\n", strings.Join(result.LearningChanged, ", "))
	}
	return 1
}

func usage() {
	fmt.Println("dflearn: project-agnostic learning record gate")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  dflearn touch [flags]")
	fmt.Println("  dflearn check [flags]")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  dflearn touch --source-project tspit --summary \"baseline gate run\" --decision \"keep baseline profile\"")
	fmt.Println("  dflearn check --base origin/main --head HEAD")
}
