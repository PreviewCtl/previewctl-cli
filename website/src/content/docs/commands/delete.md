---
title: previewctl delete
description: Delete a preview environment.
---

## Usage

```bash
previewctl delete [name or id]
```

**Aliases:** `rm`

## Description

Deletes a preview environment by name or ID.

If no argument is given, PreviewCtl looks for preview environments in the current workspace. If exactly one exists, it is deleted automatically. If multiple exist, they are listed and you must specify one by name or ID.

This command:

1. Looks up the environment by name, falling back to ID
2. Stops and removes all Docker containers on the preview network
3. Removes the Docker network
4. Cleans up port mappings and environment records from the local database

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `[name or id]` | No | Name or ID of the preview environment to delete. If omitted, auto-selects from the current workspace. |

## Examples

Delete by name:

```bash
previewctl delete my-project-main
```

Delete by ID:

```bash
previewctl delete abc123
```

Auto-select from current workspace:

```bash
previewctl delete
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
