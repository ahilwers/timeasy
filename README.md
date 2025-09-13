# timeasy

A time tracker application that ist build for simplicity.

The software is still work in progress - only the standalone iOS app is usable at the moment.

Visit the timeasy website: <https://timeasy.org>

<p>
    <a href="https://apps.apple.com/us/app/timeasy-time-tracking/id1620757668?itsct=apps_box_badge&amp;itscg=30200">
      <img src="tracking-app/assets-readme/appstore_logo.png" alt="Download on the App Store"/>
    </a>
</p>
<!--
<p>
    <a href="https://play.google.com/store/apps/details?id=com.hilwerssoftware.timeasy">
        <img src="tracking-app/assets-readme/google_play_logo.png" alt="Get it on Google Play" />
    </a>
</p>
-->

# Repository structure

This repository currently consists of two applications. 

The simple time tracking app can be found in the directory "tracking-app". The app can be installed from the app store of your mobile platform with the links above.

The second app is the timeasy server that can be found in the directory "server". The server is not completed yet and it's currently not used by the app.

## CI/CD and Docker Images

This project uses GitHub Actions for continuous integration and deployment. Docker images are automatically built and pushed to Docker Hub for both the server and web client components.

### Automated Builds

**Server** (`ahilwers/timeasy-server`):
- Builds on changes to `server/**` directory
- Runs Go tests and builds the application
- Creates Docker images for production deployment

**Web Client** (`ahilwers/timeasy-webclient`):
- Builds on changes to `webclient/**` directory  
- Runs Angular tests, linting, and builds the application
- Creates Docker images for web deployment

### Docker Image Tagging

Images are automatically tagged based on the build context:

**Development builds** (on branch pushes):
- `develop`, `main` - Latest development versions
- `feature-branch-name` - Feature branch testing
- `branch-abc123` - Branch with commit hash for traceability

**Release builds** (on git tags):
- `1.2.3` - Exact version from git tag `v1.2.3`
- `1.2` - Major.minor version
- `latest` - Latest stable release (main/develop only)

### Creating a Release

To create a new release with versioned Docker images:

```bash
git tag v1.2.3
git push origin v1.2.3
```

This will trigger the CI pipeline to build and push Docker images with proper version tags to Docker Hub.
