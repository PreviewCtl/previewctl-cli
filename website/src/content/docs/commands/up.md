---
title: previewctl up
description: Build and deploy preview services to Docker.
---

## Usage

```bash
previewctl up [flags]
```

## Description

Reads `.previewctl/preview.yml`, resolves service dependencies, and starts the preview stack in Docker.

The `up` command will:

1. Validate the config file
2. Detect the current Git branch (used for naming)
3. Resolve the service dependency graph (DAG)
4. Build services (Dockerfile, Nixpacks, or Railpack)
5. Create the Docker network
6. Deploy all services in dependency order

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--preview-id` | `string` | generated | Preview ID to deploy |
| `--secret` | `string[]` | — | Secret in `KEY=VALUE` format (repeatable) |
| `--env-file` | `string` | `.env` | Path to `.env` file |

## Examples

Basic usage:

```bash
previewctl up
```

With secrets:

```bash
previewctl up --secret STRIPE_KEY=sk_test_abc --secret DB_PASSWORD=hunter2
```

With a custom env file:

```bash
previewctl up --env-file .env.staging
```

With a specific preview ID:

```bash
previewctl up --preview-id my-feature-preview
```

## How it works

PreviewCtl reads the `depends_on` fields in your config to build a directed acyclic graph (DAG) of service dependencies. Services with no dependencies start first, and dependent services wait until their requirements are ready.

Environment variables support templating:

- **`${services.<name>.host}`** — resolves to the container hostname
- **`${services.<name>.port}`** — resolves to the service port
- **`${services.<name>.env.<VAR>}`** — references another service's env var
- **`${secrets.<KEY>}`** — injects a secret from `--secret` or `--env-file`
- **`${Generate(n)}`** — generates a random string of length `n`
