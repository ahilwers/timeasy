import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:timeasy/services/subscription_required_exception.dart';

class ApiService {
  final String baseUrl;
  String? _token;

  ApiService({required this.baseUrl, String? token}) {
    _token = token;
  }

  void updateToken(String token) {
    _token = token;
  }

  Future<http.Response> get(String endpoint,
      {Map<String, String>? params}) async {
    final uri = Uri.parse('$baseUrl$endpoint').replace(queryParameters: params);
    var response = await http.get(uri, headers: _createHeaders());
    _checkForSubscriptionError(response);
    return response;
  }

  Future<http.Response> post(String endpoint,
      {Map<String, dynamic>? data}) async {
    final uri = Uri.parse('$baseUrl$endpoint');
    var response = await http.post(uri,
        headers: _createHeaders(), body: jsonEncode(data));
    _checkForSubscriptionError(response);
    return response;
  }

  bool isSuccessStatusCode(int statusCode) {
    return statusCode >= 200 && statusCode < 300;
  }

  /// Throws a [SubscriptionRequiredException] when the server responds with
  /// 402 Payment Required so callers can react with the paywall flow instead
  /// of a generic error.
  void _checkForSubscriptionError(http.Response response) {
    if (response.statusCode == 402) {
      throw SubscriptionRequiredException();
    }
  }

  Map<String, String> _createHeaders() {
    return {
      'Content-Type': 'application/json',
      if (_token != null) 'Authorization': 'Bearer $_token',
    };
  }
}
