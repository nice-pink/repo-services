.PHONY: build mcp-server test audit-stdout clean

# Build all binaries (mirrors the existing build script)
build:
	mkdir -p bin/
	for d in cmd/*/; do \
		[ -L "$${d%/}" ] && continue; \
		echo "build $$d"; \
		cd "$$d" && go build -o ../../bin && cd ../..; \
	done

# Build only the MCP server binary
mcp-server:
	mkdir -p bin/
	cd cmd/mcp-server && go build -o ../../bin/mcp-server .

# Run tests
test:
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...

# audit-stdout: verify that pkg/util, pkg/manifest, and pkg/runner contain
# no fmt.Print*, println(, os.Stdout writes, or goutil/log imports.
# Uses inverted grep: grep finding matches is the FAILURE case (exit 0 means
# something was found → bad). The '! grep ...' pattern inverts this.
#
# Exit 0 = clean (nothing found)  Exit 1 = violations found (build should fail)
audit-stdout:
	@echo "Auditing pkg/util, pkg/manifest, pkg/runner for stdout writes and goutil/log imports..."
	@! grep -rn \
		-e 'fmt\.Print' \
		-e 'fmt\.Fprint[^l]' \
		-e 'os\.Stdout\.Write' \
		-e 'println(' \
		-e '"github.com/nice-pink/goutil/pkg/log"' \
		pkg/util/ pkg/manifest/ pkg/runner/ \
		|| (echo "FAIL: stdout write or goutil/log import found in pkg/{util,manifest,runner}" && exit 1)
	@echo "PASS: audit-stdout clean"

clean:
	rm -f bin/deploy bin/promote bin/git bin/mcp-server cover.out
