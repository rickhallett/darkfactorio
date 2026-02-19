package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rickhallett/darkfactorio/internal/onboard"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "scaffold":
		return runScaffold(args[1:])
	case "validate-artifacts":
		return runValidate(args[1:])
	case "checklist":
		return runChecklist(args[1:])
	case "help", "-h", "--help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", args[0])
		usage()
		return 2
	}
}

func runScaffold(args []string) int {
	fs := flag.NewFlagSet("scaffold", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	project := fs.String("project", "", "project slug/name (required)")
	candidateProducer := fs.String("candidate-producer", "", "candidate producer id")
	holdoutProducer := fs.String("holdout-producer", "", "holdout producer id")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	manifest, err := onboard.Scaffold(onboard.ScaffoldOptions{
		Root:              ".",
		Project:           *project,
		CandidateProducer: *candidateProducer,
		HoldoutProducer:   *holdoutProducer,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Printf("shadow-pack scaffold created: %s\n", manifest)
	return 0
}

func runValidate(args []string) int {
	fs := flag.NewFlagSet("validate-artifacts", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	manifest := fs.String("manifest", "", "manifest path (required)")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if err := onboard.ValidateArtifacts(onboard.ValidateOptions{
		Root:     ".",
		Manifest: *manifest,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Printf("artifact validation passed: %s\n", *manifest)
	return 0
}

func runChecklist(args []string) int {
	fs := flag.NewFlagSet("checklist", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	project := fs.String("project", "", "project slug/name (required)")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	manifest, err := onboard.Scaffold(onboard.ScaffoldOptions{
		Root:    ".",
		Project: *project,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Printf("checklist scaffold generated via manifest: %s\n", manifest)
	return 0
}

func usage() {
	fmt.Println("dfonboardv01")
	fmt.Println("  scaffold --project <name> [--candidate-producer <id>] [--holdout-producer <id>]")
	fmt.Println("  validate-artifacts --manifest shadowpacks/<project>/manifest.json")
	fmt.Println("  checklist --project <name>")
}
