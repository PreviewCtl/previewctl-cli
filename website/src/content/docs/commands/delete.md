---
title: previewctl delete
description: Delete a preview environment.
---

## Usage

```bash
previewctl delete <name or id>
```

**Aliases:** `rm`

## Description

Deletes a preview environment by name or ID.

This command:

1. Looks up the environment by name, falling back to ID
2. Stops and removes all Docker containers on the preview network
3. Removes the Docker network
4. Cleans up port mappings and environment records from the local database

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name or id>` | Yes | Name or ID of the preview environment to delete |

## Examples

Delete by name:

```bash
previewctl delete my-project-main
```

Delete by ID:

```bash
previewctl delete abc123
```

Using the alias:

```bash
previewctl rm my-project-main
```

## Output

```
stopped frontend
stopped api
stopped postgres
stopped redis
removing network "my-project-main"...
preview environment "my-project-main" deleted
```
