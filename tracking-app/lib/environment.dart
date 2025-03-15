class Environment {
  static const String DEV_ENVIRONMENT = "dev";

  static const _ENVIRONMENT =
      String.fromEnvironment("environment", defaultValue: DEV_ENVIRONMENT);

  static late final apiBaseUrl = _config["apiBaseUrl"];
  static late final authUrl = _config["authUrl"];

  static late final Map<String, dynamic> _config =
      Environment._ENVIRONMENT == DEV_ENVIRONMENT
          ? _developmentConfig
          : _productionConfig;

  static const Map<String, dynamic> _developmentConfig = {
    "apiBaseUrl": "http://localhost:8080/api/v1",
    "authUrl": "http://localhost:8180/realms/timeasy"
  };

  static const Map<String, dynamic> _productionConfig = {
    "apiBaseUrl": "https://api.timeasy.org/api/v1",
    "authUrl": "https://auth.timeasy.org/realms/timeasy"
  };
}
