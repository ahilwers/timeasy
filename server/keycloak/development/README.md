# Local Keycloak for development

This Compose file runs a local Keycloak instance and a separate PostgreSQL database for authentication. It is intended for development only; it uses publicly documented passwords and unpinned images. Do not expose it to the internet or use it as a production configuration.

## Start and configure

From the repository root, with Docker and Docker Compose installed:

```bash
docker compose -f server/keycloak/development/docker-compose.yaml up -d
```

1. Open <http://localhost:8180> and sign in to the administration console with `admin` / `admin`.
2. Create a realm by importing [`server/keycloak/realm-export.json`](../realm-export.json). The Compose file does **not** import it automatically.
3. Create a test user in the `timeasy` realm and set a password. Ensure the user can log in and complete any required actions.
4. Check the realm's clients, redirect URIs and allowed web origins for your local client URLs. The mobile client uses `org.timeasy.app://callback` for login and logout; add it to the `timeasy-tracking-app` client's valid redirect and post-logout redirect URIs, since the export still contains an older localhost URL. For a confidential client, configure an appropriate client secret; the masked secret in the export is not a usable credential. Never put real secrets in the repository.
5. Configure the server and clients to use the same Keycloak URL and realm. `localhost` refers to the device or container making the request, so another device or container may need a different reachable address.

The export contains development-oriented settings, including permissive origins and redirect URIs. Review and restrict these, authentication flows, registration, credentials and user-management settings before considering a deployment. See the [server documentation](../../README.md) and [deployment guide](../../docker/README.md).

Check logs or stop the services from the repository root:

```bash
docker compose -f server/keycloak/development/docker-compose.yaml logs -f keycloak
docker compose -f server/keycloak/development/docker-compose.yaml down
```

`down` leaves the database volume in place. Removing volumes destroys Keycloak users and configuration; back up anything you need first. The Compose file uses `bitnami/keycloak:latest`; image availability and behavior may change, so choose and test a supported, pinned image for your own deployment.

## Theme

The Compose file mounts [`themes/`](themes/) into Keycloak. It includes the [Tailcloakify](https://github.com/ALMiG-Kompressoren-GmbH/tailcloakify) login theme. Check theme compatibility when changing the Keycloak image version.
