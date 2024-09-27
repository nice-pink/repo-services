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