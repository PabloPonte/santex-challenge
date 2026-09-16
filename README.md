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

### Test e2e
Para correr las pruebas e2e es necesario primero crear el ambiente virtual de python e instalar las dependencias:
```sh
cd e2e
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

## Consistencia

Las escrituras se persisten primero en PostgreSQL y luego se reflejan de forma síncrona en Redis. Si Redis falla, la API responde `503`; la mutación ya persistida puede reconciliarse mediante `POST /api/v1/internal/cache/refresh`. La evaluación externa nunca consulta PostgreSQL: una feature no encontrada devuelve `404` y Redis no disponible devuelve `503`.

---
## Frontend

Una pequeña aplicación React para administrar las feature flags está en [frontend/README.md](frontend/README.md).

Para iniciar el frontend localmente, primero ejecutá la API y luego:

```sh
cd frontend
npm install
npm run dev
```

---
## Notas y TODO

Algunas notas y supuestos están en [notes.md](notes.md).

### Pendientes y Mejoras

- Separar las APIs internal y external para poder deployarlas y asegurarlas de forma independiente.
- Agregar Metricas y observabilidad a la API external
- Mejorar considerablemente los test E2E, incluyendo casos de error y de concurrencia.
- Explorar mas el dominio para enriquecer las features, por ejemplo trabajar con vigencias y/o flags programaticas.
- Incorporar seguridad heredada a la API external, para poder identificar le usuario que consulta desde el token de autorización.
- Migrar la persistencia a una base mas flexible, como DynamoDB, no es necesario usar una base relacional para este caso de uso.

