import 'dart:io';

class Environment {
  static const String DEV_ENVIRONMENT = "dev";

  static const _ENVIRONMENT =
      String.fromEnvironment("ENVIRONMENT", defaultValue: DEV_ENVIRONMENT);

  static late final apiBaseUrl = _getApiBaseUrl();
  static late final authUrl = _getAuthUrl();

  /// Public website shown in the settings view. Store compliance: this URL
  /// must never point at a page with a purchase flow (Paddle checkout), since
  /// the app links to it (Apple guideline 3.1.1 / Google Play payments policy).
  /// Keep it to sign-in and information content — every purchase inside the
  /// app runs through the App Store / Google Play.
  static late final websiteUrl = _getWebsiteUrl();

  /// Privacy policy, linked from the paywall (required for auto-renewable
  /// subscriptions; linking to legal pages is always allowed).
  static const String privacyPolicyUrl = "https://timeasy.org/privacy";

  /// Numeric App Store ID of timeasy (the part after `id` in the store URL).
  /// Required by `in_app_review` to open the App Store review page; Android
  /// derives the listing from the package name.
  static const String appStoreId = "1620757668";

  /// Terms of use, linked from the paywall. There is no own terms page yet,
  /// so iOS falls back to Apple's standard EULA (accepted for auto-renewable
  /// subscriptions). Google Play does not require it — replace this with the
  /// own terms page for both platforms once it exists.
  static String? get termsOfUseUrl =>
      Platform.isIOS
          ? "https://www.apple.com/legal/internet-services/itunes/dev/stdeula/"
          : null;

  static String _getApiBaseUrl() {
    if (_ENVIRONMENT == DEV_ENVIRONMENT) {
      if (Platform.isAndroid) {
        return "http://10.0.2.2:8080/api/v1";
      } else {
        return "http://localhost:8080/api/v1";
      }
    }
    return _productionConfig["apiBaseUrl"];
  }

  static String _getAuthUrl() {
    if (_ENVIRONMENT == DEV_ENVIRONMENT) {
      if (Platform.isAndroid) {
        return "http://10.0.2.2:8180/realms/timeasy";
      } else {
        return "http://localhost:8180/realms/timeasy";
      }
    }
    return _productionConfig["authUrl"];
  }

  static String _getWebsiteUrl() {
    if (_ENVIRONMENT == DEV_ENVIRONMENT) {
      return "http://localhost:4200";
    }
    return _productionConfig["websiteUrl"];
  }

  static String getEnvironment() {
    return _ENVIRONMENT;
  }

  static const Map<String, dynamic> _productionConfig = {
    "apiBaseUrl": "https://api.timeasy.org/api/v1",
    "authUrl": "https://auth.timeasy.org/realms/timeasy",
    "websiteUrl": "https://timeasy.org"
  };
}
