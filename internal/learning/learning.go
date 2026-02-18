package learning

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type TouchOptions struct {
	Root          string
	When          time.Time
	SourceProject string
	SourceRefs    []string
	Summary       string
	Decisions     []string
	Evidence      []string
	NextActions   []string
}

type CheckOptions struct {
	Root string
	Base string
	Head string
}

type CheckResult struct {
	Passed             bool
	SubstantiveChanged []string
	LearningChanged    []string
}

func Touch(opts TouchOptions) (string, error) {
	if opts.Root == "" {
		opts.Root = "."
	}
	if opts.When.IsZero() {
		opts.When = time.Now().UTC()
	}
	if opts.SourceProject == "" {
		opts.SourceProject = "unknown"
	}
	if strings.TrimSpace(opts.Summary) == "" {
		opts.Summary = "automatic learning capture"
	}
	if len(opts.NextActions) == 0 {
		opts.NextActions = []string{"Triage outcomes and schedule next capture"}
	}

	day := opts.When.Format("2006-01-02")
	year := opts.When.Format("2006")
	journalDir := filepath.Join(opts.Root, "learning", "journal", year)
	if err := os.MkdirAll(journalDir, 0o755); err != nil {
		return "", err
	}
	journalPath := filepath.Join(journalDir, day+".md")

	if _, err := os.Stat(journalPath); errors.Is(err, os.ErrNotExist) {
		header := fmt.Sprintf("# Learning Log %s\n\n", day)
		header += "This is an append-only operational learning record for darkfactorio, agnostic of source projects.\n\n"
		if err := os.WriteFile(journalPath, []byte(header), 0o644); err != nil {
			return "", err
		}
	}

	entry := buildEntry(opts)
	f, err := os.OpenFile(journalPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.WriteString(entry); err != nil {
		return "", err
	}
	return journalPath, nil
}

func Check(opts CheckOptions) (CheckResult, error) {
	if opts.Root == "" {
		opts.Root = "."
	}
	if opts.Head == "" {
		opts.Head = "HEAD"
	}
	if opts.Base == "" {
		opts.Base = "HEAD~1"
	}

	changed, err := gitDiffNames(opts.Root, opts.Base, opts.Head)
	if err != nil {
		return CheckResult{}, err
	}

	var substantive []string
	var learningChanged []string
	for _, p := range changed {
		norm := filepath.ToSlash(strings.TrimSpace(p))
		if norm == "" {
			continue
		}
		if strings.HasPrefix(norm, "learning/") {
			learningChanged = append(learningChanged, norm)
			continue
		}
		// Ignore repo-only metadata churn that should not require log entries.
		if norm == ".gitignore" {
			continue
		}
		substantive = append(substantive, norm)
	}

	sort.Strings(substantive)
	sort.Strings(learningChanged)

	if len(substantive) == 0 {
		return CheckResult{Passed: true, SubstantiveChanged: substantive, LearningChanged: learningChanged}, nil
	}

	hasJournal := false
	for _, p := range learningChanged {
		if strings.HasPrefix(p, "learning/journal/") || strings.HasPrefix(p, "learning/decisions/") {
			hasJournal = true
			break
		}
	}

	return CheckResult{
		Passed:             hasJournal,
		SubstantiveChanged: substantive,
		LearningChanged:    learningChanged,
	}, nil
}

func buildEntry(opts TouchOptions) string {
	var b strings.Builder
	ts := opts.When.UTC().Format(time.RFC3339)
	b.WriteString(fmt.Sprintf("## %s\n", ts))
	b.WriteString(fmt.Sprintf("- Source Project: `%s`\n", opts.SourceProject))
	if len(opts.SourceRefs) > 0 {
		b.WriteString(fmt.Sprintf("- Source Refs: `%s`\n", strings.Join(opts.SourceRefs, "`, `")))
	}
	b.WriteString(fmt.Sprintf("- Summary: %s\n", opts.Summary))
	b.WriteString("- Key Decisions:\n")
	for _, d := range ensureList(opts.Decisions, "No explicit decision recorded") {
		b.WriteString(fmt.Sprintf("  - %s\n", d))
	}
	b.WriteString("- Evidence:\n")
	for _, e := range ensureList(opts.Evidence, "No evidence links attached") {
		b.WriteString(fmt.Sprintf("  - %s\n", e))
	}
	b.WriteString("- Next Actions:\n")
	for _, n := range ensureList(opts.NextActions, "No next action recorded") {
		b.WriteString(fmt.Sprintf("  - %s\n", n))
	}
	b.WriteString("\n")
	return b.String()
}

func ensureList(in []string, fallback string) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return []string{fallback}
	}
	return out
}

func gitDiffNames(root, base, head string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", fmt.Sprintf("%s...%s", base, head))
	cmd.Dir = root
	var out bytes.Buffer
	var er bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &er
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git diff failed: %v: %s", err, strings.TrimSpace(er.String()))
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
		return []string{}, nil
	}
	return lines, nil
}
