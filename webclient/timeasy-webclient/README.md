# TimeasyWebclient

This project was generated using [Angular CLI](https://github.com/angular/angular-cli) version 19.0.7.

## Development server

To start a local development server, run:

```bash
npm start
```

Once the server is running, open your browser and navigate to `http://localhost:4200/`. The application will automatically reload whenever you modify any of the source files.

## RevenueCat web billing key

The web client does **not** contain a RevenueCat key. When billing is enabled, the server's public `GET /api/v1/billing/status` response includes the instance's **public** RevenueCat Web Billing API key (`webApiKey`), and the client uses it to configure `purchases-js`.

That means:

- the Docker image is environment-neutral — no key is baked in, and self-hosted instances are not tied to the hosted instance's RevenueCat project;
- self-hosters who enable billing set `TIMEASY_REVENUECAT_WEB_API_KEY` on the **server** (see `server/README.md`) instead of rebuilding the web client;
- without billing enabled on the server, the billing UI stays hidden.

> The web key is public by design (it is part of the browser bundle at runtime). Never store a RevenueCat **secret** API key (`sk_...` / `rcsk_...`) in the web client; secret keys belong on the Go server.

## Code scaffolding

Angular CLI includes powerful code scaffolding tools. To generate a new component, run:

```bash
ng generate component component-name
```

For a complete list of available schematics (such as `components`, `directives`, or `pipes`), run:

```bash
ng generate --help
```

## Building

To build the project run:

```bash
ng build
```

This will compile your project and store the build artifacts in the `dist/` directory. By default, the production build optimizes your application for performance and speed.

## Running unit tests

To execute unit tests with the [Karma](https://karma-runner.github.io) test runner, use the following command:

```bash
ng test
```

## Running end-to-end tests

End-to-end tests use [Playwright](https://playwright.dev/) and live in `e2e/`.
There are two modes, controlled by `E2E_MODE`:

- **mock** (default, CI-safe): the Go API and Keycloak are intercepted inside the
  browser context (`e2e/support/`). No backend, database, Keycloak or Docker is
  required, and no real data is written.
- **live**: runs against the real Go API on `http://localhost:8080` and Keycloak
  on `http://localhost:8180` (realm `timeasy`). Login/logout UI tests only run
  in this mode.

```bash
npx playwright install chromium   # once

npm run e2e          # mock mode, desktop Chromium (fast local loop)
npm run e2e:all      # mock mode, desktop + mobile (what CI runs)
npm run e2e:mobile   # mock mode, mobile viewport only
npm run e2e:live     # live mode against local API + Keycloak
npm run e2e:headed   # mock mode with a visible browser
npm run e2e:report   # open the last HTML report
```

The Angular dev server is started automatically by Playwright. Pass `E2E_REUSE=1`
to reuse an already running `ng serve`. Videos are disabled by default; set
`PW_VIDEO=1` to record them for debugging.

## Additional Resources

For more information on using the Angular CLI, including detailed command references, visit the [Angular CLI Overview and Command Reference](https://angular.dev/tools/cli) page.
