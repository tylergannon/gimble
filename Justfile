origin := env("ORIGIN", "http://127.0.0.1:8080")
base_url := env("BASE_URL", origin)

build:
    go mod tidy
    cd web && pnpm install
    go generate ./...
    cd web && ORIGIN='{{origin}}' pnpm exec vp build
    go build -ldflags '-X main.origin={{origin}}' -o bin/gimble ./cmd

dev-web:
    cd web && ORIGIN='{{origin}}' pnpm exec vp dev --host 127.0.0.1 --port 5173 --strictPort

dev-go:
    go run ./cmd --proxy http://127.0.0.1:5173

e2e run="run":
    cd e2e && pnpm install
    cd e2e && pnpm exec playwright install chromium
    cd e2e && BASE_URL="{{base_url}}" SKGO_E2E_RUN={{run}} pnpm test
