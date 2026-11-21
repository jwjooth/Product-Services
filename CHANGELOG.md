# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - 1.2.0

### Changed
- Changed `price` representation throughout the API from a JSON number (`float64`) to a JSON string transporting integer cents (`"19.99"` → 1999 cents), with exact validation (rejects JSON numbers, 3-decimal values, negatives, and values over $100M). Stored column type changed to `BIGINT` cents.
- Bumped Swagger documentation to version `1.2.0`.

## [1.1.0] - 2026-10-31

### Added
- Extracted HTTP routing and middleware assembly into dedicated `app` package (`app/router.go`)
- Introduced `run()` error-returning wrapper in `main.go` with signal-context server shutdown
- Created repository `AGENTS.md` for AI agent operational guidance
- Added unit tests for `app/router` and `config` packages

### Fixed
- Fixed `ProductRepository.FindAll` count query to apply filter criteria (category, min/max price, search query) so paginated count matches filtered results

### Changed
- Refactored `ProductRepository.FindAll` query building to use shared `applyProductFilters` helper
- Updated Swagger documentation to version `1.1.0`
