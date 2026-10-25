# AGENTS.md — timeasy

Monorepo: Flutter mobile app (production, App Store), Go REST API (WIP), Angular web client (WIP). More context: `cicd.md` (branching), `server/README.md` (Keycloak setup).

## Git rule
- NEVER commit or push unless the user explicitly asks. The user reviews changes and commits themselves.

## Layout
- `tracking-app/` — Flutter app (module `timeasy`). Production. Layers: `lib/models/ → repositories/ → services/ → bloc/ → views/`.
- `server/src/` — Go API, module `timeasy-server`. Entry `cmd/service/main.go`. Clean architecture under `pkg/`: `domain/{model,repository}` → `usecase` → `transport/rest` → `database/postgresql`.
- `webclient/timeasy-webclient/` — Angular 19 + PrimeNG + keycloak-angular.
- `desktop-app/` — gitignored leftovers of an inactive Tauri app; ignore it.

## Go server (`server/src`)
- `make -C server/src build|run|test`; `make richtest` needs `richgo` installed.
- Local dev config: `server/src/config.txt` (gitignored, `flag-name value` pairs for ff.PlainParser) → `./bin/timeasy-server -config config.txt`. Priority: CLI flags > config file > `TIMEASY_*` env vars.
- Tests REQUIRE Docker: `pkg/test/testhelper.go` starts a disposable PostgreSQL 15 container via dockertest. Without Docker, `go test ./...` dies with a fatal log. Keep `-p=1` (set by `make test`) to avoid parallel containers.
- New handlers/usecases/repositories are wired manually in `cmd/service/main.go` — unwired code is unreachable.
- Config: `peterbourgon/ff` — flags (e.g. `--database-host`) or env vars with `TIMEASY_` prefix (`TIMEASY_DATABASE_HOST`, `TIMEASY_KEYCLOAK_HOST`, `TIMEASY_POLAR_ACCESS_TOKEN`, ...). Secrets stay out of source control (`.env` / secret manager, never committed). Subscription check (`pkg/subscription`, `SubscriptionCheckMiddleware`) is disabled when `TIMEASY_POLAR_ACCESS_TOKEN` is unset — self-hosted instances run without billing and never return 402.
- Subscription validation calls Polar directly: customer linked to Keycloak user via `external_customer_id` (= JWT `sub`, set when the server creates the checkout); check = `GET {polar}/v1/customers/external/{sub}` → `GET .../{id}/state` → `active_subscriptions`; in-memory TTL cache (~15 min, no webhooks, single-flight collapses concurrent checks); email-match fallback backfills the external ID. Billing endpoints (`/api/v1/billing/status`, `/api/v1/upgrade`, `/api/v1/subscription/status`, `/api/v1/billing/portal`) are auth'd but intentionally NOT behind the subscription middleware (a user without a subscription must be able to upgrade).
- PostgreSQL migrations (golang-migrate) in `pkg/database/postgresql/migrations`, paired up/down files, run automatically on server startup.
- Dev infra: `docker compose -f server/postgresql/development/docker-compose.yaml up -d` (app DB), `-f server/keycloak/development/docker-compose.yaml up -d` (Keycloak on localhost:8180, realm `timeasy`, admin/admin; import `server/keycloak/realm-export.json` and add users manually). `server/docker/docker-compose.yaml` is the full prod-like stack.
- Manual API testing: `server/service-testing/login.sh` (httpie) fetches a Keycloak token into `$TOKEN`.

## Flutter app (`tracking-app`)
- Run `flutter gen-l10n` BEFORE the first build/run/test (translations come from `lib/l10n/app_*.arb`; CI does this too).
- Lint is `flutter analyze`. Android release builds need a keystore + `android/key.properties` (see `tracking-app/README.md`).
- BLoC pattern only (flutter_bloc) — no setState for stateful flows. SQLite via sqflite + `sqflite_migration`; offline-first: local writes → `ChangelogEntry` changelog → bidirectional sync when online.
- Store compliance: app-store builds must NOT link to external purchase flows (Apple 3.1.1 / Play billing rules). The 402 paywall dialog is information-only in store builds; the "Subscribe" CTA that opens the web client exists only in dev builds. Subscriptions are purchased on timeasy.org in the browser — Polar does not process store receipts.

## Web client (`webclient/timeasy-webclient`)
- `npm start` (dev server), `npm run lint`, and `npm run test -- --watch=false --browsers=ChromeHeadless` — plain `npm test` hangs in watch mode.
- Standalone components + Angular signals for state (no BehaviorSubjects); PrimeNG UI only.
- 402 handling already exists: `license-error.interceptor.ts` redirects to `/no-license`.

## CI / branching
- Path-filtered workflows in `.github/workflows/` (`flutter-ci-build.yaml`, `server-ci-build.yaml`, `webapp-ci-build.yaml`). Flutter CI: analyze → test → APK with Loki `--dart-define` secrets; iOS build only on develop/main. Docker images (`ahilwers/timeasy-server`, `ahilwers/timeasy-webclient`) push only from develop/main/tags; releases via `git tag vX.Y.Z`.
- Branching: `feature/*` → `develop` (PR) → `main`.

## Current priorities
- Polar subscription integration is implemented end-to-end (server `pkg/subscription` + webclient billing funnel + Flutter 402 paywall); remaining: sandbox verification with a real Polar org (product ID, access token, webhook-free TTL flow), then `TIMEASY_POLAR_ENV=production` go-live.
- The old `licensemanager` service is fully removed from this repo; its own repo/deployment still needs decommissioning.
