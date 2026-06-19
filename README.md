# Run

## Docker

```
#  deploy
docker run --entrypoint /app/deploy ghcr.io/nice-pink/repo-services:latest -help

#  promote
docker run --entrypoint /app/promote ghcr.io/nice-pink/repo-services:latest -help
```

# Build command line executables

## Build single executable

1. Add executables as sub-folder into `cmd` folder. E.g. `cmd/exec`
2. Open terminal and `cd` to base folder of this repo.
3. Type `./build NAME_OF_EXECUTABLE`. E.g. `./build exec`
4. Executable will be created in `bin/NAME_OF_EXECUTABLE`. E.g. `bin/exec`
5. Run executable. E.g. `bin/exec`

## Build all

1. Add executables as sub-folder into `cmd` folder. E.g. `cmd/exec`
2. Open terminal and `cd` to base folder of this repo.
3. Type `./build`.
4. All executables will be created in `bin` folder.

# Test

## Unit tests

```
go test ./...
```

# Run

deploy app:

```
bin/deploy -app test-app -namespace test -env prod -tag abcdef -srcFolder examples/repo -base base/resources

bin/deploy -app test-envs -namespace test -env prod -tag abcdef -srcFolder examples/repo -base base/resources -exceptionalAppsFile examples/exceptional_deployments.yaml
```

promote dev app to prod:

```
bin/promote -app test-app-image -namespace test -env prod -srcEnv dev -srcFolder examples/repo -base base/resources -exceptionalAppsFile examples/exceptional_deployments.yaml
```

# MCP server

`cmd/mcp-server` exposes the deploy / promote / rollback flows over the Model
Context Protocol so an LLM client (Claude Code, Cursor, etc.) can drive image
tag changes against an ops repo. The server speaks stdio and is launched per
session by the client.

The server **mutates files but never commits or pushes**. Every successful
response includes a `commitDirective` (filesToStage, suggestedCommitMessage,
gitCommands) that the calling LLM is expected to execute. This keeps the
human-visible audit trail in git history and lets the operator inspect the
diff before publishing.

## Build

```
make mcp-server
```

Produces `bin/mcp-server`.

## Configure

Point your MCP client at the binary. A copy-paste example lives in
`.mcp.json.example`; minimal config:

```json
{
  "mcpServers": {
    "deploy-promote": {
      "command": "/absolute/path/to/repo-services/bin/mcp-server",
      "args": [],
      "env": {
        "MCP_OPS_REPO_PATH":   "/absolute/path/to/your/ops-repo",
        "MCP_ENV_ALLOWLIST":   "dev,staging,prod",
        "DS_BASE":             "base/apps",
        "DS_PATH_SCHEME":      "{base}/{app}/{env}",
        "DS_IMAGE_FILE_NAME":  "deployment.yaml",
        "DS_SRC_ENV":          "dev"
      }
    }
  }
}
```

### Required environment

| Variable | Description |
|----------|-------------|
| `MCP_OPS_REPO_PATH` | Absolute path to a local clone of the ops repo. Required; server exits with `REPO_NOT_FOUND` if missing. |

### Optional environment

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_ENV_ALLOWLIST` | _(empty — all envs accepted)_ | Comma-separated list of envs the server will operate on. Strongly recommended. |
| `MCP_LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error`. Logs always go to stderr; stdout is reserved for MCP framing. |
| `MCP_LOCK_TIMEOUT` | `30s` | Max wait for the per-repo lock. Returns `LOCK_TIMEOUT` on expiry. |
| `MCP_RUNNER_TIMEOUT` | `60s` | Max wall-clock for a runner invocation. Returns `RUNNER_TIMEOUT` on expiry; the goroutine is leaked and holds the lock until it finishes. |
| `MCP_GIT_SSH_KEY_PATH` | _(none)_ | SSH key used for `git fetch` / `git pull` against the ops repo's remote. |
| `MCP_GIT_TOKEN` | falls back to `GITHUB_TOKEN` | HTTPS token for the same. |
| `MCP_GIT_USER` | `mcp-server` | Author name for any commit objects the underlying runner creates. |
| `MCP_GIT_EMAIL` | _(empty)_ | Author email. |
| `DS_BASE` | _(empty)_ | Base folder inside the ops repo (e.g. `base/apps`). |
| `DS_NAMESPACE` | _(empty)_ | k8s namespace, also expanded into `DS_PATH_SCHEME`. |
| `DS_PATH_SCHEME` | `{base}/{namespace}/{app}/{env}` | Template for the manifest folder. Must contain `{app}` and `{env}`; `..` is rejected. |
| `DS_IMAGE_FILE_NAME` | `deployment.yaml` | Filename of the manifest the server rewrites. No `/` or `..`. |
| `DS_IMAGE_HISTORY_FILE_NAME` | _(empty)_ | If set, the server appends the new tag to this file (relative to the manifest folder) on successful deploy/promote/rollback. |
| `DS_EXCEPTIONAL_APPS_FILE` | _(empty)_ | Path to a YAML file describing apps whose image name / path deviates from the default scheme. Must exist on disk if set. |
| `DS_SRC_ENV` | `staging` | Default source env for `promote` when the caller doesn't pass one. |

## Tools

All three tools take `app` and an environment, share the same input validation
(`^[a-z0-9][a-z0-9-]*$` for app/env/namespace, `^[a-zA-Z0-9_.-]+$` for tag),
the same dirty-tree / branch-ahead pre-checks, and the same per-repo lock.

### `deploy`

Set `app`'s image tag to `tag` in `env`.

| Field | Required | Notes |
|-------|----------|-------|
| `app` | yes | Application name. |
| `env` | yes | Target environment; must pass the allowlist if one is set. |
| `tag` | yes | Image tag to write. |
| `namespace` | no | Overrides `DS_NAMESPACE`. |
| `dryRun` | no | If true, runs the pre-pull and computes affected paths without mutating files. |

### `promote`

Copy the current image tag of `app` from `srcEnv` to `destEnv`.

| Field | Required | Notes |
|-------|----------|-------|
| `app` | yes | |
| `destEnv` | yes | Target env. Must differ from `srcEnv`. |
| `srcEnv` | no | Defaults to `DS_SRC_ENV`. |
| `namespace` | no | |
| `dryRun` | no | Reads and reports the resolved tag without mutating. |

### `rollback`

Revert `app`'s manifest in `env` to the image tag it held in the **commit
immediately preceding the most recent commit that touched the manifest file**.
The previous tag is read from git history; no caller-supplied tag is needed.

| Field | Required | Notes |
|-------|----------|-------|
| `app` | yes | |
| `env` | yes | |
| `namespace` | no | |
| `dryRun` | no | Computes the previous tag and the change-set warning without mutating. |

Response fields specific to `rollback`:

- `currentTag` / `previousTag` — what's in HEAD and what's being reverted to.
- `lastCommit` / `parentCommit` / `lastCommitMessage` — the commit being
  rolled back, and its parent (the rollback target).
- `multiLineChange` — `true` if the last commit touched lines in the manifest
  beyond the image-tag substitution. When true, the response also includes a
  human-readable `warning` and a `nonTagLineChanges` count. The calling LLM
  should surface this to the operator before executing the commit directive,
  because a tag-only revert won't restore those other changes.

Errors specific to `rollback`:

- `NO_PREVIOUS_VERSION` — the manifest has no prior history (initial commit
  only), or the previous tag couldn't be extracted. Deploy an earlier tag
  explicitly via `deploy` instead.

## Response shape

Success:

```json
{
  "success": true,
  "dryRun": false,
  "app": "poma-mcp",
  "env": "prod",
  "previousTag": "v0.0.7",
  "multiLineChange": false,
  "nonTagLineChanges": 0,
  "opsRepoPath": "/abs/path",
  "commitDirective": {
    "action": "git-commit-and-push",
    "filesToStage": ["base/apps/poma-mcp/prod/deployment.yaml"],
    "suggestedCommitMessage": "Rollback poma-mcp(prod) to version: v0.0.7",
    "gitCommands": ["git -C /abs/path add -- ...", "git -C /abs/path commit -m ...", "git -C /abs/path push"]
  },
  "recoveryHint": {
    "detectCommand": "git -C /abs/path status --short -- ...",
    "discardCommand": "git -C /abs/path checkout -- ...",
    "completeCommand": "git -C /abs/path add -- ... && git -C /abs/path commit -m ... && git -C /abs/path push"
  },
  "runnerOutput": "<runner slog lines>"
}
```

Error:

```json
{
  "success": false,
  "errorCode": "DIRTY_REPO",
  "errorMessage": "working tree has uncommitted changes: [...]",
  "recoveryHint": "complete or discard prior changes before retrying (see recoveryHint)",
  "runnerOutput": ""
}
```

Error codes: `INVALID_INPUT`, `ENV_NOT_ALLOWED`, `SAME_ENV`, `PATH_ESCAPE`,
`REPO_NOT_FOUND`, `CONFIG_ERROR`, `DIRTY_REPO`, `BRANCH_AHEAD`, `PULL_FAILED`,
`NO_CURRENT_TAG`, `NO_PREVIOUS_VERSION`, `RUNNER_FAILED`, `RUNNER_PANIC`,
`RUNNER_TIMEOUT`, `LOCK_TIMEOUT`.

