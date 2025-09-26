import 'package:timeasy/utils/loki_logger.dart';

class AppLogger {
  static void d(String message,
          {dynamic error, StackTrace? stackTrace, String method = ''}) =>
      LokiLogger.d(message,
          error: error, stackTrace: stackTrace, method: method);

  static void i(String message,
          {dynamic error, StackTrace? stackTrace, String method = ''}) =>
      LokiLogger.i(message,
          error: error, stackTrace: stackTrace, method: method);

  static void w(String message,
          {dynamic error, StackTrace? stackTrace, String method = ''}) =>
      LokiLogger.w(message,
          error: error, stackTrace: stackTrace, method: method);

  static void e(String message,
          {dynamic error, StackTrace? stackTrace, String method = ''}) =>
      LokiLogger.e(message,
          error: error, stackTrace: stackTrace, method: method);
}
