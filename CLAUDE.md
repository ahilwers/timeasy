# CLAUDE.md — timeasy

This is the main time tracking application. It consists of a Flutter mobile app (production), a Go backend server (WIP), and an Angular web client (WIP).

**Platform context:** This repo is part of the Timeasy platform. Subscription management and payment processing are handled by a separate service (`licensemanager`). The Go server validates subscriptions by calling the licensemanager API.

---

## Repository Structure

```
timeasy/
├── tracking-app/          # Flutter mobile app — PRODUCTION (App Store)
├── server/                # Go backend server — WIP
│   └── src/               # Go source code
├── webclient/             # Angular web client — WIP
│   └── timeasy-webclient/
└── desktop-app/           # Desktop client (not active)
```

---

## Common Commands

### Flutter App (`tracking-app/`)
```bash
flutter gen-l10n          # REQUIRED before first build (generates translations)
flutter run               # Run in debug mode
flutter test              # Run tests
flutter build ios         # iOS release build
flutter build apk         # Android release build
```

### Go Server (`server/src/`)
```bash
make build                # Build the server
make run                  # Run the server
make test                 # Run tests
go test -v ./...          # Run tests verbose
```

### Angular Web Client (`webclient/timeasy-webclient/`)
```bash
ng serve                  # Start dev server
ng build                  # Production build
ng test                   # Run tests
ng generate component X   # Generate component
```

---

## Architecture

### Flutter App
- **Pattern:** BLoC (flutter_bloc) — no setState in complex flows
- **DB:** SQLite via sqflite, with automatic migration support
- **Auth:** OpenID Connect
- **Sync:** Offline-first. Local writes first, then bidirectional sync when online.
  - `ChangelogEntry` model tracks local changes
  - `SyncData` handles sync protocol with server
- **Layers:** models/ → repositories/ → services/ → bloc/ → views/

### Go Server
- **Pattern:** Clean architecture, domain-driven design
- **Auth:** Keycloak JWT validation
- **DB:** PostgreSQL with golang-migrate
- **Config:** Env vars prefixed with `TIMEASY_`
- **Docker:** `ahilwers/timeasy-server`
- **Layers:** domain/ → usecase/ → transport/rest/ → database/postgresql/

### Angular Web Client
- **Pattern:** Standalone components, Angular signals for state
- **UI:** PrimeNG components
- **402 handling:** Already implemented — redirects to /no-license on 402 response

---

## Domain Models

- **Project** — billable work category
- **TimeEntry** — individual tracked time record
- **Team** — user group
- **ChangelogEntry** — tracks local changes for sync
- **SyncData** — bidirectional sync payload

---

## Current TODOs

### 🔴 Flutter App — 402 Payment Required handling
When the server returns 402 (subscription required):
- During sync: show toast notification + logout user
- During login: check subscription status, show upgrade prompt if inactive
- Upgrade CTA should link to: `https://timeasy.app/signup?programId=timeasy-prod` (or equivalent)
- Reference: webclient already handles 402 → /no-license redirect

### 🔴 Go Server — Subscription validation
Every sync endpoint must validate the user's subscription:
1. Call `GET {TIMEASY_LICENSEMANAGER_URL}/api/v1/subscriptions/check?userId={keycloakUserId}&productId=timeasy`
2. If response is `{"active": false}` → return `402 Payment Required`
3. Add `TIMEASY_LICENSEMANAGER_URL` to config

### 🟡 Angular Web Client — /no-license page
- Add a proper /no-license page with upgrade CTA
- CTA links to timeasy-website signup URL
- Page should be accessible without auth

---

## Development Setup

### Flutter Requirements
- Run `flutter gen-l10n` before first build
- Requires keystore for Android builds (see tracking-app/README.md)

### Go Server Requirements
- Keycloak for auth (docker-compose in `keycloak/`)
- PostgreSQL (docker-compose in `server/`)
- Set env vars with `TIMEASY_` prefix

### Dev Infrastructure (docker-compose)
```bash
cd server/
docker-compose up -d      # Starts PostgreSQL + Keycloak
```

---

## Integration Points

- **licensemanager:** Go server calls `/api/v1/subscriptions/check` to validate subscriptions
- **timeasy-website:** Links to website for signup/upgrade flows
- **Keycloak:** Both Flutter app and Go server authenticate via Keycloak

---

## Code Conventions

- Go: Clean architecture strictly. No business logic in HTTP handlers. Errors wrapped with context.
- Flutter: BLoC pattern for all state. Repository pattern for data access.
- Angular: Signals not BehaviorSubjects. Standalone components. PrimeNG UI only.
- Tests: Table-driven tests in Go. Every service should have tests.
- Never commit `.env` files or secrets.
