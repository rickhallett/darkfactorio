package learning

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTouchCreatesJournalEntry(t *testing.T) {
	root := t.TempDir()

	path, err := Touch(TouchOptions{
		Root:          root,
		When:          time.Date(2026, 2, 18, 21, 0, 0, 0, time.UTC),
		SourceProject: "tspit",
		SourceRefs:    []string{"run-001"},
		Summary:       "baseline replay",
		Decisions:     []string{"keep profile"},
		Evidence:      []string{"runs/run-001.ndjson"},
		NextActions:   []string{"run adversarial"},
	})
	if err != nil {
		t.Fatalf("Touch failed: %v", err)
	}

	if !strings.HasSuffix(path, "learning/journal/2026/2026-02-18.md") {
		t.Fatalf("unexpected journal path: %s", path)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read journal: %v", err)
	}
	content := string(raw)
	if !strings.Contains(content, "Source Project: `tspit`") {
		t.Fatalf("missing source project in entry: %s", content)
	}
	if !strings.Contains(content, "Summary: baseline replay") {
		t.Fatalf("missing summary in entry: %s", content)
	}
}

func TestCheckRequiresLearningUpdate(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "branch", "-M", "main")

	mustWrite(t, filepath.Join(root, "README.md"), "start\n")
	runGit(t, root, "add", "README.md")
	runGitCommit(t, root, "init")
	base := strings.TrimSpace(runGit(t, root, "rev-parse", "HEAD"))

	mustWrite(t, filepath.Join(root, "core.txt"), "change\n")
	runGit(t, root, "add", "core.txt")
	runGitCommit(t, root, "feat")

	result, err := Check(CheckOptions{Root: root, Base: base, Head: "HEAD"})
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if result.Passed {
		t.Fatalf("expected gate failure without learning update")
	}

	runGit(t, root, "checkout", base)
	runGit(t, root, "checkout", "-b", "with-learning")
	mustWrite(t, filepath.Join(root, "core.txt"), "change\n")
	mustWrite(t, filepath.Join(root, "learning/journal/2026/2026-02-18.md"), "log\n")
	runGit(t, root, "add", ".")
	runGitCommit(t, root, "feat+learning")

	result, err = Check(CheckOptions{Root: root, Base: base, Head: "HEAD"})
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if !result.Passed {
		t.Fatalf("expected gate pass with learning update")
	}
}

func mustWrite(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v: %s", args, err, string(out))
	}
	return string(out)
}

func runGitCommit(t *testing.T, dir string, message string) {
	t.Helper()
	cmd := exec.Command("git", "-c", "user.name=Test User", "-c", "user.email=test@example.com", "commit", "-m", message)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git commit failed: %v: %s", err, string(out))
	}
}
