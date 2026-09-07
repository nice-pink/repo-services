.PHONY: build test audit-stdout clean

# Build all binaries (mirrors the existing build script)
build:
	mkdir -p bin/
	for d in cmd/*/; do \
		[ -L "$${d%/}" ] && continue; \
		echo "build $$d"; \
		cd "$$d" && go build -o ../../bin && cd ../..; \
	done

# Run tests
test:
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...

# audit-stdout: verify that pkg/util, pkg/manifest, and pkg/runner contain
# no fmt.Print*, println(, os.Stdout writes, or goutil/log imports.
# Kept after the MCP server moved to nice-pink/ops-repo-mcp: that server
# consumes these packages over stdio, and any stdout write here corrupts its
# JSON-RPC framing.
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
	rm -f bin/deploy bin/promote bin/git cover.out
