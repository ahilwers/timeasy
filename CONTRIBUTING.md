# Contributing to timeasy

Thank you for contributing! This repository contains the Flutter mobile app, Go API and Angular web client. Choose the components relevant to your change; you do not need to set up the entire stack for every contribution.

## Repository overview

- [`tracking-app/`](tracking-app/README.md): Flutter app. Its layers include models, repositories, services, BLoCs and views. It is offline-first: local changes are recorded in a changelog and synchronized when online.
- [`server/`](server/README.md): Go REST API, with the module in `server/src/`. Domain models and repository interfaces, use cases, REST transport and PostgreSQL implementations live under `pkg/`. Dependencies and routes are wired in `cmd/service/main.go`.
- [`webclient/timeasy-webclient/`](webclient/timeasy-webclient/README.md): Angular client, using standalone components, signals and the existing UI components.

For changes spanning components, consider API compatibility, authentication, synchronization and the corresponding tests in each affected component.

## Getting started and pull requests

1. Clone the repository and create a working branch from `develop`, for example:

   ```bash
   git clone https://github.com/ahilwers/timeasy.git
   cd timeasy
   git checkout develop
   git pull origin develop
   git checkout -b feature/short-description
   ```

2. Keep the change focused, follow the surrounding code and include tests for changed behavior. Update relevant documentation and translations when needed.
3. Run the applicable checks below. Review your diff for generated files, credentials and unrelated changes.
4. Open a pull request targeting `develop`. Describe the problem and solution, link relevant issues, and list the checks you ran and any that you could not run. Include screenshots for visible UI changes where useful. Respond to review feedback and update the PR as needed. Changes are subsequently promoted from `develop` to `main`.

## Coding and review guidelines

- Discuss substantial features, architectural changes or breaking changes in an issue before starting a large implementation.
- Follow the surrounding structure and naming. Keep business rules out of UI components and HTTP handlers; use the existing service/use-case and repository layers.
- Format changed Go files with `gofmt` and Dart files with `dart format`. Follow the existing TypeScript and template style and run the web lint command; there is no separate repository-wide formatter command.
- Add regression tests for bug fixes and cover new behavior. Pay special attention to authorization, synchronization, migrations and offline data safety.
- Avoid unrelated refactoring, mass formatting and unnecessary dependencies. Explain new dependencies and any compatibility or migration impact in the PR.
- Keep documentation in English and update user-facing translations in the supported languages. Never edit generated files as the primary source of a change.
- Keep discussions constructive and respectful. Do not put credentials, private user data or sensitive security details in public issues or pull requests.

## Flutter app

Install a Flutter SDK compatible with `tracking-app/pubspec.yaml`; CI uses Flutter 3.47.x. Platform development also requires the appropriate Android or iOS tooling. From the repository root:

```bash
cd tracking-app
flutter pub get
flutter gen-l10n
flutter analyze
flutter test
flutter run
```

Run `flutter gen-l10n` before building, running or testing, and after changing localization files. Edit the English and German ARB files in `lib/l10n/` rather than editing generated Dart translations directly; regenerate and include relevant changes to the tracked generated files. Preserve placeholders and metadata when translating.

Use the existing BLoC pattern for stateful flows and preserve offline-first behavior. Android release signing requires a keystore and `android/key.properties`; see the [Flutter README](tracking-app/README.md) for signing, telemetry and in-app purchase setup. Do not commit signing material. App-store builds must use native in-app purchases and must not link to external purchase flows.

## Go server

Install Go as specified by [`server/src/go.mod`](server/src/go.mod), Make, and Docker with a running daemon. From the repository root:

```bash
make -C server/src build
make -C server/src test
```

The full test suite uses Dockertest to start disposable PostgreSQL 15 containers. Docker must be running and able to pull the image; tests can fail fatally without it. The `test` target uses `-p=1` to avoid parallel database containers. `make -C server/src richtest` is an alternative that requires `richgo`.

For local API development, start the development PostgreSQL and Keycloak services from the repository root:

```bash
docker compose -f server/postgresql/development/docker-compose.yaml up -d
docker compose -f server/keycloak/development/docker-compose.yaml up -d
```

Import [`server/keycloak/realm-export.json`](server/keycloak/realm-export.json) into Keycloak at `http://localhost:8180` and create users in the `timeasy` realm. The development compose file uses `admin` / `admin` for the Keycloak administrator; these are local development credentials, not production settings. Check the realm's client settings as described in the [Keycloak development README](server/keycloak/development/README.md). The application database is exposed on port 5432 and its development credentials are in `server/postgresql/development/database.env`.

Run the API from its module directory:

```bash
cd server/src
make run
```

The API defaults to port 8080 and the development database and Keycloak addresses above. Configure different values with `TIMEASY_*` environment variables, command-line flags, or an ignored `server/src/config.txt`; to use a config file, run `make build` and then `./bin/timeasy-server -config config.txt`. Config lines use `flag-name value`, without `=`; precedence is command-line flags, then non-empty environment variables, then config file, then defaults. Empty environment values are ignored. See the [server README](server/README.md) for configuration and optional billing, and the [service-testing README](server/service-testing/README.md) for authenticated manual requests.

For schema changes, add paired up/down migrations in [`server/src/pkg/database/postgresql/migrations/`](server/src/pkg/database/postgresql/migrations/); migrations run at server startup. Wire new handlers, use cases and repositories in `server/src/cmd/service/main.go` as appropriate.

## Web client

Use Node.js 24.15.0 or newer and npm 11 or newer, as specified in `package.json`; `.nvmrc` selects Node 24. From the repository root:

```bash
cd webclient/timeasy-webclient
npm ci
npm run lint
npm run test -- --watch=false
npm run build
npm start
```

The development server is at `http://localhost:4200/`. For real authentication and API calls, run the server and Keycloak as described above; the development environment points to ports 8080 and 8180. The configured unit-test runner is Vitest; `--watch=false` runs it once.

For end-to-end checks, install Playwright Chromium once and run the mock suite:

```bash
npx playwright install chromium
npm run e2e:all
```

Mock mode intercepts the API and Keycloak and does not require Docker or backend services. The [web client README](webclient/timeasy-webclient/README.md) describes the other Playwright commands, including live mode. Live mode uses the local API and Keycloak and expects a development user `user1` with password `password`; it writes and deletes test data, including cleanup of an open time entry. Use a disposable development account and database, never production data.

For UI text, update both `public/i18n/en.json` and `public/i18n/de.json`, preserving interpolation keys. Follow the existing standalone-component and signal patterns.

## Secrets and billing

Never commit real credentials, API secrets, tokens, signing keys or personal test data. Use local ignored configuration or a secret manager, and check what Git will track before committing. Values passed through Flutter `--dart-define` are compiled into the app and must not be treated as secret; browser-side configuration is also public.

Billing is optional for local development: without `TIMEASY_REVENUECAT_API_KEY`, the server disables subscription checks. RevenueCat secret keys belong only on the server; the public web SDK key is supplied by the server at runtime, not baked into the web client. The Flutter app uses public platform-specific RevenueCat keys. See the component READMEs for billing configuration and platform sandbox testing.

## Continuous integration

GitHub Actions workflows live in [`.github/workflows/`](.github/workflows/). They run on pushes or manual dispatch, not on `pull_request` events. Flutter branch builds are path-filtered; server and web builds are not. Feature branches run checks but do not publish Docker images. Run relevant checks locally even when a workflow does not cover your changes.

See **[CI/CD and releases](cicd.md)** for the branch strategy, exact triggers and job dependencies, Docker image tags and the maintainer release checklist. In particular, web lint is currently non-blocking and image publication does not depend on the separate E2E job; a published image alone does not mean all checks passed.
