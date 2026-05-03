# Repository Guidelines

## Project Structure & Module Organization

This is a Go module for Brightmoon, a CLI and library set for Touhou archive formats. CLI entry points live in `cmd/brightmoon` and `cmd/titles_th`. Reusable public packages are in `pkg/crypto` and `pkg/pbgarc`; command-specific internals for title extraction are under `internal/titles`. Tests are colocated with implementation files as `*_test.go`. Documentation is in `README.md` and `README.titles_th.md`; release helper scripts are in `scripts/`.

## Build, Test, and Development Commands

- `make build`: compile every package with `go build -v ./...`.
- `make build-brightmoon`: build the `brightmoon` CLI to `./brightmoon`.
- `make build-titles`: build the `titles_th` CLI to `./titles_th`.
- `make test`: run all tests with verbose output and the race detector.
- `make test-cover`: generate `coverage.out` and print coverage by function.
- `make lint`: run `golangci-lint` v2.7.1 in Docker.
- `make fmt` and `make vet`: run `go fmt ./...` and `go vet ./...`.
- `make mod-tidy` / `make mod-verify`: maintain and verify module dependencies.

## Coding Style & Naming Conventions

Use standard Go formatting; run `make fmt` before committing. Keep package names short and lower-case, matching current directories such as `pbgarc`, `crypto`, and `fileutil`. Exported identifiers require clear GoDoc when they are part of reusable `pkg/` APIs. Prefer explicit error returns with useful context over logging and continuing. Keep archive format logic separated by type-specific files, following patterns like `kanako.go` with `kanako_test.go`.

## Testing Guidelines

Use Go’s built-in `testing` package. Add or update colocated `*_test.go` files for behavior changes, including error paths and malformed archive data where practical. Run `make test` for normal verification and `make test-cover` when changing shared archive, crypto, parser, or filesystem code. Avoid tests that depend on local files under `tmp/`; use test fixtures or temporary directories instead.

## Commit & Pull Request Guidelines

History uses Conventional Commits, often with scopes, for example `fix(titles_th): ...` and `chore(brightmoon): ...`. Keep commits focused and describe the user-visible effect in English when possible. Pull requests should include a short summary, linked issue if applicable, test results such as `make test`, and notes for CLI behavior changes. Update README files when flags, commands, archive support, or installation steps change.

## Security & Configuration Tips

Do not commit game data, generated extraction output, secrets, or local artifacts. Use `make clean` to remove binaries, coverage files, `dist/`, and `tmp/`. Validate all paths and external archive inputs, and preserve clear errors for unsupported or ambiguous archive formats.
