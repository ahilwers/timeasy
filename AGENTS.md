# AGENTS.md — timeasy

Monorepo: Flutter mobile app (production, App Store), Go REST API (WIP), Angular web client (WIP). More context: `CLAUDE.md` (architecture, TODOs), `cicd.md` (branching), `server/README.md` (Keycloak setup).

## Layout
- `tracking-app/` — Flutter app (module `timeasy`). Production. Layers: `lib/models/ → repositories/ → services/ → bloc/ → views/`.
- `server/src/` — Go API, module `timeasy-server`. Entry `cmd/service/main.go`. Clean architecture under `pkg/`: `domain/{model,repository}` → `usecase` → `transport/rest` → `database/postgresql`.
- `webclient/timeasy-webclient/` — Angular 19 + PrimeNG + keycloak-angular.
- `desktop-app/` — gitignored leftovers of an inactive Tauri app; ignore it.

## Go server (`server/src`)
- `make -C server/src build|run|test`; `make richtest` needs `richgo` installed.
- Tests REQUIRE Docker: `pkg/test/testhelper.go` starts a disposable PostgreSQL 15 container via dockertest. Without Docker, `go test ./...` dies with a fatal log. Keep `-p=1` (set by `make test`) to avoid parallel containers.
- New handlers/usecases/repositories are wired manually in `cmd/service/main.go` — unwired code is unreachable.
- Config: `peterbourgon/ff` — flags (e.g. `--database-host`) or env vars with `TIMEASY_` prefix (`TIMEASY_DATABASE_HOST`, `TIMEASY_KEYCLOAK_HOST`, `TIMEASY_LICENSE_MANAGER_HOST`, ...). Subscription check (`pkg/license`, `LicenseCheckMiddleware`) is silently disabled when the license-manager vars are unset.
- PostgreSQL migrations (golang-migrate) in `pkg/database/postgresql/migrations`, paired up/down files, run automatically on server startup.
- Dev infra: `docker compose -f server/postgresql/development/docker-compose.yaml up -d` (app DB), `-f server/keycloak/development/docker-compose.yaml up -d` (Keycloak on localhost:8180, realm `timeasy`, admin/admin; import `server/keycloak/realm-export.json` and add users manually). `server/docker/docker-compose.yaml` is the full prod-like stack.
- Manual API testing: `server/service-testing/login.sh` (httpie) fetches a Keycloak token into `$TOKEN`.

## Flutter app (`tracking-app`)
- Run `flutter gen-l10n` BEFORE the first build/run/test (translations come from `lib/l10n/app_*.arb`; CI does this too).
- Lint is `flutter analyze`. Android release builds need a keystore + `android/key.properties` (see `tracking-app/README.md`).
- BLoC pattern only (flutter_bloc) — no setState for stateful flows. SQLite via sqflite + `sqflite_migration`; offline-first: local writes → `ChangelogEntry` changelog → bidirectional sync when online.

## Web client (`webclient/timeasy-webclient`)
- `npm start` (dev server), `npm run lint`, and `npm run test -- --watch=false --browsers=ChromeHeadless` — plain `npm test` hangs in watch mode.
- Standalone components + Angular signals for state (no BehaviorSubjects); PrimeNG UI only.
- 402 handling already exists: `license-error.interceptor.ts` redirects to `/no-license`.

## CI / branching
- Path-filtered workflows in `.github/workflows/` (`flutter-ci-build.yaml`, `server-ci-build.yaml`, `webapp-ci-build.yaml`). Flutter CI: analyze → test → APK with Loki `--dart-define` secrets; iOS build only on develop/main. Docker images (`ahilwers/timeasy-server`, `ahilwers/timeasy-webclient`) push only from develop/main/tags; releases via `git tag vX.Y.Z`.
- Branching: `feature/*` → `develop` (PR) → `main`.

## Current priorities
- Flutter 402 Payment Required handling is NOT implemented (webclient is): during sync show toast + logout; during login check subscription and show upgrade prompt.
- Go server subscription validation is DONE (`pkg/license` + `LicenseCheckMiddleware`); licensemanager env vars are `TIMEASY_LICENSE_MANAGER_*`.
