import 'dart:io';

class Environment {
  static const String DEV_ENVIRONMENT = "dev";

  static const _ENVIRONMENT =
      String.fromEnvironment("ENVIRONMENT", defaultValue: DEV_ENVIRONMENT);

  static late final apiBaseUrl = _getApiBaseUrl();
  static late final authUrl = _getAuthUrl();

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

  static String getEnvironment() {
    return _ENVIRONMENT;
  }

  static const Map<String, dynamic> _productionConfig = {
    "apiBaseUrl": "https://api.timeasy.org/api/v1",
    "authUrl": "https://auth.timeasy.org/realms/timeasy"
  };
}
