# timeasy

A time tracker application that ist build for simplicity.

<p>
    <a href="https://play.google.com/store/apps/details?id=com.hilwerssoftware.timeasy">
        <img alt="Get it on Google Play" src="assets-readme/google_play_logo.png" />
    </a>

    <a href="https://apps.apple.com/us/app/timeasy-time-tracking/id1620757668?itsct=apps_box_badge&amp;itscg=30200">
      <img src="assets-readme/appstore_logo.png" alt="Download on the App Store">
    </a>
</p>

![](assets-readme/screenshot_01.png)
![](assets-readme/screenshot_02.png)
![](assets-readme/screenshot_03.png)

## Getting Started

### Keystore

In order to compile for Android the app needs to be signed with an appropriate key. To create a keystore and reference it please refer to the flutter documentation: https://docs.flutter.dev/deployment/android

After generating the keystore create a file named "key.properties" in the "android"-folder:

    storePassword=<password for your keystore>
    keyPassword=<password for your key>
    keyAlias=upload
    storeFile=<location of the key store file, such as /home/<user name>/keystore.jks>

### Translations

Before you begin you must generate the translation files. Otherwise the project will not be compilable:

    flutter gen-l10n

### Application icon

For setting the application icon the package "flutter_launcher_icons"
(https://pub.dev/packages/flutter_launcher_icons) is used. The configuration can be found in the
pubspec.yaml.

Just place your icon in the path configured in the pubspec.yaml and run

flutter pub run flutter_launcher_icons:main

