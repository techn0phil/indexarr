# Backend Unit Tests Plan

Goal: reach strong backend unit coverage with deterministic tests and no real external service calls.

**Status**: In Progress - scanner stale deletion + radarr/sonarr importer flow tests added, 0 failing

## Foundation
- [x] Create shared test helper for in-memory SQLite setup
- [x] Create shared test helper for auth token generation and request helpers
- [x] Add testing README notes and coverage commands

## Repository
- [x] Add tests for query filter builder and pagination behavior
- [x] Add tests for movie query/read paths
- [x] Add tests for insert/update movie and episode paths (13 tests, 2 skipped due to series schema)
- [x] Add tests for series query/read paths (unblocked with migration-aware test DB helper)
- [x] Add tests for stats queries
- [x] Add tests for mutations: season and series count update paths
- [x] Add tests for purge and delete helpers
- [x] Add tests for exists helpers
- [x] Add tests for user repository CRUD and edge cases

## Services
- [x] Add tests for parser helpers and filename parsing matrix
- [x] Add tests for auth service credential and token flows
- [x] Add tests for broadcaster register/unregister and message behavior
- [x] Add tests for extractor parsing helpers (no real process)
- [x] Add tests for scheduler orchestration with fake importers
- [x] Add tests for scanner path and stale media helper logic
- [x] Add tests for filesystem wrapper delegation
- [x] Add tests for radarr importer mapping and flow with doubles
- [x] Add tests for sonarr importer mapping and flow with doubles

## External Clients (with test doubles only)
- [x] Add tests for TMDB client request/response handling
- [x] Add tests for TVDB client token and request handling
- [x] Add tests for Radarr client request/response handling
- [x] Add tests for Sonarr client request/response handling

## API
- [x] Add tests for auth middleware enabled and disabled behavior
- [x] Add tests for auth handlers (config, login, logout, me)
- [x] Add tests for password and user admin handlers
- [x] Add tests for media handlers (list and detail)
- [x] Add tests for config, stats, and scan status handlers
- [x] Add tests for scan trigger/stop/refresh handlers (admin gating + success/error paths)
- [x] Add route wiring smoke tests
- [x] Add websocket handler unit tests for initial status paths

## Verification
- [x] Run repository tests
- [x] Run services tests
- [x] Run API tests
- [x] Run full test suite with race detector
- [x] Generate and review coverage report
- [x] Confirm no test performs real external network/service/process calls

## Handoff Snapshot (2026-08-25, update #3)

Current branch delta since update #2:
- New services tests added:
	- `internal/services/scanner_test.go`: stale movie/episode deletion paths and library-path gating behavior.
	- `internal/services/radarr_importer_test.go`: map helpers, cached pending count -> import flow, stale TMDB deletion.
	- `internal/services/sonarr_importer_test.go`: map helpers, cached pending count -> import flow, season/episode sync path, stale Sonarr ID deletion.
- Backend testing docs updated in `backend/go/README.md` with package/race/coverage commands and deterministic test guidance.

Verification run completed:
	- `go test ./internal/repository ./internal/services ./internal/api`
	- `go test ./... -race`
	- `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`
	- total coverage: 39.7%

External-call guard audit:
- Pattern audit run over `internal/**/*_test.go` for direct network/process call markers (`exec.Command`, `http.Get`, `http.Post`, `net.Dial`, `http.DefaultClient`).
- Matches were only URL string literals and expected local-baseURL assertions; no direct real external call usage detected in tests.

## Handoff Snapshot (2026-08-25, update #2)

Current branch state:
- Migration-aware repository test helper added and used for series + user repository coverage.
- New repository test files: series/stats queries, mutation flows, and user CRUD/error paths.
- Auth API tests expanded for change-password and admin user-management handlers.
- Auth API tests now also cover scan trigger/movies/series, stop, and movie/series refresh handlers (including admin gating).
- New services tests added for extractor helper parsing, scheduler coordination/contexts, scanner path helper matching, filesystem wrapper count/status delegation, and external clients (TMDB/TVDB/Radarr/Sonarr) using httptest doubles.
- Verification run completed:
	- go test ./... -race
	- go test ./... -coverprofile=coverage.out
	- total coverage: 31.7%

Main remaining work:
- Optional: expand service coverage for startup backfill helpers (`series_totals_startup_backfill.go`).
- Optional: raise repository/services critical-path coverage beyond current 39.7% total.

Recommended next tasks (in order):
1. Add focused tests for `series_totals_startup_backfill.go` branches.
2. Add additional importer error-path assertions (status update failures, partial episode errors aggregation).
3. Add package-level coverage thresholds to CI (non-blocking warning first).

Resume commands:
- cd backend/go
- go test ./internal/repository ./internal/services ./internal/api
- go test ./... -race
- go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out
