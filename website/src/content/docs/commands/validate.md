---
title: previewctl validate
description: Validate the preview config file.
---

## Usage

```bash
previewctl validate
```

## Description

Validates the `.previewctl/preview.yml` configuration file in the current directory.

Checks for:

- Valid YAML syntax
- Required fields (`version`, `services`)
- Valid build types (`dockerfile`, `nixpacks`, `railpack`)
- Valid `depends_on` references (no missing services)
- No circular dependencies in the service graph
- Valid environment variable template syntax

## Examples

```bash
cd my-project
previewctl validate
```

On success:

```
preview config is valid
```

On failure:

```
Error: service "api" depends on "db" which is not defined
```

## When to use

- Before running `previewctl up` to catch config errors early
- In CI pipelines to validate config changes in pull requests
- After editing `preview.yml` to verify syntax
