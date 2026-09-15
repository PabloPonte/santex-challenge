# Repository Guidelines

## Project Structure & Architecture

Build a feature-flag administration platform composed of a Go API and a React backoffice. The API exposes an internal management surface for feature CRUD operations and manual cache refresh, plus an external surface that evaluates a feature for a specific user. Keep the public API documentation and OpenAPI contract alongside the backend documentation, and update them whenever the public contract changes.

Organize the Go application as a layered system: `cmd` → router → handler/controller → processor/service → repository/client. Each layer owns one responsibility and depends only on the narrow interfaces it needs. Construct dependencies at the application boundary and inject them downward; do not let handlers create services or services create clients directly.

Use PostgreSQL as the source of truth and Redis as the read mirror. A feature contains a unique, identifying name, description, status (`open`, `closed`, or `whitelisted`), the timestamp of its most recent status change, and a whitelist of users. Every create, update, or delete must keep the Redis representation consistent with PostgreSQL. Provide an internal manual cache-refresh mechanism that rebuilds the mirror from PostgreSQL.

The external evaluation API must read only from Redis. `open` enables the feature for every user, `closed` disables it for every user, and `whitelisted` enables it only for users in the feature's whitelist. If Redis is unavailable, return an appropriate dependency error; do not fall back to PostgreSQL. Define routes, request and response bodies, and status codes in OpenAPI rather than inventing undocumented conventions.

Do not implement authentication or authorization in this scope: the backoffice and internal API are assumed to run in a trusted network, and the external API is assumed to be consumed through an already secured BFF.

## Configuration & Development Commands

Use Gin for HTTP routing. Read external values, including PostgreSQL and Redis connection settings, from environment variables through a dedicated configuration package; do not embed URLs or credentials in source code. Document required variables and local defaults in the backend README.

Start supplied dependencies from the repository root:

```sh
docker compose up
```

Run Go workflows from the backend directory:

```sh
go generate ./...    # regenerate mocks
go test ./...        # run the test suite
go test -cover ./... # report package coverage
go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./... # scan reachable vulnerabilities
```

## Go Style & Naming

Use `gofmt` formatting, idiomatic Go error handling, and small, focused packages. Name packages with short lowercase words; use `PascalCase` for exported identifiers and `camelCase` for unexported ones. Define interfaces near their consumers and keep them minimal. Preserve external JSON contract names exactly; use camelCase for JSON fields and represent identifiers and timestamps exactly as documented in OpenAPI.

## Testing Guidelines

Write black-box unit tests for every package except `cmd`, configuration, and generated mocks. Target at least 80% coverage unless additional coverage has poor cost-benefit. Use Testify assertions and require helpers. Keep setup and fixtures in `*_helper_test.go`; put behavior-focused cases in the primary `*_test.go` file.

Every tested package must have generated mocks in its own `mocks/` subdirectory, named `<package>_mocks.go`. Add package-level `//go:generate` directives using Uber `mockgen`; never hand-write mocks. Test PostgreSQL and Redis repositories against real backing services started with Testcontainers, not mocks.

Cover unique feature names, status transitions and timestamps, whitelist evaluation, cache synchronization after mutations, manual refresh, downstream PostgreSQL and Redis failures, cache-only external evaluation, and concurrent updates to the same feature.

Run `govulncheck` before delivering changes to Go code or dependencies. Resolve vulnerabilities reachable from application code, or document why a remediation is not currently possible.

Maintain the real Python/pytest E2E suite whenever public feature behavior changes. E2E tests must use the API and backing services through their HTTP APIs, generate unique test data, verify persisted resources and their external evaluation, and delete only the resources they created. Add or update a success case for business-rule changes and a status/error assertion for request-validation and cache-availability changes.

## Commit & Pull Request Guidelines

Use short imperative commits, such as `Add feature evaluation endpoint`, and keep each commit focused. Pull requests must describe behavior changes, validation performed, relevant issue links, and API request/response examples. Update the OpenAPI contract whenever the API changes.

Do not create commits, push branches, or open pull requests automatically. Perform any of these Git actions only when the user explicitly requests it.
