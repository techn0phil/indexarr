# Frontend Unit Tests Plan (React)

## Objective
Build and roll out a pragmatic frontend test suite with enough coverage to protect critical behavior, targeting an initial baseline of 60-70% line coverage.

## Scope (Phase 1)
- Authentication flow (app auth gate + login behavior)
- List behavior (filters, search, view mode persistence, infinite pagination triggers)
- Detail behavior (load, refresh action, role-based menu visibility)
- Admin users CRUD flows (create, edit, delete, set password)
- Core unit foundations (API client contracts and pagination hook)

## Progress Snapshot
- Overall progress: 100%
- Status: In progress
- Last update: 2026-08-26

## To-Do Checklist

### A. Foundation
- [x] Install and configure test stack (Vitest + React Testing Library + jsdom)
- [x] Add test scripts in package.json
- [x] Add vitest config with jsdom + setup file + CSS support
- [x] Add shared test setup (jest-dom, cleanup, browser API mocks)
- [x] Add shared test utilities and reusable fixtures

### B. Core Unit Tests
- [x] Test API client request/response contracts
- [x] Test API client error handling paths
- [x] Test useInfiniteList initial load, loadMore, reset, hasMore, error

### C. Auth + Routing
- [x] Test App auth-loading gate and login gate behavior
- [x] Test redirect behavior for root and unknown routes
- [x] Test LoginPage success/failure/loading/required fields

### D. List & Filtering
- [x] Test ListFilms initial fetch and render
- [x] Test ListFilms filter apply and multi-filter composition
- [x] Test ListFilms search integration
- [x] Test ListFilms view toggle persistence
- [x] Test ListSeries equivalent flows

### E. Detail Pages
- [x] Test MovieDetail load and not-found path
- [x] Test MovieDetail refresh action and role gating
- [x] Test SeriesDetail load, season interactions, refresh role gating

### F. Admin Users
- [x] Test UsersPage list fetch and loading states
- [x] Test create user modal flow + validation
- [x] Test edit user flow
- [x] Test delete user flow
- [x] Test set-password flow

### G. Stabilization
- [x] Run focused suites repeatedly and remove flaky assertions
- [x] Run full suite and capture coverage baseline
- [x] Document baseline and follow-up gap list

## Milestones
- Milestone 1: Tooling + core unit tests green (completed)
- Milestone 2: Auth/list/detail P0 tests green (completed)
- Milestone 3: UsersPage CRUD tests green (completed)
- Milestone 4: Coverage baseline report published (completed)

## Risk Notes
- WebSocket behavior and IntersectionObserver can create flaky tests if over-asserted.
- Route-level tests need stable provider wrappers to avoid brittle setup.
- Coverage should be improved with behavior-driven assertions rather than snapshots.

## Current Execution Notes
- Plan documented.
- Foundation completed (Vitest + RTL + setup utilities).
- First P0 suites completed and passing:
	- src/api/client.test.ts
	- src/hooks/useInfiniteList.test.ts
	- src/pages/LoginPage.test.tsx
- Added and validated passing suites:
	- src/App.test.tsx
	- src/components/LanguageToggle.test.tsx
	- src/components/ScanStatusCard.test.tsx
	- src/components/Sidebar.test.tsx
	- src/components/ThemeToggle.test.tsx
	- src/components/Topbar.test.tsx
	- src/components/UserMenu.test.tsx
	- src/hooks/useAppContext.test.tsx
	- src/pages/ListFilms.test.tsx
	- src/pages/ListSeries.test.tsx
	- src/pages/MovieDetail.test.tsx
	- src/pages/SeriesDetail.test.tsx
	- src/pages/UsersPage.test.tsx
- Current validated count: 67 passing tests across 16 files.
- Current coverage baseline (source-wide): 71.19% lines / 69.25% branches / 62.10% functions (full suite run on 2026-08-26).
- Latest focused validation run: 18 passing tests across 5 files.
- Latest full coverage validation: 67 passing tests across 16 files.
- Follow-up gap list (highest impact):
	- src/components/Layout.tsx (0% lines)
	- src/components/StatCard.tsx (0% lines)
	- src/components/MovieCard.tsx (0% lines)
	- src/components/MovieCardList.tsx (0% lines)
	- src/components/SeriesCard.tsx (0% lines)
	- src/components/SeriesCardList.tsx (0% lines)
	- src/pages/SeriesDetail.tsx (53.97% lines)
	- src/api/client.ts mutation/scan helpers (partially uncovered)
- Coverage target status: 60-70% line objective exceeded (71.19%).
- Next step: lift card/layout/stat components and deepen SeriesDetail/useAppContext branch coverage.
- Hook stability fix applied in src/hooks/useInfiniteList.ts: stable default filters reference to avoid rerender loops.
