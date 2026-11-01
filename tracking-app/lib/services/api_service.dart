import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:timeasy/services/rate_limited_exception.dart';
import 'package:timeasy/services/subscription_required_exception.dart';

class ApiService {
  final String baseUrl;
  final http.Client _client;
  String? _token;

  ApiService({required this.baseUrl, String? token, http.Client? client})
      : _client = client ?? http.Client() {
    _token = token;
  }

  void updateToken(String token) {
    _token = token;
  }

  Future<http.Response> get(String endpoint,
      {Map<String, String>? params}) async {
    final uri = Uri.parse('$baseUrl$endpoint').replace(queryParameters: params);
    var response = await _client.get(uri, headers: _createHeaders());
    _checkResponseForKnownErrors(response);
    return response;
  }

  Future<http.Response> post(String endpoint,
      {Map<String, dynamic>? data}) async {
    final uri = Uri.parse('$baseUrl$endpoint');
    var response = await _client.post(uri,
        headers: _createHeaders(), body: jsonEncode(data));
    _checkResponseForKnownErrors(response);
    return response;
  }

  bool isSuccessStatusCode(int statusCode) {
    return statusCode >= 200 && statusCode < 300;
  }

  /// Throws a [SubscriptionRequiredException] on 402 Payment Required so
  /// callers can react with the paywall flow instead of a generic error, and a
  /// [RateLimitedException] on 429 Too Many Requests so callers can retry
  /// later without surfacing an error to the user.
  void _checkResponseForKnownErrors(http.Response response) {
    if (response.statusCode == 402) {
      throw SubscriptionRequiredException();
    }
    if (response.statusCode == 429) {
      throw RateLimitedException(_parseRetryAfter(response));
    }
  }

  int? _parseRetryAfter(http.Response response) {
    // Header names are compared case-insensitively because not every HTTP
    // implementation normalizes them to lowercase.
    for (final entry in response.headers.entries) {
      if (entry.key.toLowerCase() == 'retry-after') {
        return int.tryParse(entry.value.trim());
      }
    }
    return null;
  }

  Map<String, String> _createHeaders() {
    return {
      'Content-Type': 'application/json',
      if (_token != null) 'Authorization': 'Bearer $_token',
    };
  }
}
