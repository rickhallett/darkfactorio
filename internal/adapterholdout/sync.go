package adapterholdout

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rickhallett/darkfactorio/internal/shadowpack"
)

type Config struct {
	Project              string `json:"project"`
	HoldoutProducer      string `json:"holdout_producer"`
	HoldoutRepo          string `json:"holdout_repo"`
	HoldoutSHA           string `json:"holdout_sha"`
	HoldoutResultsSource string `json:"holdout_results_source"`
	ExpectedSHA256       string `json:"expected_sha256"`
}

type SyncOptions struct {
	Root           string
	ConfigPath     string
	OutPath        string
	ProvenancePath string
	Validate       bool
}

type SyncResult struct {
	Project          string `json:"project"`
	OutPath          string `json:"out_path"`
	ProvenancePath   string `json:"provenance_path"`
	ResultsSHA256    string `json:"results_sha256"`
	ShadowValidation string `json:"shadow_validation"`
}

func Sync(opts SyncOptions) (SyncResult, error) {
	if opts.Root == "" {
		opts.Root = "."
	}
	if opts.ConfigPath == "" {
		return SyncResult{}, fmt.Errorf("config_path is required")
	}
	cfg, err := loadJSON[Config](filepath.Join(opts.Root, opts.ConfigPath))
	if err != nil {
		return SyncResult{}, err
	}
	if cfg.Project == "" || cfg.HoldoutProducer == "" || cfg.HoldoutRepo == "" || cfg.HoldoutSHA == "" || cfg.HoldoutResultsSource == "" {
		return SyncResult{}, fmt.Errorf("config missing required fields")
	}

	if opts.OutPath == "" {
		opts.OutPath = filepath.ToSlash(filepath.Join("shadowpacks", cfg.Project, "holdout.json"))
	}
	if opts.ProvenancePath == "" {
		opts.ProvenancePath = filepath.ToSlash(filepath.Join("shadowpacks", cfg.Project, "holdout-provenance.json"))
	}

	body, err := fetch(cfg.HoldoutResultsSource)
	if err != nil {
		return SyncResult{}, err
	}
	sum := sha256.Sum256(body)
	sumHex := hex.EncodeToString(sum[:])
	if cfg.ExpectedSHA256 != "" && !strings.EqualFold(cfg.ExpectedSHA256, sumHex) {
		return SyncResult{}, fmt.Errorf("sha mismatch: expected %s got %s", cfg.ExpectedSHA256, sumHex)
	}

	outAbs := filepath.Join(opts.Root, opts.OutPath)
	if err := os.MkdirAll(filepath.Dir(outAbs), 0o755); err != nil {
		return SyncResult{}, err
	}
	if err := os.WriteFile(outAbs, body, 0o644); err != nil {
		return SyncResult{}, err
	}

	prov := map[string]any{
		"holdout_producer": cfg.HoldoutProducer,
		"holdout_repo":     cfg.HoldoutRepo,
		"holdout_sha":      cfg.HoldoutSHA,
		"results_path":     filepath.ToSlash(opts.OutPath),
		"results_sha256":   sumHex,
	}
	provAbs := filepath.Join(opts.Root, opts.ProvenancePath)
	if err := writeJSON(provAbs, prov); err != nil {
		return SyncResult{}, err
	}

	res := SyncResult{
		Project:        cfg.Project,
		OutPath:        filepath.ToSlash(opts.OutPath),
		ProvenancePath: filepath.ToSlash(opts.ProvenancePath),
		ResultsSHA256:  sumHex,
	}

	if opts.Validate {
		manifest := filepath.ToSlash(filepath.Join("shadowpacks", cfg.Project, "manifest.json"))
		rep, err := shadowpack.Evaluate(opts.Root, manifest)
		if err != nil {
			return SyncResult{}, err
		}
		if rep.Passed {
			res.ShadowValidation = "pass"
		} else {
			res.ShadowValidation = "fail"
			return res, fmt.Errorf("shadowpack failed: %v", rep.Failures)
		}
	}

	return res, nil
}

func fetch(src string) ([]byte, error) {
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		resp, err := http.Get(src) //nolint:gosec
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("http status %d", resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	if strings.HasPrefix(src, "file://") {
		src = strings.TrimPrefix(src, "file://")
	}
	return os.ReadFile(src)
}

func loadJSON[T any](path string) (T, error) {
	var out T
	f, err := os.Open(path)
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

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
