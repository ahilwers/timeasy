# timeasy web client

The Angular web client provides a browser interface for timeasy projects, time entries, and statistics. It requires the Go API and Keycloak for normal authenticated use; unlike the mobile app, it is not an offline-first client. The web client and server are works in progress. See the [repository overview](../../README.md), [server instructions](../../server/README.md), [mobile app](../../tracking-app/README.md), and [contribution guide](../../CONTRIBUTING.md).

> **Self-hosters:** The production Docker image compiles `apiUrl` and `authUrl` from `src/environments/environment.ts`, currently pointing to the hosted timeasy.org services. Configure URLs reachable by your users' browsers and rebuild the image for your instance. Of these settings, only the public RevenueCat web billing key is supplied at runtime by the server. See [endpoint configuration](#endpoint-configuration).

## Prerequisites and development

Use Node.js and npm versions satisfying `package.json` (`node >=24.15.0`, `npm >=11`); CI uses Node 24. From `webclient/timeasy-webclient/`:

```bash
npm ci
npm start
```

Open `http://localhost:4200/`. The development server reloads as source files change. To use the application, start the API, PostgreSQL, and Keycloak, import the `timeasy` realm, and create a user as described in the [server setup](../../server/README.md). Keycloak uses the `timeasy-webclient` public client. Ensure its allowed redirect URIs, web origins, and logout redirects match the browser's URL; the imported realm includes localhost:4200. The browser must be able to reach both the API and Keycloak, with suitable CORS and authentication configuration.

## Endpoint configuration

Endpoints are compiled into the client from `src/environments/`:

| Configuration | API | Keycloak |
| --- | --- | --- |
| Development (`npm start`) | `http://localhost:8080/api/v1` | `http://localhost:8180` |
| Production (`npm run build`) | `https://api.timeasy.org/api/v1` | `https://auth.timeasy.org` |

`angular.json` replaces `environment.ts` with `environment.development.ts` for the development configuration. For a different deployment, set the appropriate URLs in these files and rebuild; they are not runtime environment variables. `localhost` refers to the browser's machine, not the server or container. If using another host or port, also configure the corresponding Keycloak client origins and redirects.

## Build and checks

```bash
npm run lint
npm run test -- --watch=false
npm run build
```

Unit tests use Angular's Vitest runner; `npm test` without the watch option runs in watch mode. The default build is optimized for production and outputs to `dist/timeasy-webclient/`. The included `Dockerfile` builds the client and serves it with nginx; `nginx.conf` falls back to `index.html` for application routes. Other static hosting must also support client-side route fallback. The image embeds the production API and auth URLs, even though the web billing key is obtained at runtime.

### Temporary security override

`package.json` overrides transitive `minimatch` v10 with `^11.0.5` to fix [GHSA-g3hv-75wg-m249](https://github.com/advisories/GHSA-g3hv-75wg-m249). The current ESLint, `@eslint/config-array`, and typescript-eslint releases still require v10. These lint-only consumers use the retained `Minimatch` constructor, `match()` method, `pattern` property, and `minimatch()` function, not the APIs removed in v11 (such as `makeRe()` and `AST`). Node requirements remain compatible with this project's Node 24 baseline. Remove the override once upstream dependency ranges accept a patched release; rerun lint, tests, and build when updating it.

## End-to-end tests

Playwright tests in `e2e/` start the Angular development server automatically. Install Chromium once, then run:

```bash
npx playwright install chromium
npm run e2e          # mocked API and authentication, desktop Chromium
npm run e2e:all      # mocked API and authentication, desktop and mobile viewport
npm run e2e:mobile   # mocked, mobile viewport
npm run e2e:headed   # mocked, visible browser
npm run e2e:live     # real local API and Keycloak
npm run e2e:report   # open the last HTML report
```

Mock mode is the default and needs no backend, database, Keycloak, or Docker. CI runs both desktop and mobile mocked projects. Set `E2E_REUSE=1` to reuse an existing dev server, `E2E_PORT` to change its port, or `PW_VIDEO=1` to retain video on failure.

Live mode requires the API at `localhost:8080`, Keycloak at `localhost:8180`, realm `timeasy`, the dev server at `localhost:4200`, and the local test account expected by [`e2e/auth.setup.ts`](e2e/auth.setup.ts) and [`e2e/login.spec.ts`](e2e/login.spec.ts). Its setup uses a direct-access token grant. **Use a disposable local account and dataset:** live tests write real data, and setup stops/deletes an open timer and deletes projects and entries named `e2e-*`. The stored browser session in `playwright/.auth/` contains authentication state and must not be shared or committed.

## Localization

Translations use `@ngx-translate` and are loaded from `public/i18n/en.json` and `public/i18n/de.json`. Keep keys in both files aligned. The app prefers the authenticated Keycloak profile locale, then the browser language, with English as fallback. Date/number formatting also uses browser locale settings. No Flutter localization generation step is needed for this client.

## Optional billing

Billing is disabled by default on the server. When enabled, the server's public `GET /api/v1/billing/status` returns the public RevenueCat Web Billing SDK key as `webApiKey` when configured. Set `TIMEASY_REVENUECAT_WEB_API_KEY` **on the server**, not in the Angular build. Without the public web key, web checkout is unavailable even if server subscription checking is enabled. The server's RevenueCat secret key must never be placed in browser code or client configuration.

The client uses `@revenuecat/purchases-js` for embedded Paddle checkout and the returned management URL for subscriptions. A `402` response redirects to `/upgrade`. The authenticated Keycloak user ID is the RevenueCat App User ID shared with the mobile clients; cross-platform access requires correctly configured products and a common entitlement. Mobile purchases use the native stores, not Paddle. See [server billing setup](../../server/README.md#billing-optional) and the [mobile billing instructions](../../tracking-app/README.md#billing-and-logging).
