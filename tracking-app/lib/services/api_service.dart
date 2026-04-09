import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:timeasy/exceptions/subscription_expired_exception.dart';

class ApiService {
  final String baseUrl;
  String? _token;

  ApiService({required this.baseUrl, String? token}) {
    _token = token;
  }

  void updateToken(String token) {
    _token = token;
  }

  // New helper method to check for subscription errors
  void _checkForSubscriptionError(http.Response response) {
    if (response.statusCode == 402) {
      throw const SubscriptionExpiredException();
    }
  }

  Future<http.Response> get(String endpoint,
      {Map<String, String>? params}) async {
    final uri = Uri.parse('$baseUrl$endpoint').replace(queryParameters: params);
    final response = await http.get(uri, headers: _createHeaders());
    _checkForSubscriptionError(response);  // Check for 402 status
    return response;
  }

  Future<http.Response> post(String endpoint,
      {Map<String, dynamic>? data}) async {
    final uri = Uri.parse('$baseUrl$endpoint');
    final response = await http.post(uri,
        headers: _createHeaders(), body: jsonEncode(data));
    _checkForSubscriptionError(response);  // Check for 402 status
    return response;
  }

  bool isSuccessStatusCode(int statusCode) {
    return statusCode >= 200 && statusCode < 300;
  }

  Map<String, String> _createHeaders() {
    return {
      'Content-Type': 'application/json',
      if (_token != null) 'Authorization': 'Bearer $_token',
    };
  }
}
