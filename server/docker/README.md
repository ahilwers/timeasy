# Self-hosting the server and web client

The checked-in [`docker-compose.yaml`](docker-compose.yaml) is a **starting example, not a ready-to-run or production-secure deployment**. It does not include Keycloak or a TLS reverse proxy. It uses local image names, hardcoded database credentials, an unpinned PostgreSQL image, and `https://auth.timeasy.org` as the Keycloak host. It publishes PostgreSQL, the API and web client on host ports. Review and adapt it before starting anything; never use those defaults on a public host.

**Web client blocker:** the production web build currently embeds `https://api.timeasy.org/api/v1` and `https://auth.timeasy.org`. The Dockerfile does not substitute these at runtime. A self-hoster must change the URLs and rebuild the web image as described below; changing the server or Compose environment alone does not redirect the browser. Only the optional public RevenueCat billing key is supplied by the server at runtime.

The server and web client are works in progress. Plan to test and maintain your own deployment; the example has no automatic Keycloak setup, TLS, health checks, backups or database upgrade procedure. The complete setup and login sequence below has not been verified end to end.

## 1. Plan the deployment

You need Docker with Compose, PostgreSQL storage, a separately deployed Keycloak realm, DNS and a TLS reverse proxy. Choose browser-reachable HTTPS URLs for the web client, API and Keycloak. The proxy must forward `/api/v1/...` to the server **without stripping that path** and forward the web URL to the web client. The example web container serves static files; its nginx configuration does not proxy the API or Keycloak. Configure Keycloak's external hostname and proxy settings for your installation as well.

Set up Keycloak separately. The API fetches signing keys from `{TIMEASY_KEYCLOAK_HOST}/realms/{TIMEASY_KEYCLOAK_REALM}/protocol/openid-connect/certs`, so the host must also be reachable from the server container. Its default realm is `timeasy`; the API expects the token's `sub` to be a UUID. The [realm export](../keycloak/realm-export.json) is a starting point, not a complete production configuration: import/review it, create users, replace any credentials, and review registration and security policies. In particular, the export enables registration and disables brute-force protection. The [development Keycloak Compose](../keycloak/development/docker-compose.yaml) has `admin`/`admin` and other development credentials; do not expose it as a production setup.

For the web client, configure the Keycloak public client `timeasy-webclient` with your actual web URL in its valid redirect URIs, post-logout redirect URIs and web origins; the export contains localhost entries. For the mobile client, configure the public client `timeasy-tracking-app` with `org.timeasy.app://callback` in **both** valid redirect URIs and valid post-logout redirect URIs. The checked-in export instead has `http://localhost:4000/`, while the current app uses `org.timeasy.app://callback` for login and logout. Review and remove unwanted development entries.

The API's token verifier uses the configured realm's signing keys but does not explicitly enforce issuer or audience claims. Treat the realm and its clients as a security boundary, restrict who can obtain tokens, and assess this limitation before exposing the service. The API also responds with a wildcard CORS origin; plan your exposure accordingly.

## 2. Build a web image for your URLs

In a working copy, edit [`webclient/timeasy-webclient/src/environments/environment.ts`](../../webclient/timeasy-webclient/src/environments/environment.ts). For example, replace the placeholders with your actual HTTPS hosts:

```ts
export const environment = {
  production: true,
  apiUrl: 'https://api.example.com/api/v1',
  authUrl: 'https://auth.example.com',
};
```

These URLs are used by browsers, so do **not** use Docker-internal names such as `timeasy-server` or `localhost` for a remote deployment. `authUrl` is the Keycloak base URL, without `/realms/timeasy`; `apiUrl` includes `/api/v1`. The web client currently fixes the realm and client ID to `timeasy` and `timeasy-webclient` in [`app.config.ts`](../../webclient/timeasy-webclient/src/app/app.config.ts). Align Keycloak with these values or change the client code too. The production build uses `environment.ts`, not `environment.development.ts`.

From the repository root, build local images after making the web changes:

```bash
docker build -t timeasy-server:local server/src
docker build -t timeasy-webclient:local webclient/timeasy-webclient
```

Use these tags in your Compose copy. Alternatively, use a suitable published server image by explicit tag; the example's `timeasy-server` and `timeasy-webclient` names do not reference the published `ahilwers/` images. A stock published web image still contains the URLs it was built with. Rebuild your web image after URL changes.

The Flutter app has separate built-in API and Keycloak URLs in [`tracking-app/lib/environment.dart`](../../tracking-app/lib/environment.dart); changing the web image or server does not reconfigure an existing mobile app. A mobile deployment targeting your instance requires appropriate app configuration and a rebuild, plus the Keycloak redirect changes above.

## 3. Prepare a private Compose copy

From the repository root, copy the example outside the checkout so deployment secrets are not accidentally committed:

```bash
mkdir -p "$HOME/timeasy-deploy"
chmod 700 "$HOME/timeasy-deploy"
cp server/docker/docker-compose.yaml "$HOME/timeasy-deploy/compose.yaml"
chmod 600 "$HOME/timeasy-deploy/compose.yaml"
```

Edit that copy **before** running it:

1. Change the server and web image names to the images built above (or your chosen explicit server image tag). Pin a PostgreSQL image/version compatible with your data rather than using unqualified `postgres`. For a new installation, the repository's separate development database example uses `postgres:17.6`; verify your chosen image's data directory and upgrade process. For an existing volume, do not change PostgreSQL major versions without a tested migration and backup.
2. Replace the database user/password in both `POSTGRES_USER` / `POSTGRES_PASSWORD` and `TIMEASY_DATABASE_USER` / `TIMEASY_DATABASE_PASSWORD`; ensure `POSTGRES_DB` and `TIMEASY_DATABASE_NAME` match. Remove the database's `ports` mapping so it is not published to the host. PostgreSQL initialization variables only initialize a **new** data directory; editing them does not change existing database credentials. Keep secrets outside version control and restrict access to the Compose file and Docker daemon.
3. Set `TIMEASY_KEYCLOAK_HOST` to your own reachable Keycloak base URL, without a trailing slash, and set `TIMEASY_KEYCLOAK_REALM` if not `timeasy`. For the web client as currently built, keep the realm `timeasy` unless you also change its source.
4. Put the API and web client behind your TLS proxy. For a proxy running on the same host, replace their port mappings with `127.0.0.1:8080:8080` and `127.0.0.1:8081:80` respectively and proxy to those loopback ports. If the proxy runs in a container or on another host, adapt the networking and firewall instead; its route must reach the services without making the raw HTTP ports public.
5. Set `TIMEASY_RATE_LIMIT_TRUSTED_PROXIES` to only the actual reverse proxy IPs/CIDRs as seen by the API, and ensure the proxy strips untrusted incoming forwarding headers before setting its own. With no proxy, disable trust with an explicit empty command-line flag: add `command: ["./timeasy-server", "-rate-limit-trusted-proxies="]` under the server service. An empty environment variable is ignored by the current parser and leaves the trust-all default in place. The default allows spoofing of the IP used for rate limiting. See [rate limiting](../README.md#rate-limiting).
6. Leave `TIMEASY_REVENUECAT_API_KEY` unset for billing-free operation, or configure all required RevenueCat settings as described in [Billing](../README.md#billing-optional). Do not put a secret RevenueCat key in the web source or image.

The example uses fixed container names and a globally named `timeasy-data` volume, so another deployment can collide with or reuse them. Change these names for separate installations. The server connects to PostgreSQL with `sslmode=disable` and has no database TLS setting: keep that connection on a protected local/private network, or provide an appropriate secure tunnel. The server stores external integration tokens in the database without application-level encryption; restrict database and backup access and use encrypted storage/backups.

## 4. Start and verify

After configuring the copy, Keycloak, proxy and images, run:

```bash
docker compose -f "$HOME/timeasy-deploy/compose.yaml" config --quiet
docker compose -f "$HOME/timeasy-deploy/compose.yaml" up -d
docker compose -f "$HOME/timeasy-deploy/compose.yaml" ps
docker compose -f "$HOME/timeasy-deploy/compose.yaml" logs --tail=100 timeasy-server
```

`config --quiet` checks Compose syntax, not security, image availability or service connectivity. `depends_on` does not wait for PostgreSQL readiness. The API pings PostgreSQL and runs migrations at startup; if startup fails, check the database and server logs, then restart the server when PostgreSQL is ready. Back up before upgrades; migrations run automatically and the database account must have the necessary schema permissions.

If using the loopback bindings above, check the public billing status and that an unauthenticated API request is rejected:

```bash
curl -fsS http://127.0.0.1:8080/api/v1/billing/status
curl -i http://127.0.0.1:8080/api/v1/projects
```

With billing disabled the first response is `{"enabled":false}`; the second should return 401. These checks do not prove that login works. Visit your HTTPS web URL, sign in with a Keycloak user and verify an authenticated API request in the browser. Also test the externally exposed HTTPS endpoints and mobile login/logout if used. Never share tokens, credentials or unredacted Compose output when troubleshooting.

## 5. Backups and updates

Back up both the application database and Keycloak's database/configuration. Preserve your deployment configuration securely as well. Keycloak user IDs identify the owners of timeasy data; recreating users instead of restoring them can break access to that data. Backups include personal data and may include integration tokens, so encrypt them, control access and keep a separate copy outside the Docker host.

For example, create an application database dump with a unique filename:

```bash
umask 077
mkdir -p "$HOME/timeasy-deploy/backups"
docker compose -f "$HOME/timeasy-deploy/compose.yaml" exec -T timeasy-database \
  sh -c 'exec pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --format=custom' \
  > "$HOME/timeasy-deploy/backups/timeasy-$(date -u +%Y%m%dT%H%M%SZ).dump"
```

Check that the command succeeded; a failed command can leave an empty or incomplete file. This covers only the application database, not your separately operated Keycloak installation. Follow your Keycloak deployment's database backup instructions too. A Docker volume alone is not a backup, and copying a live PostgreSQL data directory is not a substitute for a consistent database backup.

Test restoration regularly into an isolated, empty PostgreSQL database with `pg_restore`, using a compatible PostgreSQL version and the corresponding Keycloak backup. Verify login, project ownership and time entries before relying on the procedure. Do not test restore commands against your live database.

For updates:

1. Review changes, take verified backups and test the update in a separate environment.
2. Pull the chosen published image versions or rebuild your custom images, retaining your API/authentication URL changes. Keep the previous images and deployment configuration for recovery.
3. Update the explicit tags in your Compose copy and run `docker compose -f "$HOME/timeasy-deploy/compose.yaml" up -d`.
4. Check logs, HTTPS access, login and data. Migrations run automatically; reverting an image does not undo a database migration. Recovery may require restoring a compatible backup.

Keep PostgreSQL major-version upgrades separate from routine image updates and follow PostgreSQL's upgrade procedure. Monitor logs, available disk space, backup jobs and certificate expiry. Avoid `docker compose down -v` unless you explicitly intend to delete the Compose-managed data volume.
