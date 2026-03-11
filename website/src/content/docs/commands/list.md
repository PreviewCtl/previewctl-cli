---
title: previewctl list
description: List preview environments.
---

## Usage

```bash
previewctl list [flags]
```

**Aliases:** `ls`

## Description

Lists preview environments tracked by PreviewCtl.

By default, only previews belonging to the current workspace (directory) are shown. Use `--all` to display every preview environment across all workspaces.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-a`, `--all` | `bool` | `false` | List environments across all workspaces |

## Examples

List environments for the current project:

```bash
previewctl list
```

List all environments globally:

```bash
previewctl list --all
```

## Output

```
ID        NAME              BRANCH    STATUS   WORKSPACE              CREATED
abc123    my-project-main   main      active   /home/user/my-project  2026-03-11 10:00:00
def456    my-project-feat   feature   active   /home/user/my-project  2026-03-11 11:30:00
```

| Column | Description |
|--------|-------------|
| ID | Unique identifier for the preview environment |
| NAME | Auto-generated name based on project and branch |
| BRANCH | Git branch the preview was created from |
| STATUS | Current status (`active`, `stopped`, etc.) |
| WORKSPACE | Directory the preview was created in |
| CREATED | Timestamp of creation |
