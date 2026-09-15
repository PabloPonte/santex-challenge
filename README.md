# Feature Flag Manager

Backend Go para administrar feature flags y evaluarlas desde Redis. PostgreSQL es la fuente de verdad; Redis es un espejo de lectura. No incluye frontend ni autenticación en este alcance.

## Requisitos

- Go 1.27
- Docker Compose
- Python 3 con `pytest` para E2E

## Inicio local

```sh
cp .env.example .env
make deps-up
make migrate-up
make run
```

La API escucha en `http://localhost:8080`. Verificá dependencias con `curl -i http://localhost:8080/healthz`. El contrato está en [docs/openapi.yaml](docs/openapi.yaml).

## Configuración

| Variable | Default local | Uso |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | Dirección HTTP de Gin |
| `DATABASE_URL` | `postgres://feature_flags:feature_flags@localhost:5432/feature_flags?sslmode=disable` | PostgreSQL |
| `REDIS_URL` | `redis://localhost:6379/0` | Redis |

El proceso Go no carga `.env` automáticamente; exportá las variables o ejecutá `source .env` antes de `make run`.

## Comandos

`make help` muestra los objetivos. Los principales son `deps-up`, `migrate-up`, `run`, `test`, `test-cover`, `e2e` y `vulncheck`. `test` y `test-cover` inician PostgreSQL y Redis efímeros mediante Testcontainers, por lo que requieren el daemon Docker. `deps-reset` elimina de forma intencional los volúmenes locales de PostgreSQL y Redis.

## Consistencia

Las escrituras se persisten primero en PostgreSQL y luego se reflejan de forma síncrona en Redis. Si Redis falla, la API responde `503`; la mutación ya persistida puede reconciliarse mediante `POST /api/v1/internal/cache/refresh`. La evaluación externa nunca consulta PostgreSQL: una feature no encontrada devuelve `404` y Redis no disponible devuelve `503`.
