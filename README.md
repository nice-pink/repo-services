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

The deploy / promote / rollback MCP server moved to its own repository:
**[nice-pink/ops-repo-mcp](https://github.com/nice-pink/ops-repo-mcp)**. It
consumes this repo as a Go module dependency (`pkg/runner`, `pkg/manifest`,
`pkg/util`, `pkg/exceptional`) and carries its own prebuilt binaries and
install script.

Two things in here still exist for its benefit:

- `make audit-stdout` — `pkg/{util,manifest,runner}` must never write to
  stdout. The MCP server speaks JSON-RPC over stdio, so any stray
  `fmt.Print` in those packages corrupts its framing. Run it before tagging.
- `examples/` — `ops-repo-mcp` keeps a copy of this fixture tree for its
  stdout-purity regression test. Structural changes here should be mirrored
  there.
