# AGENTS.md

## Scope
- Applies to the entire repository unless a deeper `AGENTS.md` overrides it.
- This repository is a small Go library that wraps `sqlx` and `squirrel`; favor library-safe changes over app-specific assumptions.

## Repository Snapshot
- Module path: `github.com/adlandh/easql`.
- Primary package: `easql` at the repository root.
- Main dependencies used in code and tests: `github.com/jmoiron/sqlx`, `github.com/Masterminds/squirrel`, `github.com/DATA-DOG/go-sqlmock`, and `github.com/stretchr/testify/assert`.
- There is no app binary, code generator, or Makefile in this repo.
- CI runs on Go `1.18` in `/.github/workflows/test.yml`.
- Local lint config exists in `/.golangci.yml` and is also mirrored from a remote config in CI.

## External Agent Rules
- No `.cursorrules` file exists.
- No `.cursor/rules/` directory exists.
- No `.github/copilot-instructions.md` file exists.
- If any of those files are added later, treat them as higher-priority repository instructions and merge them into your working assumptions.

## Working Style
- Keep changes small, focused, and consistent with existing patterns.
- Fix root causes instead of layering on narrow workarounds.
- Avoid unrelated refactors while working on a requested task.
- Preserve the package's simple API shape; avoid expanding public surface area unless the task requires it.
- Prefer updates that keep both `DB` and `Tx` behavior aligned.
- When adding behavior, consider whether the context-aware and non-context APIs should both change.

## Build, Test, And Lint Commands
- Build all packages: `go build ./...`
- Run all tests: `go test ./...`
- Run all tests with race detection: `go test -race ./...`
- Run all tests with coverage, matching the pre-push hook closely: `go test -cover -race ./...`
- Run CI-style coverage locally: `go test -race -coverprofile=coverage.txt -covermode=atomic ./...`
- Run linters: `golangci-lint run`
- Format changed files: `gofmt -w *.go`
- Format imports and files with the configured formatter: `goimports -w *.go`

## Single-Test Commands
- Run one test in the root package: `go test -run '^TestQueryInsert$' .`
- Run one test with verbose output: `go test -v -run '^TestQueryInsert$' .`
- Run a subset of tests by prefix or pattern: `go test -run '^(TestQueryInsert|TestQueryUpdate)$' .`
- Re-run a single test without cached results: `go test -count=1 -run '^TestQueryInsert$' .`
- Run a single test with race detection: `go test -race -run '^TestQueryInsert$' .`
- Because this repo currently uses only the root package, `.` is usually enough; use `./...` only if packages are added later.

## Hooks And CI
- `lefthook` defines a `pre-push` hook in `/.lefthook.yml`.
- The hook runs `golangci-lint run` and `go test -cover -race ./...`.
- Before finishing a non-trivial change, run the same commands locally when possible.
- GitHub Actions also runs lint and tests on pushes and pull requests to `main` and `master`.

## Validation Expectations
- Prefer targeted commands first, then broader repo-wide checks if needed.
- For a code path change, start with the narrowest relevant `go test -run ...` command.
- For interface or shared-query behavior changes, finish with `go test ./...`.
- For concurrency-sensitive or transaction-sensitive changes, also run `go test -race ./...`.
- If you touch formatting or imports only, still run at least the most relevant package tests if the environment allows it.
- Call out any command you could not run, plus the likely reason.

## Formatting And Imports
- Always run `gofmt` on changed Go files before finishing.
- Keep imports in standard Go format: standard library first, blank line, then third-party imports.
- Use `goimports` if you add, remove, or reorder imports.
- Keep import aliases minimal; use them only when they improve clarity or avoid collisions.
- The existing tests alias `github.com/Masterminds/squirrel` as `sq`; follow that established alias where needed.
- Do not leave unused imports, dead constants, or stale helper functions.

## Package And API Design
- This repo favors thin wrappers around existing `sqlx` and `squirrel` APIs.
- Keep package APIs minimal and favor clear names over abbreviations.
- Preserve the current separation between exported wrapper types (`DB`, `Tx`) and unexported implementation types (`queryer`, `queryerContext`).
- Prefer constructor functions like `NewDB` for exported types rather than exposing struct fields.
- Maintain interface assertions such as `var _ Queryer = (*DB)(nil)` when they protect public contracts.
- If you add a new capability, think through all relevant interfaces in `interfaces.go` before editing concrete types.

## Types And Signatures
- Follow the style already used in the surrounding package.
- Use concrete types for constructors and returned wrappers when the caller benefits from package-specific methods.
- Use narrow interfaces for internal abstraction boundaries.
- Prefer existing standard-library types such as `context.Context`, `sql.Result`, and `error` over custom wrappers.
- Keep signatures symmetric between context and non-context variants where practical.
- Do not introduce generics just because the toolchain supports them; this repo targets a small, compatibility-friendly API.

## Naming Conventions
- Use Go naming conventions consistently.
- Exported identifiers use PascalCase and should remain short and obvious: `DB`, `Tx`, `BeginContext`.
- Unexported helpers use lowerCamelCase: `newTx`, `execQuery`, `queryBuilder`.
- Preserve initialisms in their existing uppercase form when already established in the repo, such as `DB` and `Tx`.
- Prefer names that describe behavior over storage details.
- Avoid introducing shorthand that is less clear than the current package vocabulary.

## Error Handling
- Prefer explicit error handling and early returns.
- Wrap returned errors with context using `fmt.Errorf(... %w ...)`.
- Match the existing error style where operations are labeled clearly, for example `error begin`, `error get`, `error exec`, or `error rollback`.
- Keep error messages lowercase and without trailing punctuation.
- Do not panic for normal runtime failures.
- When a helper centralizes repeated error-prone logic, prefer extracting it rather than duplicating the same wrapping code.

## Control Flow And Function Shape
- Prefer small functions with one clear responsibility.
- Return early on error instead of adding extra nesting.
- Keep shared logic in private helpers when it clearly reduces duplication, as seen in `execQuery`.
- Avoid unnecessary indirection if a direct call is simpler.
- Do not introduce clever abstractions that hide SQL execution flow.

## Context Usage
- Use `context.Context` as the first parameter in context-aware methods.
- Pass contexts through directly; do not store them on structs.
- If a behavior exists in both context and non-context forms, keep them semantically aligned.
- Do not create background contexts inside library code unless there is no caller-provided alternative and the change is explicitly intended.

## Testing Style
- Prefer table-driven tests where they improve coverage or reduce duplication.
- The existing suite also uses focused direct tests; follow the surrounding style in the file you edit.
- Use `testing` plus `testify/assert` for readable assertions.
- When using `sqlmock`, set expectations before the action and assert `mock.ExpectationsWereMet()` after it.
- Keep helper functions small and local to the test file when they are only used there.
- Use `t.Parallel()` only when the test is truly safe with shared globals and mock state.
- If you add exported behavior, add or update tests in `easql_test.go` or a nearby `*_test.go` file.

## SQL And Query Builder Conventions
- Build queries with `squirrel` rather than hand-writing SQL strings in production code.
- Preserve the existing split between read operations using `SelectBuilder` and write operations using insert, update, and delete builders.
- Reuse internal helper paths for SQL generation when behavior is identical.
- In tests, match SQL strings and arguments precisely enough to verify behavior without overfitting to irrelevant formatting.

## Comments And Documentation
- Keep comments minimal and useful.
- Preserve package comments and exported documentation where they already exist.
- Add comments only when a block is non-obvious or an exported API needs documentation.
- Update `README.md` when public usage or package behavior changes materially.
- Do not add noisy comments that restate the code.

## Dependencies And Files
- Do not add new dependencies unless they are necessary for the requested task.
- Prefer standard library and existing dependencies first.
- Do not commit generated output unless the repository already tracks it.
- There is no evidence of generated code in this repo today; keep it that way unless the task explicitly introduces it.

## Safety
- Never delete user work or run destructive Git commands unless explicitly requested.
- If the worktree contains unrelated changes, leave them alone.
- If you must touch a file with user changes, read it carefully and preserve their intent.
- Call out assumptions, skipped checks, or environment limitations in the handoff.

## Good Defaults For Agents
- Start by reading `interfaces.go` when the task affects public behavior.
- Read both `queryer.go` and `queryerContext.go` before changing query execution semantics.
- Read `db.go` and `tx.go` together before changing transaction behavior.
- If you modify one side of the API pair (`DB` vs `Tx`, or context vs non-context), check whether the matching side should change too.
- End with formatting plus the narrowest relevant tests, then broaden validation if the change affects shared behavior.
