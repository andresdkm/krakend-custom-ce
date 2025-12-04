# Agent Instructions

## Build & Test
- **Unit Tests:** `go test -v ./...` (Primary development loop)
- **Single Test:** `go test -v ./pkg/hotels/transformers -run TestName`
- **Build Plugins:** Uses Docker. Run `./src/test_plugin.sh` to build & verify plugins.
- **Docker Build:** `make buildg` (Builds KrakenD CE image)
- **Lint:** `go vet ./...`

## Code Style
- **Go Version:** 1.25.3
- **Imports:** Grouped: standard lib first, then 3rd party/local.
- **Naming:** CamelCase. Plugin logic often uses `init()` for registration.
- **Error Handling:** Explicit `if err != nil`. Wrap errors for context.
- **Types:** Use `interface{}`/`reflect` for dynamic data; strong types for core logic.
- **Structure:** `pkg/` (shared logic), `src/` (plugin implementations).
- **Formatting:** Standard `gofmt`.
