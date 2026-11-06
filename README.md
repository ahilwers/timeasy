# timeasy

timeasy is a time tracker designed for simplicity, particularly for freelance software developers. Track time against projects, review your work and export entries without making time tracking the focus of your day.

Choose between **running timeasy yourself** and a **hosted timeasy subscription**. To use the hosted service rather than maintain your own installation, see [timeasy.org](https://timeasy.org) for the offer and current availability.

> **Project status:** The server and web client are under development. The updated, connected mobile app has not yet been released to the app stores; the older store version is not compatible with this setup. Features, deployment instructions and billing integrations may change.

## What you can do

- Start and stop time entries and organize them by project.
- Review recorded time, including weekly statistics, and export entries as CSV or Excel files.
- Connect timeasy projects to GitHub repositories, GitLab projects or Jira projects, then track time against individual issues within them.
- Use the Flutter mobile app with local storage and synchronization when connected to a server.
- Run your own server and web client, or use the hosted service instead.

Availability and behavior can differ between clients and deployments. Issue integrations associate timeasy entries with issues; do not assume they create work logs in the external service.

## Getting started

### Hosted timeasy

If you prefer not to operate the infrastructure yourself, visit [timeasy.org](https://timeasy.org) for information about the hosted subscription, features and availability. The hosted service is separate from the software you can run from this repository.

### Self-hosting

The server uses PostgreSQL for application data and Keycloak for authentication. The web client connects to both the API and Keycloak. You will need Docker and Docker Compose, a Keycloak installation, persistent storage and appropriate network, domain and TLS configuration.

1. Read the [self-hosting and deployment guide](server/docker/README.md).
2. Use the [example Docker Compose file](server/docker/docker-compose.yaml) as a starting point, adapting the images, credentials and URLs for your environment. **It is an example, not a ready-to-use production installation:** it references local image names, does not include Keycloak, exposes ports and contains development credentials.
3. Configure Keycloak, the server and the web client consistently. The web client's API and authentication URLs are compiled into its image and default to timeasy.org; rebuild it with your own URLs. Start the stack and verify login and API access as described in the guide.
4. Before exposing the installation, secure it and set up backups and updates. Billing is optional: without a RevenueCat secret API key, the self-hosted server does not enforce subscriptions.

The mobile app can also be built from source; see the [mobile documentation](tracking-app/README.md). For security and data-loss considerations, follow the deployment guide rather than exposing the example configuration as-is.

## Repository and documentation

| Component | Description | Documentation |
| --- | --- | --- |
| `server/src/` | Go REST API, PostgreSQL persistence and synchronization | [Server README](server/README.md) |
| `server/docker/` | Example server, database and web client deployment | [Self-hosting guide](server/docker/README.md) · [Compose example](server/docker/docker-compose.yaml) |
| `server/keycloak/` | Realm export and local Keycloak setup | [Development Keycloak README](server/keycloak/development/README.md) |
| `server/service-testing/` | Manual API testing examples | [API testing README](server/service-testing/README.md) |
| `webclient/timeasy-webclient/` | Angular web client | [Web client README](webclient/timeasy-webclient/README.md) |
| `tracking-app/` | Flutter mobile app | [Mobile app README](tracking-app/README.md) |

## Contributing

Contributions, bug reports and suggestions are welcome. Start with the [contribution guide](CONTRIBUTING.md) for local setup, coding and testing guidelines, and the pull request workflow. For an issue, include the affected component, steps to reproduce and relevant logs with credentials and personal information removed.

The main development tools are Go, Node.js/npm, Flutter and Docker; which ones you need depends on the component. The component READMEs describe the required versions and commands. Server integration tests use disposable PostgreSQL containers and require a working Docker daemon.

## CI and container images

The development and release flow is `feature/*` → `develop` → `main` → `vX.Y.Z`. GitHub Actions builds and tests the components using the workflows in [`.github/workflows/`](.github/workflows/).

- **Server and web client:** build on every push, without path filters. Docker publishing is restricted to `develop`, `main` and `v*` tags; feature branches do not publish images.
- **Flutter:** push builds are path-filtered for app/workflow changes (path filters do not apply to tag pushes). CI generates localizations, analyzes, tests and uploads an APK; unsigned iOS builds run only on `develop` and `main`. There is no automatic store submission.
- **Manual builds:** all three workflows support `workflow_dispatch`. They do not currently run on `pull_request` events.

Published images are [`ahilwers/timeasy-server`](https://hub.docker.com/r/ahilwers/timeasy-server) and [`ahilwers/timeasy-webclient`](https://hub.docker.com/r/ahilwers/timeasy-webclient). Branch builds publish branch and short-SHA tags; a stable Git tag such as `v1.2.3` publishes `1.2.3`, `1.2` and `latest`. Default-branch builds can also update `latest`, so use a verified exact version or digest for deployments.

### Creating a release

After merging and verifying the intended release commit on `main`, maintainers create and push a new version tag, for example:

```bash
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

This triggers the server and web image builds, not an automatic deployment or GitHub Release. Check the complete tag workflows before announcing a release: web image publishing does not wait for the separate E2E job.

See **[CI/CD and releases](cicd.md)** for the full release checklist, trigger strategy, job dependencies, image tagging rules, prerequisites and verification steps.

## License

The repository is licensed under the [GNU Affero General Public License, version 3](LICENSE). Review the license when modifying or operating the software, including its terms for network use of modified versions.
