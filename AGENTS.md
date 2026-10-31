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
- Config: `peterbourgon/ff` — flags (e.g. `--database-host`) or env vars with `TIMEASY_` prefix (`TIMEASY_DATABASE_HOST`, `TIMEASY_KEYCLOAK_HOST`, `TIMEASY_REVENUECAT_API_KEY`, ...). Secrets stay out of source control (`.env` / secret manager, never committed). Subscription check (`pkg/subscription`, `SubscriptionCheckMiddleware`) is disabled when `TIMEASY_REVENUECAT_API_KEY` is unset — self-hosted instances run without billing and never return 402. `TIMEASY_SUBSCRIPTION_ALLOWLIST` (comma-separated Keycloak user IDs) grants full access without a subscription (owner/testers).
- Subscription validation calls RevenueCat API v2: App User ID = Keycloak JWT `sub` on every platform; check = `GET https://api.revenuecat.com/v2/projects/{project_id}/customers/{sub}` → configured entitlement ID in `active_entitlements`. Required server settings when enabled: `TIMEASY_REVENUECAT_API_KEY`, `TIMEASY_REVENUECAT_PROJECT_ID`, `TIMEASY_REVENUECAT_ENTITLEMENT_ID` (internal `entl...` ID; nothing hardcoded). Scoped secret key needs only Customer information → Customers → Read; in-memory TTL cache (~15 min, no webhooks, single-flight collapses concurrent checks). Web purchases run through Paddle via RevenueCat's Web SDK in the Angular client; iOS/Android purchases through StoreKit/Play Billing via `purchases_flutter`. Checkout and subscription management are client-side (`CustomerInfo.managementURL`), so the server only exposes `GET /api/v1/billing/status` (public) and `GET /api/v1/subscription/status` (auth'd) — both intentionally outside the subscription middleware.
- PostgreSQL migrations (golang-migrate) in `pkg/database/postgresql/migrations`, paired up/down files, run automatically on server startup.
- Dev infra: `docker compose -f server/postgresql/development/docker-compose.yaml up -d` (app DB), `-f server/keycloak/development/docker-compose.yaml up -d` (Keycloak on localhost:8180, realm `timeasy`, admin/admin; import `server/keycloak/realm-export.json` and add users manually). `server/docker/docker-compose.yaml` is the full prod-like stack.
- Manual API testing: `server/service-testing/login.sh` (httpie) fetches a Keycloak token into `$TOKEN`.

## Flutter app (`tracking-app`)
- Run `flutter gen-l10n` BEFORE the first build/run/test (translations come from `lib/l10n/app_*.arb`; CI does this too).
- Lint is `flutter analyze`. Android release builds need a keystore + `android/key.properties` (see `tracking-app/README.md`).
- BLoC pattern only (flutter_bloc) — no setState for stateful flows. SQLite via sqflite + `sqflite_migration`; offline-first: local writes → `ChangelogEntry` changelog → bidirectional sync when online.
- Store compliance: app-store builds must NOT link to external purchase flows (Apple 3.1.1 / Play billing rules). The 402 paywall dialog offers native in-app purchase (monthly/yearly via RevenueCat) plus "Restore purchases" — no external links. Web purchases (Paddle) happen on timeasy.org; Paddle never processes store receipts.

## Web client (`webclient/timeasy-webclient`)
- `npm start` (dev server), `npm run lint`, and `npm run test -- --watch=false --browsers=ChromeHeadless` — plain `npm test` hangs in watch mode.
- The public RevenueCat web key is **not** part of the client build: the server returns it in `GET /api/v1/billing/status` (`webApiKey`) when `TIMEASY_REVENUECAT_WEB_API_KEY` is set. Docker images are environment-neutral; self-hosters configure billing on the server only. Secret RC keys never belong in the client.
- Standalone components + Angular signals for state (no BehaviorSubjects); PrimeNG UI only.
- 402 handling: `license-error.interceptor.ts` redirects to `/upgrade`; the page uses `@revenuecat/purchases-js` (embedded Paddle checkout) and "manage subscription" via `CustomerInfo.managementURL`.

## CI / branching
- Path-filtered workflows in `.github/workflows/` (`flutter-ci-build.yaml`, `server-ci-build.yaml`, `webapp-ci-build.yaml`). Flutter CI: analyze → test → APK with Loki `--dart-define` secrets; iOS build only on develop/main. Docker images (`ahilwers/timeasy-server`, `ahilwers/timeasy-webclient`) push only from develop/main/tags; releases via `git tag vX.Y.Z`.
- Branching: `feature/*` → `develop` (PR) → `main`.

## Current priorities
- Polar was replaced by **RevenueCat + Paddle**: the server checks RevenueCat entitlements, the webclient uses `@revenuecat/purchases-js` (Paddle checkout), Flutter uses `purchases_flutter` with native in-app purchases. Remaining: create the RevenueCat project + Paddle sandbox config, fill `revenueCatWebApiKey` in the webclient environments and pass `--dart-define REVENUECAT_IOS_API_KEY=...` / `REVENUECAT_ANDROID_API_KEY=...` for Flutter builds, sandbox verification (purchase, cancel, restore, cross-platform), then production (Paddle domain approval, production keys).
- The old `licensemanager` service is fully removed from this repo; its own repo/deployment still needs decommissioning.
