# timeasy mobile app

The Flutter mobile client tracks projects and time entries locally and can synchronize them with a timeasy server. Local tracking works without signing in; synchronization requires a reachable server and a Keycloak account. The updated, server-connected app is not yet released to the App Store or Google Play. The server and web client are still under development. See the [repository overview](../README.md), [server instructions](../server/README.md), and [contribution guide](../CONTRIBUTING.md).

## Prerequisites

- Flutter with Dart compatible with `pubspec.yaml` (Dart `^3.13.0`); CI uses Flutter 3.47.x.
- For Android, the Android SDK and Java 17. For iOS, macOS and Xcode; the Xcode project targets iOS 15 or later. See the [Flutter setup guide](https://docs.flutter.dev/get-started/install) for platform-specific tools and device setup.
- A simulator, emulator, or physical device. A Go server, PostgreSQL, and Keycloak are needed for sign-in and synchronization, but not for local-only tracking.

Run all commands below from `tracking-app/`:

```bash
flutter pub get
flutter gen-l10n
flutter devices
flutter run
```

Select a device with `flutter run -d <device-id>` if needed. Generate localizations before the first build, run, or test and whenever translations change.

## Server connectivity and configuration

`lib/environment.dart` selects endpoints at **build time** using `--dart-define=ENVIRONMENT=...`. The default is `dev`, including for release builds:

| Environment | API | Keycloak realm |
| --- | --- | --- |
| `dev`, Android emulator | `http://10.0.2.2:8080/api/v1` | `http://10.0.2.2:8180/realms/timeasy` |
| `dev`, other platforms | `http://localhost:8080/api/v1` | `http://localhost:8180/realms/timeasy` |
| Any value other than `dev` | `https://api.timeasy.org/api/v1` | `https://auth.timeasy.org/realms/timeasy` |

For example, `flutter run --dart-define=ENVIRONMENT=production` selects the hosted endpoints. These URLs are not independently configurable via Dart defines; a different server requires changing `lib/environment.dart` and rebuilding. `localhost` on a physical device refers to that device, and `10.0.2.2` is an Android emulator host alias, not a physical-device address. Use addresses reachable by the device and configure Keycloak's issuer, HTTP/TLS settings, and redirect URIs accordingly.

For a local backend, follow the [server setup](../server/README.md): start PostgreSQL and Keycloak, import `server/keycloak/realm-export.json`, create a user, and run the API. The mobile app uses Keycloak client `timeasy-tracking-app` and redirect URI `org.timeasy.app://callback`. Check the imported client's allowed redirect URIs: the checked-in realm export does not include this callback by default. The `dev` website link is `http://localhost:4200`, which also needs adjustment for access from an Android emulator or physical device.

Changes are stored in SQLite and queued for synchronization. When authenticated, the app attempts synchronization on relevant events and periodically while running. Check the sync status in Settings before signing out or switching accounts: **sign-out and account changes clear locally cached projects, time entries, and pending changes**, so unsynchronized work can be lost.

## Analyze, test, and build

```bash
flutter analyze
flutter test
flutter build apk
flutter build appbundle
flutter build ios --release --no-codesign
```

Use the platform's normal signing and provisioning process for distribution; a no-codesign iOS build is only a build check. Android debug builds need no release keystore. For a distributable Android release, follow [Flutter's Android deployment instructions](https://docs.flutter.dev/deployment/android) and create the ignored `android/key.properties` with `storeFile`, `storePassword`, `keyAlias`, and `keyPassword`. An absolute `storeFile` path avoids ambiguity. Keep the keystore and passwords out of the repository. If this file is absent, the Gradle release configuration falls back to the **debug signing key**; that output is not suitable for a store release.

For an iOS archive, configure a signing team and provisioning in Xcode and use `flutter build ipa`. Add the appropriate `--dart-define` options to every run or build that needs non-default configuration.

## Localization and assets

Translations are in `lib/l10n/app_en.arb` and `lib/l10n/app_de.arb`; English is the template in `l10n.yaml`. Update both languages as appropriate, then run `flutter gen-l10n`. Generated localization code is used by the app.

Launcher icon sources and settings are in `pubspec.yaml`. After updating the source assets, regenerate the platform icons with:

```bash
dart run flutter_launcher_icons
```

## Billing and logging

Billing is optional and controlled by the server. When its public `/api/v1/billing/status` reports billing disabled, the mobile app does not configure RevenueCat. To test native subscriptions, enable billing on the server, configure the RevenueCat project, products and offering, and provide the corresponding **public** platform SDK key at build time:

```bash
flutter run --dart-define='REVENUECAT_IOS_API_KEY=<public-ios-key>'
flutter run --dart-define='REVENUECAT_ANDROID_API_KEY=<public-android-key>'
```

Use the matching define for `flutter build` as well. Missing keys do not prevent a build, but native purchase offerings are unavailable. The app expects the RevenueCat entitlement lookup key `timeasy_pro`; the server is configured with the corresponding internal entitlement ID. The Keycloak user ID is the RevenueCat App User ID across clients. See [server billing setup](../server/README.md#billing-optional).

In-app purchases and restore use StoreKit or Google Play Billing; the app must not link to external purchase or web subscription-management flows. Test actual transactions with the appropriate store sandbox/test accounts, configured products, and eligible test builds; Play Billing may require distribution through a Play testing track. A RevenueCat Test Store is a separate simulated option. A web purchase can unlock the same account when the shared entitlement is configured, but the mobile app does not link to its web management portal.

Optional Loki logging uses the compile-time defines `LOKI_ENDPOINT` and `LOKI_BEARER_TOKEN`; without an endpoint, logs remain local. **Dart defines are embedded in the app and are not secret storage.** Never ship sensitive tokens or server-side RevenueCat secret keys in a client build, and do not commit credentials or signing material.
