# CI/CD and releases

This guide describes the current GitHub Actions workflows and the maintainer release process. See the [repository overview](README.md), [contribution guide](CONTRIBUTING.md) and [deployment guide](server/docker/README.md) for application setup and operation.

## Branching strategy

```text
feature/* → develop → main → vX.Y.Z
```

- Create feature branches from `develop` and submit pull requests back to `develop`.
- Use `develop` for integration and testing of upcoming changes.
- Promote reviewed and tested changes from `develop` to `main` through a pull request.
- Create release tags on the intended `main` commit after verifying the relevant workflows.

This is the release convention, not an enforcement implemented by the workflows: a pushed `v*` tag can publish images regardless of which branch contains its commit. Check the target commit carefully.

## Build triggers and checks

All three workflows support manual execution with `workflow_dispatch` from the GitHub Actions UI. None currently declares a `pull_request` trigger. Pushes to branches in this repository run the applicable workflows; a pull request from a fork does not itself start these workflows in the upstream repository.

| Workflow | Automatic trigger | Checks and outputs |
| --- | --- | --- |
| [Flutter CI-Build](.github/workflows/flutter-ci-build.yaml) | Pushes affecting `tracking-app/**` or the Flutter workflow; GitHub does not apply path filters to tag pushes | Dependency installation, localization generation, analysis, tests, release APK build and artifact upload. A separate iOS job builds without code signing and uploads an artifact only when the selected ref is `develop` or `main`. |
| [Server CI-Build](.github/workflows/server-ci-build.yaml) | Every push; no path or branch filter | Go build and tests; a separate Docker job publishes images only on eligible refs. |
| [WebApp CI-Build](.github/workflows/webapp-ci-build.yaml) | Every push; no path or branch filter | Dependency installation, lint, unit tests and production build; separate mocked desktop/mobile Playwright tests and Docker publishing. |

The server and web workflows therefore run even for changes outside their component directories, including documentation-only pushes. Feature branches are built and tested but do not publish Docker images.

### Job dependencies and limitations

- Server Docker publishing requires the server build/test job to succeed.
- Web E2E tests and Docker publishing both depend on the web build job, but **publishing does not wait for E2E tests**. Check the complete workflow result before approving a release, not just image availability.
- Web lint currently uses `continue-on-error: true`; lint failures do not block publishing. Contributors should still fix lint failures.
- The Flutter iOS job is independent of the Android/test job. Its artifact is an unsigned build check, not an App Store submission. The Android APK is also only uploaded as a workflow artifact; these workflows do not publish either mobile app to a store.
- Workflows do not automatically deploy the images to a running installation or create a GitHub Release.

## Docker images and tags

The images are published to Docker Hub:

- [`ahilwers/timeasy-server`](https://hub.docker.com/r/ahilwers/timeasy-server)
- [`ahilwers/timeasy-webclient`](https://hub.docker.com/r/ahilwers/timeasy-webclient)

Both publishing jobs allow only `refs/heads/develop`, `refs/heads/main` and refs starting with `refs/tags/v`. This also applies to manually dispatched runs. Publishing needs the repository secrets `DOCKER_USERNAME` and `DOCKER_PASSWORD` with access to the target image repositories; fork contributors do not need these secrets for local development.

The workflows use [`docker/metadata-action`](https://github.com/docker/metadata-action) to generate tags. For either image, the meaningful version and branch tags are:

| Git ref | Image tags |
| --- | --- |
| Branch `develop` | `develop`, `develop-<short-sha>` |
| Branch `main` | `main`, `main-<short-sha>` |
| Stable tag `v1.2.3` | `1.2.3`, `1.2`, `latest` |
| Prerelease tag `v1.2.3-rc.1` | `1.2.3-rc.1`; no stable `1.2` or automatic `latest` tag |
| Other branch, such as `feature/example` | No image publication |

The short SHA is seven characters by default. The SHA rule also runs for tag builds; use the semantic-version tags for releases. The workflows do not configure a major-only alias such as `1`, and the image's semantic-version tag omits the Git tag's leading `v`.

**`latest` has two sources in the current configuration:** stable semantic-version tags generate it automatically, and a branch run also explicitly generates it when that branch is the repository's configured default branch (provided it is eligible for publishing). Consequently, `latest` is **not a reliable indicator of the latest tested stable release**. A later default-branch build or an older-version release can move it. The `1.2` alias can likewise move when another release in that series is built; there is no version-order guard.

For deployments, select and verify an exact version or image digest rather than relying on `latest`, `main`, `develop` or a major/minor alias. Tags can be overwritten by repeated publication; use a digest when immutability is required. Self-hosters must still rebuild the web image for their own API and authentication URLs as described in the [deployment guide](server/docker/README.md).

## Creating a release

This procedure is for maintainers with permission to push tags and publish images. It uses one repository tag for the server and web client builds; it does not update version files or release the mobile apps automatically.

1. Merge the intended changes into `main` through the normal review process. Run the applicable checks in [CONTRIBUTING.md](CONTRIBUTING.md) and wait for the complete workflows on the intended commit, including the web E2E job.
2. Review compatibility, database migrations and release notes. Test upgrades and recovery where relevant. Update any component version files needed for the release in reviewed commits; creating a Git tag does not edit them. Flutter uses `tracking-app/pubspec.yaml` for its application version/build number.
3. Choose a new, unused semantic version (`MAJOR.MINOR.PATCH`) and verify the clean checkout and target commit. The following commands are an example; replace `1.2.3` with the intended version:

   ```bash
   git switch main
   git pull --ff-only origin main
   git status --short
   git log -1 --oneline
   git tag --list v1.2.3
   ```

   Stop if the working tree is not clean, the commit is not the intended release, or the tag already exists. The current `main` workflow results must correspond to this commit.

4. Create an annotated tag and push **that tag only**:

   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git show v1.2.3 --no-patch
   git push origin v1.2.3
   ```

   Pushing the tag starts the server and web workflows and their eligible image-publishing jobs. Flutter's push workflow can also run on the tag, but its iOS job is skipped because the ref is a tag, not `main` or `develop`. Use valid semantic versions, such as `v1.2.3` or `v1.2.3-rc.1`; the broad `v*` publishing condition does not itself validate the version.

5. Inspect the **tag-triggered** workflows in [GitHub Actions](https://github.com/ahilwers/timeasy/actions). Confirm that tests, web E2E checks and both image publications succeeded. The images are published by separate workflows, not atomically; one image can become available before the other or before a failure elsewhere.
6. Verify the intended versions in Docker Hub. For example:

   ```bash
   docker pull ahilwers/timeasy-server:1.2.3
   docker pull ahilwers/timeasy-webclient:1.2.3
   ```

   Review the image metadata and perform deployment smoke tests, including authentication and data access. Pulling an image alone is not a deployment test.

7. Create a GitHub Release for the existing tag and write release notes covering changes, compatibility, migrations and upgrade instructions. This step is manual. Deploy the chosen images separately using the [self-hosting update procedure](server/docker/README.md#5-backups-and-updates) or your managed deployment process.

For prerelease verification, use a new tag such as `v1.2.3-rc.1` and check its full workflow results before creating the final version. Do not move or reuse a published release tag. If a run fails, investigate before rerunning; a rerun can rebuild and overwrite image tags. When fixing the release commit itself, use a new version/tag and explain the correction in the release notes.
