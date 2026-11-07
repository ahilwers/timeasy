# timeasy-server

This is the timeasy REST API. The server and web client are works in progress.

## Self-hosting

See the [self-hosting guide](docker/README.md) and the linked [example Compose file](docker/docker-compose.yaml). The example needs changes before use: it does not include Keycloak or TLS, and the web client's production API and authentication URLs are embedded at build time. Rebuild the web image for your own browser-reachable URLs; a server environment change alone does not reconfigure it.

## Development

The [development PostgreSQL](postgresql/development/docker-compose.yaml) and [Keycloak](keycloak/development/docker-compose.yaml) Compose files use development credentials and are not production configurations. Import [realm-export.json](keycloak/realm-export.json), review the clients and add users. The mobile app's callback is `org.timeasy.app://callback`, which must be added to the tracking client's valid redirect and post-logout redirect URIs; the export currently contains `http://localhost:4000/` instead.

From the repository root, build or test the Go server with:

```bash
make -C server/src build
make -C server/src test
```

Tests require Docker: the test helper starts disposable PostgreSQL containers. `make test` runs them with `-p=1`. The [service-testing](service-testing/) scripts are development helpers; `login.sh` contains a hardcoded client secret and prints the supplied password and token. Do not use it with production credentials or share its output.

### Database migrations

The server applies embedded [golang-migrate](https://github.com/golang-migrate/migrate) migrations automatically at startup. The account needs schema permissions; back up before upgrading. For creating migrations, see the [migrate CLI instructions](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate).

### Docker dependency security notes

As of 2026-11-07, the Docker CLI dependency is pinned to at least `v29.2.0`, which fixes GHSA-p436-gjf2-799p and uses the split Moby API/client modules instead of importing the legacy Docker Engine module.

`github.com/docker/docker v28.5.2+incompatible` remains in the module graph and `go.sum` through `golang-migrate`'s upstream tests (`database/postgres.test` → `github.com/dhui/dktest`). It is not imported by the server or our tests; verify with `go list -deps -test ./...`. Do not manually delete its checksums: `go mod tidy` retains them for that upstream test dependency.

Dependabot alerts [#256](https://github.com/ahilwers/timeasy/security/dependabot/256), [#257](https://github.com/ahilwers/timeasy/security/dependabot/257), [#258](https://github.com/ahilwers/timeasy/security/dependabot/258), and [#259](https://github.com/ahilwers/timeasy/security/dependabot/259) concern Docker Engine code, with no patched release of the legacy `github.com/docker/docker` module. Removing its direct requirement from our `go.mod` does not patch the upstream dependency or the host daemon. Track upstream migration away from this module and keep the Docker daemon used for development/tests updated independently. These alerts have not been dismissed.

## Server configuration

Settings can be supplied as `TIMEASY_*` environment variables, command-line flags or a [local config file](#local-config-file). The Compose example overrides some development-oriented defaults:

| Environment variable | Server default | Purpose |
| --- | --- | --- |
| `TIMEASY_SERVER_PORT` | `8080` | HTTP listen port |
| `TIMEASY_DATABASE_HOST` / `TIMEASY_DATABASE_PORT` | `localhost` / `5432` | PostgreSQL address from the server's network |
| `TIMEASY_DATABASE_NAME` / `TIMEASY_DATABASE_USER` / `TIMEASY_DATABASE_PASSWORD` | `timeasy` / `dbuser` / `dbpassword` | Database and credentials; change for deployment |
| `TIMEASY_KEYCLOAK_HOST` / `TIMEASY_KEYCLOAK_REALM` | `http://localhost:8180` / `timeasy` | Keycloak base URL and realm used to fetch signing keys |
| `TIMEASY_LOG_LEVEL` | `INFO` | `DEBUG`, `INFO`, `WARN`, or `ERROR` |
| `TIMEASY_LOKI_ENDPOINT` / `TIMEASY_LOKI_BEARER_TOKEN` | empty | Optional Loki export; both are needed |

See [configuration.go](src/pkg/configuration/configuration.go) for all flags and defaults, including sync scheduler settings. Billing and rate-limit settings are described below. In Compose, `localhost` inside the API container refers to that container, not the host or another service.

## Billing (optional)

Billing is optional. When `TIMEASY_REVENUECAT_API_KEY` is unset, subscription checking is disabled and the subscription middleware does not return 402. Authentication and the other deployment requirements still apply.

To monetize a hosted instance, subscription entitlements are checked against [RevenueCat](https://www.revenuecat.com). Purchases can come from the web client (Paddle as merchant of record) or from in-app purchases (App Store / Google Play). All products are attached to the same RevenueCat entitlement, and every platform uses the Keycloak user ID as RevenueCat App User ID — so one subscription unlocks the web client and the mobile apps alike.

```text
TIMEASY_REVENUECAT_API_KEY=<V2 secret API key>     # setting this enables billing
TIMEASY_REVENUECAT_WEB_API_KEY=<public web SDK key> # optional; enables web purchases
TIMEASY_REVENUECAT_PROJECT_ID=<project ID>         # required when billing is enabled
TIMEASY_REVENUECAT_ENTITLEMENT_ID=<entl... ID>     # required; internal ID, not lookup key
TIMEASY_SUBSCRIPTION_CACHE_TTL_MINUTES=15          # how long subscription states are cached
```

**RevenueCat setup:**
1. Create a project and add a **Web** config for **Paddle** (sandbox first). Paste a Paddle API key with the permissions RevenueCat asks for — including `Customer portal sessions (write)`, which RevenueCat needs for the "manage subscription" links. Keep "Automatic purchase tracking" and "Autogenerated user IDs" enabled.
2. Create the products/prices in Paddle (e.g. monthly and yearly), import the prices into RevenueCat, and attach both to your chosen entitlement (for example, `timeasy_pro`). The server requires its internal `entl...` ID, not that lookup name.
3. Connect the App Store and Google Play apps in the same project and attach their subscription products to the same entitlement. The current Flutter client expects the lookup key `timeasy_pro`; use that key with the unmodified app, while configuring its corresponding internal `entl...` ID on the server.
4. Create a **V2 Secret API Key** (Project settings → API keys) for the server. Grant only **Customer information → Customers → Read**. The server's v2 customer lookup includes `active_entitlements`, so it does not need separate Subscriptions or Purchases access.
5. Configure all three values on the server: secret API key, Project ID (Project settings → General), and the entitlement's internal `entl...` ID (Product catalog → Entitlements). These identifiers are required when billing is enabled and are not hardcoded; without the API key, self-hosted mode remains unchanged.

**Web purchases:** set `TIMEASY_REVENUECAT_WEB_API_KEY` to the **public** Web Billing SDK key from the RevenueCat web config. When billing is enabled, the server returns it in `GET /api/v1/billing/status`; the billing key does not need to be baked into the web build. The API and authentication URLs are different: those are embedded in the web build and must be customized for self-hosting. Without the public billing key, web purchases are unavailable; mobile purchases require their own configured public SDK keys.

The server checks `GET https://api.revenuecat.com/v2/projects/{project_id}/customers/{keycloak-sub}` and looks for the configured entitlement ID in `active_entitlements`. Active results are cached in memory per server instance for `TIMEASY_SUBSCRIPTION_CACHE_TTL_MINUTES` (15 minutes by default); inactive results are cached for at most one minute. Provider updates, including cancellations and refunds, may not be reflected until the cache expires. RevenueCat lookup errors cause the subscription middleware to **fail open** (allow access and log the error), including configuration or provider errors; a customer-not-found response is treated as inactive. The force-refresh status endpoint instead returns 502 on lookup errors.

**Checkout and subscription management live in the clients, not on the server:**
- Web: `@revenuecat/purchases-js` with the RevenueCat **Web Billing public key** opens the embedded Paddle checkout. "Manage subscription" opens `CustomerInfo.managementURL` (Paddle customer portal).
- Flutter: `purchases_flutter` with the per-platform public keys purchases through StoreKit / Play Billing. The app only configures the SDK when `/api/v1/billing/status` reports `enabled: true`, so self-hosted instances make no RevenueCat or store calls.

The server therefore only exposes `GET /api/v1/billing/status` (public) and `GET /api/v1/subscription/status` (authenticated, force refresh).

**Localization:** Paddle Checkout, subscription emails and the Paddle customer portal are localized — the webclient passes its active UI language to the SDK and Paddle auto-detects the browser language as fallback. App Store/Play purchase sheets are localized by the operating system.

### Free accounts for the owner/testers

Keycloak user IDs listed in `TIMEASY_SUBSCRIPTION_ALLOWLIST` (comma-separated) bypass subscription checks without RevenueCat API calls. This does not create a subscription or management URL. Alternatively, set `subscription-allowlist <id1>,<id2>` in `config.txt` and restart the server. You can find a user's ID in the Keycloak admin console (Users → select user → copy ID) or in the `sub` claim of their access token.

### Local config file

Instead of exporting environment variables, create `server/src/config.txt` with your local settings and start the built server from `server/src`:

```bash
./bin/timeasy-server -config config.txt
```

Format: one `<flag-name> <value>` pair per line (space-separated, **no `=`**); lines starting with `#` are comments. Use flag names without the `TIMEASY_` prefix, e.g. `revenuecat-api-key sk_xxx`. Priority in the current `ff/v3` parser: command-line flags > non-empty `TIMEASY_*` environment variables > config file > defaults. Empty environment values are ignored; use an explicit command-line value to clear a non-empty default. `server/src/config.txt` is gitignored; restrict access and keep any secrets out of version control. Passing `-config config.txt` is needed to load it; Compose's example uses environment variables instead.

## Rate limiting

The server applies IP-based rate limiting to routes including the public billing status endpoint; CORS preflight requests are handled before the limiter. Each client IP gets its own token bucket; requests above the limit are answered with `429 Too Many Requests`, a `Retry-After` header and `X-RateLimit-Limit` / `X-RateLimit-Remaining` / `X-RateLimit-Reset` headers.

```bash
TIMEASY_RATE_LIMIT_ENABLED=true                       # default; set to false to disable entirely
TIMEASY_RATE_LIMIT_REQUESTS_PER_MINUTE=300            # sustained requests per client IP
TIMEASY_RATE_LIMIT_BURST=100                          # allowed short-term spike above that rate
TIMEASY_RATE_LIMIT_TRUSTED_PROXIES=0.0.0.0/0,::/0     # CIDRs allowed to set X-Forwarded-For
```

Limits apply per server instance and reset on restart. Clients behind the same NAT may share a bucket; monitor usage and adjust the limits for your traffic.

**Behind a reverse proxy** the server can only apply per-client limits if it trusts the proxy to set `X-Forwarded-For`. Narrow `TIMEASY_RATE_LIMIT_TRUSTED_PROXIES` to your proxy's IP/CIDR; otherwise every request appears to come from the proxy and all users share one bucket. Trusting all proxies (`0.0.0.0/0`, the default) keeps self-hosted setups working but allows clients to spoof their IP by sending the header themselves. The `client_ip` field of the HTTP request logs shows what the server actually sees — verify it after changing the setting.

`config.txt` alternative:

```
rate-limit-enabled true
rate-limit-requests-per-minute 300
rate-limit-burst 100
rate-limit-trusted-proxies 10.0.0.0/8
```
