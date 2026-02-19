package adaptertelemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	RuntimeInputPath string  `json:"runtime_input_path"`
	BillingInputPath string  `json:"billing_input_path"`
	OutRuntimePath   string  `json:"out_runtime_path"`
	OutEconPath      string  `json:"out_econ_path"`
	MinAvailability  float64 `json:"min_availability_percent"`
	MaxErrorRate     float64 `json:"max_error_rate_percent"`
	MaxP95LatencyMs  float64 `json:"max_p95_latency_ms"`
	MaxCostDeltaPct  float64 `json:"max_cost_delta_percent"`
	MaxTokenDeltaPct float64 `json:"max_token_delta_percent"`
}

type RuntimeInput struct {
	AvailabilityPercent float64 `json:"availability_percent"`
	ErrorRatePercent    float64 `json:"error_rate_percent"`
	P95LatencyMs        float64 `json:"p95_latency_ms"`
}

type BillingInput struct {
	ProviderCostUSD float64 `json:"provider_cost_usd"`
	InternalCostUSD float64 `json:"internal_cost_usd"`
	ProviderTokens  int64   `json:"provider_tokens"`
	InternalTokens  int64   `json:"internal_tokens"`
}

type Result struct {
	RuntimePath string `json:"runtime_path"`
	EconPath    string `json:"econ_path"`
}

func Sync(root string, configPath string) (Result, error) {
	if root == "" {
		root = "."
	}
	if configPath == "" {
		return Result{}, fmt.Errorf("config_path is required")
	}
	cfg, err := loadJSON[Config](filepath.Join(root, configPath))
	if err != nil {
		return Result{}, err
	}
	if cfg.RuntimeInputPath == "" || cfg.BillingInputPath == "" || cfg.OutRuntimePath == "" || cfg.OutEconPath == "" {
		return Result{}, fmt.Errorf("config missing required paths")
	}

	rin, err := loadJSON[RuntimeInput](filepath.Join(root, cfg.RuntimeInputPath))
	if err != nil {
		return Result{}, err
	}
	bin, err := loadJSON[BillingInput](filepath.Join(root, cfg.BillingInputPath))
	if err != nil {
		return Result{}, err
	}

	runtimeOut := map[string]any{
		"availability_percent":     rin.AvailabilityPercent,
		"min_availability_percent": cfg.MinAvailability,
		"error_rate_percent":       rin.ErrorRatePercent,
		"max_error_rate_percent":   cfg.MaxErrorRate,
		"p95_latency_ms":           rin.P95LatencyMs,
		"max_p95_latency_ms":       cfg.MaxP95LatencyMs,
	}
	econOut := map[string]any{
		"provider_cost_usd":       bin.ProviderCostUSD,
		"internal_cost_usd":       bin.InternalCostUSD,
		"max_delta_percent":       cfg.MaxCostDeltaPct,
		"provider_tokens":         bin.ProviderTokens,
		"internal_tokens":         bin.InternalTokens,
		"max_token_delta_percent": cfg.MaxTokenDeltaPct,
	}

	runtimeAbs := filepath.Join(root, cfg.OutRuntimePath)
	econAbs := filepath.Join(root, cfg.OutEconPath)
	if err := writeJSON(runtimeAbs, runtimeOut); err != nil {
		return Result{}, err
	}
	if err := writeJSON(econAbs, econOut); err != nil {
		return Result{}, err
	}
	return Result{RuntimePath: cfg.OutRuntimePath, EconPath: cfg.OutEconPath}, nil
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
