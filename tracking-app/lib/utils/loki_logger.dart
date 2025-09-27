import 'dart:convert';
import 'dart:io';

import 'package:logger/logger.dart' as pretty_logger;
import 'package:logging/logging.dart';
import 'package:package_info_plus/package_info_plus.dart';
import 'package:timeasy/environment.dart';

class LokiLogger {
  static LokiLogger? _instance;
  late Logger _lokiLogger;
  late pretty_logger.Logger _fallbackLogger;
  static String? _globalClientId;

  // Compile-time configuration
  static const String _lokiEndpoint = String.fromEnvironment(
    'LOKI_ENDPOINT',
    defaultValue: '',
  );
  static const String _lokiBearerToken = String.fromEnvironment(
    'LOKI_BEARER_TOKEN',
    defaultValue: '',
  );

  static const String _applicationName = 'timeasy-tracking-app';
  static const String _serviceName = 'timeasy';
  static String _serviceVersion = '1.0.0';

  LokiLogger._() {
    _initializeLokiLogger();
    _initializeFallbackLogger();
  }

  static LokiLogger get instance {
    _instance ??= LokiLogger._();
    return _instance!;
  }

  // Call this after Flutter bindings are initialized
  static Future<void> initializeVersion() async {
    try {
      final packageInfo = await PackageInfo.fromPlatform();
      _serviceVersion = '${packageInfo.version}+${packageInfo.buildNumber}';
      print('[LokiLogger] Version updated to: $_serviceVersion');
    } catch (e) {
      print('[LokiLogger] Failed to get package info: $e');
    }
  }

  // Set the global client ID to be included in all logs
  static void setClientId(String? clientId) {
    _globalClientId = clientId;
    print('[LokiLogger] Client ID updated to: $clientId');
  }

  void _initializeLokiLogger() {
    try {
      Logger.root.level = Level.ALL;

      _lokiLogger = Logger(_serviceName);

      if (_lokiEndpoint.isNotEmpty) {
        print('[LokiLogger] Loki logging enabled');
      } else {
        print('[LokiLogger] Loki logging disabled - no endpoint');
      }
    } catch (e) {
      print('[LokiLogger] Failed to initialize Loki logger: $e');
      _lokiLogger = Logger(_serviceName); // Fallback to basic logger
    }
  }

  void _initializeFallbackLogger() {
    _fallbackLogger = pretty_logger.Logger(
      printer: pretty_logger.PrettyPrinter(
        methodCount: 0,
        errorMethodCount: 5,
        lineLength: 120,
        colors: true,
        printEmojis: false,
        dateTimeFormat: pretty_logger.DateTimeFormat.onlyTimeAndSinceStart,
      ),
    );
  }

  static void d(String message,
      {dynamic error, StackTrace? stackTrace, String method = ''}) {
    instance._logDebug(message, error, stackTrace, method);
  }

  static void i(String message,
      {dynamic error, StackTrace? stackTrace, String method = ''}) {
    instance._logInfo(message, error, stackTrace, method);
  }

  static void w(String message,
      {dynamic error, StackTrace? stackTrace, String method = ''}) {
    instance._logWarn(message, error, stackTrace, method);
  }

  static void e(String message,
      {dynamic error, StackTrace? stackTrace, String method = ''}) {
    instance._logError(message, error, stackTrace, method);
  }

  void _logDebug(
      String message, dynamic error, StackTrace? stackTrace, String method) {
    _logWithLevel(Level.FINE, message, error, stackTrace, method);
  }

  void _logInfo(
      String message, dynamic error, StackTrace? stackTrace, String method) {
    _logWithLevel(Level.INFO, message, error, stackTrace, method);
  }

  void _logWarn(
      String message, dynamic error, StackTrace? stackTrace, String method) {
    _logWithLevel(Level.WARNING, message, error, stackTrace, method);
  }

  void _logError(
      String message, dynamic error, StackTrace? stackTrace, String method) {
    _logWithLevel(Level.SEVERE, message, error, stackTrace, method);
  }

  void _logWithLevel(Level level, String message, dynamic error,
      StackTrace? stackTrace, String method) {
    var logMessage = message;

    if (error != null) {
      logMessage += ' | Error: $error';
    }
    if (stackTrace != null) {
      logMessage += ' | StackTrace: $stackTrace';
    }

    if (_lokiEndpoint.isNotEmpty) {
      _sendToLokiDirect(level, logMessage, error, stackTrace, method);
    }

    try {
      // Send via logging framework (for console output)
      _lokiLogger.log(level, logMessage);
    } catch (e) {
      print('[LokiLogger] Loki logging failed: $e');
    }

    // Always log locally as fallback using pretty logger
    _logLocally(level, message, error, stackTrace, method);
  }

  void _logLocally(Level level, String message, dynamic error,
      StackTrace? stackTrace, String method) {
    if (method.isNotEmpty) {
      message = '[$method] $message';
    }
    if (level == Level.FINE) {
      _fallbackLogger.d(message, error: error, stackTrace: stackTrace);
    } else if (level == Level.INFO) {
      _fallbackLogger.i(message, error: error, stackTrace: stackTrace);
    } else if (level == Level.WARNING) {
      _fallbackLogger.w(message, error: error, stackTrace: stackTrace);
    } else if (level == Level.SEVERE) {
      _fallbackLogger.e(message, error: error, stackTrace: stackTrace);
    }
  }

  void _sendToLokiDirect(Level logLevel, String message, dynamic error,
      StackTrace? stackTrace, String method) {
    if (_lokiEndpoint.isEmpty || _lokiBearerToken.isEmpty) {
      return;
    }

    String level = 'info';
    if (logLevel == Level.FINE)
      level = 'debug';
    else if (logLevel == Level.INFO)
      level = 'info';
    else if (logLevel == Level.WARNING)
      level = 'warning';
    else if (logLevel == Level.SEVERE) level = 'error';

    final payload = {
      "streams": [
        {
          "stream": {
            "application": _applicationName,
            "service": _serviceName,
            "env": Environment.getEnvironment(),
            "level": level,
            "version": _serviceVersion,
            if (method.isNotEmpty) "method": method,
            if (_globalClientId != null) "client_id": _globalClientId!,
          },
          "values": [
            [
              "${DateTime.now().microsecondsSinceEpoch * 1000}", // Nanosecond timestamp
              jsonEncode({
                "client_id": _globalClientId ?? "not_set",
                "msg": message,
                "logger": _serviceName,
                if (error != null) "error": error.toString(),
                if (stackTrace != null) "stackTrace": stackTrace.toString(),
              })
            ]
          ]
        }
      ]
    };
    _sendToLokiAsync(payload);
  }

  static void _sendToLokiAsync(Map<String, dynamic> payload) async {
    try {
      final client = HttpClient();
      final request = await client.postUrl(Uri.parse(_lokiEndpoint));

      request.headers.set('Content-Type', 'application/json');
      request.headers.set('Authorization', 'Bearer $_lokiBearerToken');

      request.add(utf8.encode(jsonEncode(payload)));

      final response = await request.close();

      if (response.statusCode != 204) {
        final responseBody = await response.transform(utf8.decoder).join();
        print('[LokiLogger] HTTP ${response.statusCode}: $responseBody');
      }

      client.close();
    } catch (e) {
      print('[LokiLogger] Failed to send log to Loki: $e');
    }
  }
}
