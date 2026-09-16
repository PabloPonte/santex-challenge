# Feature Flag Manager

Go backend for managing feature flags and evaluating them from Redis. PostgreSQL is the source of truth; Redis is the read mirror. This scope does not include authentication.

## Requirements

- Go 1.27 (should work in earlier versions, but not tested)
- Docker Compose
- Python 3 for E2E

## Local setup

```sh
cp .env.example .env
make deps-up
make migrate-up
make run
```

The API listens on `http://localhost:8080`. Check dependencies with `curl -i http://localhost:8080/healthz`. The contract is available at [docs/openapi.yaml](docs/openapi.yaml).

## Repository Structure

```
santex-challenge/
├── cmd/api/main.go              # Entry point: wiring/composition root
├── internal/
│   ├── config/                  # Loads configuration from environment variables
│   ├── domain/                  # Pure entities and value objects (Feature, Status, Evaluation)
│   ├── repository/              # FeatureRepository and FeatureCache interfaces (ports) + domain errors
│   │   └── postgres/            # Concrete PostgreSQL implementation (pgx)
│   ├── cache/redis/             # Concrete Redis read-mirror implementation
│   ├── service/                 # Business logic/orchestration (use cases), validation
│   └── http/
│       ├── handler/             # Gin controllers: HTTP parsing <-> domain, error mapping
│       └── router/              # Route definitions (internal/external) and middleware
├── db/migrations/               # SQL migrations (golang-migrate) for the features table
├── docs/openapi.yaml            # OpenAPI contract (source of truth for the public API)
├── e2e/                         # Python/pytest E2E suite against the real API
├── frontend/                    # React + TypeScript + Vite backoffice
│   └── src/
│       ├── api.ts               # Typed HTTP client + error handling (ApiError)
│       ├── App.tsx              # Views: inventory, editor, evaluation panel
│       └── status.ts            # Presentation helpers (labels, date formatting)
├── docker-compose.yml           # PostgreSQL + Redis for local development
├── Makefile                     # Automates dependencies, migrations, tests, E2E, vulncheck
├── notes.md                     # Assumptions and design decisions
└── README.md                    # Setup guide and consistency contract
```

## Configuration

| Variable | Local default | Usage |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | Gin HTTP address |
| `DATABASE_URL` | `postgres://feature_flags:feature_flags@localhost:5432/feature_flags?sslmode=disable` | PostgreSQL |
| `REDIS_URL` | `redis://localhost:6379/0` | Redis |

The Go process does not load `.env` automatically; export the variables or run `source .env` before `make run`.

## Commands

`make help` lists the available targets. The main ones are `deps-up`, `migrate-up`, `run`, `test`, `test-cover`, `e2e`, and `vulncheck`. `test` and `test-cover` start ephemeral PostgreSQL and Redis instances through Testcontainers, so they require the Docker daemon. `deps-reset` intentionally deletes the local PostgreSQL and Redis volumes.

### E2E tests

To run the E2E tests, first create the Python virtual environment and install the dependencies:
```sh
cd e2e
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

## Consistency

Writes are persisted to PostgreSQL first and then synchronously mirrored to Redis. If Redis fails, the API returns `503`; the already-persisted mutation can be reconciled through `POST /api/v1/internal/cache/refresh`. External evaluation never queries PostgreSQL: an unknown feature returns `404`, and unavailable Redis returns `503`.

---
## Frontend

A small React application for managing feature flags is available in [frontend/README.md](frontend/README.md).

To start the frontend locally, run the API first and then:

```sh
cd frontend
npm install
npm run dev
```

---
## Notes and TODO

Notes and assumptions are available in [notes.md](notes.md).

### Pending work and improvements

- Separate the internal and external APIs so they can be deployed and secured independently.
- Add metrics and observability to the external API.
- Substantially improve the E2E tests, including error and concurrency cases.
- Explore the domain further to enrich features, for example by supporting validity periods and/or programmatic flags.
- Add inherited security to the external API so the requesting user can be identified from the authorization token.
- Migrate persistence to a more flexible database such as DynamoDB; a relational database is not required for this use case.
- Improve OpenAPI documentation examples with more realistic data.
