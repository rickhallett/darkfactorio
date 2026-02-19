# Thin Layer External Ingestion Record (2026-02-19)

## Objective

Add thin-layer scenario emitters directly inside selected OSS repos, run them, ingest outputs into `darkfactorio` shadow packs, and verify independent candidate/holdout comparison gates.

## External Repos Updated (local clones)

1. `/tmp/df-ext/linenoise`
   - commit: `e3066e5`
   - change: `darkfactorio_generate.sh` + `darkfactorio/*` artifacts and README
2. `/tmp/df-ext/cjson`
   - commit: `08fe7af`
   - change: `darkfactorio_generate.sh` + `darkfactorio/*` artifacts and README
3. `/tmp/df-ext/inih`
   - commit: `cf56774`
   - change: `darkfactorio_generate.sh` + `darkfactorio/*` artifacts and README

Commit message in each repo:

- `chore(darkfactorio): add thin-layer scenario emitter for external candidate/holdout ingestion`

## Ingestion Targets in darkfactorio

- `shadowpacks/antirez-linenoise/candidate.json`
- `shadowpacks/antirez-linenoise/holdout.json`
- `shadowpacks/davegamble-cjson/candidate.json`
- `shadowpacks/davegamble-cjson/holdout.json`
- `shadowpacks/benhoyt-inih/candidate.json`
- `shadowpacks/benhoyt-inih/holdout.json`

## Shadow-Pack Gate Results

1. `antirez-linenoise-shadowpack-v0.1`
   - pass: true
   - overlap: 10
   - mismatch_rate: 0.00%
   - p95_latency_drift: 8.87%
2. `davegamble-cjson-shadowpack-v0.1`
   - pass: true
   - overlap: 10
   - mismatch_rate: 0.00%
   - p95_latency_drift: 9.06%
3. `benhoyt-inih-shadowpack-v0.1`
   - pass: true
   - overlap: 10
   - mismatch_rate: 0.00%
   - p95_latency_drift: 1.09%

## Decision

External thin-layer ingestion path is validated for first three OSS candidates. Continue by replacing scripted checks with richer behavioral scenarios per project while preserving producer separation.
