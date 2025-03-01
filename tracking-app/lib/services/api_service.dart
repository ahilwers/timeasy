import 'dart:convert';

import 'package:http/http.dart' as http;

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
    return await http.get(uri, headers: _createHeaders());
  }

  Future<http.Response> post(String endpoint,
      {Map<String, dynamic>? data}) async {
    final uri = Uri.parse('$baseUrl$endpoint');
    return await http.post(uri,
        headers: _createHeaders(), body: jsonEncode(data));
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
