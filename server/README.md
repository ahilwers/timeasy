# timeasy-server

This is the server component of timeasy. The server is not completed yet and isn't currently used at all.

## Development

### Authentication

The authentication is handled by Keycloak. To start Keycloak, just use the docker-compose.yaml in the "keycloak" directory. There's also a realm export file named realm-export.json in this directory that you can use to import the timeasy realm into Keycloak. Keep in mind that even after importing the realm you'll need to add users to your realm.

### Service-Testing

In the directory "service-testing" you'll find a script named "login.sh" that helps you authenticating with Keycloak. It uses [httpie](https://httpie.io/) to contact the Keycloak service and spawns a shell where the shell variable "TOKEN" is filled with the token Keycloak returned during login. You can then use this token for your REST calls:

```
http localhost:8080/api/v1/projects "Authorization:Bearer $TOKEN" < project.json
```

## Database-Migration

The project ist using golang-migrate (https://github.com/golang-migrate) for migrations. To install the CLI tool to create migrations follow these instructions: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate

## Billing (optional)

timeasy can run as a fully functional self-hosted instance without any billing configuration — by default, subscription checking is disabled and no endpoint returns 402.

To monetize a hosted instance, subscriptions are checked directly against [Polar](https://polar.sh):

```bash
TIMEASY_POLAR_ACCESS_TOKEN=<organization access token>   # setting this enables billing
TIMEASY_POLAR_PRODUCT_ID=<subscription product id>       # any active subscription for this product grants access
# TIMEASY_POLAR_ORGANIZATION_ID is optional — the organization is implied by the access token
TIMEASY_POLAR_ENV=sandbox                                # or "production"
TIMEASY_SUBSCRIPTION_CACHE_TTL_MINUTES=15                # how long subscription states are cached
TIMEASY_WEBCLIENT_URL=https://timeasy.org                # base URL for checkout success redirects
```

**Polar access token:** create an *Organization Access Token* with exactly these scopes — nothing else is needed:

| Scope | Used for |
|---|---|
| `checkouts:write` | create hosted checkout sessions (`POST /api/v1/upgrade`) |
| `customers:read` | look up the customer by external ID and read subscription state |
| `customers:write` | backfill `external_id` for purchases made outside the app funnel |
| `customer_sessions:write` | short-lived customer portal sessions ("manage subscription") |

Note: `customer_portal:write` is **not** required — the server creates a customer session and redirects to the portal URL it returns.

**Polar language:** The hosted **checkout** supports localization via a beta feature. Enable **“Localized Checkout”** under the organization's beta features in the Polar dashboard — without it the checkout stays English even when a valid `locale` is sent (verified against the sandbox). Once enabled, the webclient's active UI language is sent with the upgrade request (fallback: the browser's `Accept-Language` header; only valid BCP 47 tags are forwarded). Note: the checkout content is localized even though Polar may keep the page's `<html lang="en">` attribute.

The hosted **customer portal** is English-only regardless — it has no locale parameter and is rendered with `lang="en"`, independent of the customer's `locale` and the browser language.

How it works: the server links each Keycloak user to a Polar customer via `external_customer_id` when it creates the hosted checkout (`POST /api/v1/upgrade`). Subscription checks resolve the customer by that external ID and read the customer state (`active_subscriptions`). Results are cached in memory for the configured TTL; cancellations take effect within that window. Users who purchased outside the app are matched by email once and their external ID is backfilled.

### Local config file

Instead of exporting environment variables, you can put the settings into a local config file and start the server with it:

```bash
cp config.txt config.txt.local   # optional: keep a template
./bin/timeasy-server -config config.txt
```

Format: one `<flag-name> <value>` pair per line (space-separated, **no `=`**), `#` starts a comment. Use the flag names without the `TIMEASY_` prefix, e.g. `polar-access-token polar_oat_xxx`. Priority: command-line flags > config file > `TIMEASY_*` environment variables. `server/src/config.txt` is gitignored because it contains secrets.
