# PreviewCtl CLI

PreviewCtl is a command-line tool for creating and managing ephemeral preview environments locally using Docker. Define your services, builds, and dependencies in a single YAML config file and bring them up with one command.

## Features

- **Multi-service orchestration** — Define frontend, backend, databases, and workers in one config and deploy them together with automatic dependency ordering.
- **Multiple build strategies** — Build services from a `Dockerfile`, [Nixpacks](https://nixpacks.com), or [Railpack](https://railpack.com).
- **Pre-built images** — Pull and run existing Docker images (e.g. `postgres:16`, `redis:7`) alongside your custom builds.
- **Template variables** — Reference service hosts, ports, and env vars across services with `${services.<name>.<property>}` syntax.
- **Secret management** — Inject secrets via `--secret` flags, `.env` files, or OS environment variables. Reference them in config with `${secrets.<KEY>}`.
- **Auto-generated values** — Use `${Generate(n)}` to produce random strings (e.g. database passwords).
- **Stable port mappings** — Host port assignments are persisted in a local SQLite database and reused across deploys.
- **Git branch awareness** — Preview environments are scoped to the current Git branch, so parallel branches get isolated stacks.
- **DAG-based deployment** — Services are deployed in topological order derived from `depends_on` declarations, with cycle detection.

## Prerequisites

- **Docker** — A running Docker daemon is required.
- **Go 1.26+** — To build from source.
- **Nixpacks** *(optional)* — Required only if you use `build.type: nixpacks`.
- **Railpack** *(optional)* — Required only if you use `build.type: railpack`.

## Installation

```bash
go install github.com/previewctl/previewctl-cli@latest
```

Or build from source:

```bash
git clone https://github.com/previewctl/previewctl-cli.git
cd previewctl-cli
go build -o previewctl .
```

## Quick Start

```bash
# 1. Initialize a new config in your project
previewctl init

# 2. Edit the generated config
#    .previewctl/preview.yml

# 3. Bring up the preview environment
previewctl up

# 4. List active environments
previewctl list

# 5. Tear it down
previewctl delete <name>
```

## Configuration

PreviewCtl reads its configuration from `.previewctl/preview.yml` in your project root. Running `previewctl init` scaffolds a starter config.

### Example

```yaml
version: 1

preview:
  ttl: 24h

services:
  frontend:
    build:
      type: dockerfile
      context: .
      dockerfile: frontend/Dockerfile
    port: 3000
    env:
      API_URL: http://${services.api.host}:${services.api.port}
    depends_on:
      - api

  api:
    build:
      type: railpack
      context: ./api
    port: 8080
    env:
      DATABASE_HOST: ${services.postgres.host}
      DATABASE_PORT: ${services.postgres.port}
      DATABASE_URL: ${services.postgres.env.POSTGRES_URL}
      STRIPE_API_KEY: ${secrets.STRIPE_API_KEY}
    depends_on:
      - postgres
      - redis

  worker:
    build:
      type: nixpacks
      context: ./worker
    env:
      REDIS_HOST: ${services.redis.host}
      REDIS_PORT: ${services.redis.port}
    depends_on:
      - redis

  postgres:
    image: postgres:16
    port: 5432
    volumes:
      - /var/lib/postgresql/data
    env:
      POSTGRES_DB: mydb
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${Generate(16)}
      POSTGRES_URL: postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${services.postgres.host}:${services.postgres.port}/${POSTGRES_DB}?sslmode=disable

  redis:
    image: redis:7
    port: 6379
```

### Config Reference

| Field | Description |
|---|---|
| `version` | Config version (currently `1`). |
| `preview.ttl` | Time-to-live for the preview environment (e.g. `24h`). |
| `services.<name>.build.type` | Build strategy: `dockerfile`, `nixpacks`, or `railpack`. |
| `services.<name>.build.context` | Build context path relative to the project root. |
| `services.<name>.build.dockerfile` | Dockerfile path relative to the build context (default: `Dockerfile`). |
| `services.<name>.image` | Pre-built Docker image to pull and run (mutually exclusive with `build`). |
| `services.<name>.port` | Container port to expose. A random host port is allocated and persisted. |
| `services.<name>.env` | Key-value map of environment variables. Supports template expressions. |
| `services.<name>.volumes` | List of volume mount paths inside the container. |
| `services.<name>.depends_on` | List of service names this service depends on (controls deploy order). |

### Template Expressions

Template expressions use `${...}` syntax inside env values:

| Expression | Resolves To |
|---|---|
| `${services.<name>.host}` | The Docker network alias of the service (its service name). |
| `${services.<name>.port}` | The container port defined for the service. |
| `${services.<name>.env.<KEY>}` | A resolved env var from an upstream service. |
| `${secrets.<KEY>}` | A secret provided via `--secret`, `.env` file, or OS environment. |
| `${Generate(n)}` | A cryptographically random string of length `n`. |
| `${VAR_NAME}` | A reference to another env var within the same service (bare name, no dots). |

## Commands

### `previewctl init`

Initialize PreviewCtl in the current repository. Creates the `.previewctl/` directory and a starter `preview.yml` config file. Also adds `.previewctl/data/` to `.gitignore`.

```bash
previewctl init
```

### `previewctl up`

Build and deploy all services defined in `.previewctl/preview.yml`. Resolves template variables, builds images, creates a Docker network, and starts containers in dependency order.

```bash
previewctl up [flags]
```

| Flag | Description |
|---|---|
| `--preview-id <id>` | Custom preview ID (defaults to an auto-generated value based on directory name and branch). |
| `--secret KEY=VALUE` | Pass a secret inline. Can be repeated. |
| `--env-file <path>` | Path to a `.env` file (defaults to `.env` in the working directory). |

### `previewctl list`

List preview environments tracked by PreviewCtl.

```bash
previewctl list [flags]
```

| Flag | Description |
|---|---|
| `-a`, `--all` | Show environments across all workspaces (default: current workspace only). |

Aliases: `ls`

### `previewctl delete <name>`

Delete a preview environment by name or ID. Stops and removes all Docker containers and the network, then cleans up the local database records.

```bash
previewctl delete <name>
```

Aliases: `rm`

### `previewctl validate`

Validate the `.previewctl/preview.yml` config file without deploying anything. Checks YAML schema, required fields, supported build types, dependency references, and cycle detection.

```bash
previewctl validate
```

## How It Works

1. **Config loading & validation** — The YAML config is parsed with strict field checking. The validator ensures required fields are set, build types are valid, `depends_on` references exist, env var expressions are well-formed, and the dependency graph is acyclic.

2. **Variable resolution** — Services are resolved in topological (dependency) order. Template expressions (`${...}`) are expanded: service hosts resolve to Docker network aliases, ports to their configured values, secrets are looked up from the merged secret sources, and `${Generate(n)}` produces random strings. Intra-service env var references are also resolved in dependency order.

3. **Docker orchestration** — A dedicated Docker bridge network is created for the preview. Each service is deployed in order: custom builds are executed (via `docker build`, `nixpacks build`, or `railpack build`), pre-built images are pulled if not cached, and containers are started with the resolved environment variables, volume mounts, and port bindings.

4. **State persistence** — Preview environments and their port mappings are stored in a local SQLite database at `~/.previewctl/data/previewctl.db`. This allows stable port reuse across re-deploys and environment listing/deletion.

## Project Structure

```
├── cmd/                  # CLI command definitions (Cobra)
│   ├── root.go           # Root command, DB initialization
│   ├── init.go           # previewctl init
│   ├── up.go             # previewctl up
│   ├── list.go           # previewctl list
│   ├── delete.go         # previewctl delete
│   └── validate.go       # previewctl validate
├── core/                 # Shared core library (separate Go module)
│   ├── constants/        # Config directory/file name constants
│   ├── dag/              # Generic directed acyclic graph with topological sort
│   ├── deployment/       # Service deployment order resolution
│   ├── resolver/         # Template variable resolution engine
│   ├── secrets/          # Secret parsing (.env files, CLI flags, OS env)
│   ├── types/            # PreviewConfig and ServiceConfig type definitions
│   ├── validator/        # Config validation rules
│   └── yaml/             # Embedded default config template
├── internal/             # Internal implementation packages
│   ├── build/
│   │   ├── docker/       # Docker client, image building, container lifecycle, networking
│   │   ├── nixpacks/     # Nixpacks build integration
│   │   └── railpack/     # Railpack build integration
│   ├── identity/         # Preview ID generation and validation
│   ├── initializer/      # Repo initialization logic
│   ├── store/            # Store interfaces and types
│   │   └── database/     # SQLite-backed persistence (environments, port mappings, migrations)
│   └── up/               # Orchestration logic for the up command
├── main.go               # Entry point
└── go.mod                # Module definition
```
