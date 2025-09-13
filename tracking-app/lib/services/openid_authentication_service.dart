import 'dart:convert';
import 'dart:io';

import 'package:http/http.dart' as http;
import 'package:openid_client/openid_client.dart';
import 'package:openid_client/openid_client_io.dart' as io;
import 'package:timeasy/models/api_credentials.dart';
import 'package:url_launcher/url_launcher.dart';

class OpenIdAuthenticationService {
  final scopes = ['profile', 'offline_access'];
  final clientId = 'timeasy-tracking-app';
  final String keycloakUri;

  OpenIdAuthenticationService(this.keycloakUri) {}

  Future<ApiCredentials?> authenticate() async {
    var client = await getClient();
    var authenticator = io.Authenticator(client,
        scopes: scopes, port: 4000, urlLancher: _urlLauncher);
    var c = await authenticator.authorize();
    _closeWebView();

    var token = await c.getTokenResponse();
    var userInformation = await c.getUserInfo();

    var apiCredential = ApiCredentials();
    apiCredential.accessToken = token.accessToken;
    apiCredential.refreshToken = token.refreshToken;
    apiCredential.logoutUrl = c.generateLogoutUrl()?.toString();
    apiCredential.email = userInformation.email;
    apiCredential.username = userInformation.preferredUsername;
    apiCredential.name = userInformation.name;
    apiCredential.credentialJson = json.encode(c.toJson());
    return apiCredential;
  }

  Future<void> logout(ApiCredentials apiCredential) async {
    print('Starting logout process...');
    print('Logout URL: ${apiCredential.logoutUrl}');
    print('Access Token: ${apiCredential.accessToken?.substring(0, 20)}...');
    
    if (apiCredential.logoutUrl != null) {
      try {
        // Try different logout approaches
        await _performKeycloakLogout(apiCredential);
      } catch (e) {
        print('All logout attempts failed: $e');
      }
    } else {
      print('No logout URL available');
    }
  }

  Future<void> _performKeycloakLogout(ApiCredentials apiCredential) async {
    final logoutUrl = apiCredential.logoutUrl!;
    print('Attempting logout with URL: $logoutUrl');
    
    // Method 1: Try GET request first
    try {
      print('Trying GET request to logout URL...');
      final response = await http.get(Uri.parse(logoutUrl));
      print('GET Logout response status: ${response.statusCode}');
      print('GET Logout response body: ${response.body}');
      
      if (response.statusCode >= 200 && response.statusCode < 400) {
        print('Successfully logged out from Keycloak via GET');
        return;
      }
    } catch (e) {
      print('GET logout failed: $e');
    }
    
    // Method 2: Try POST request with refresh token
    try {
      print('Trying POST request to logout endpoint...');
      final uri = Uri.parse(logoutUrl);
      final logoutEndpoint = '${uri.scheme}://${uri.host}:${uri.port}/realms/${_extractRealm(logoutUrl)}/protocol/openid-connect/logout';
      
      final response = await http.post(
        Uri.parse(logoutEndpoint),
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: {
          'client_id': clientId,
          'refresh_token': apiCredential.refreshToken ?? '',
        },
      );
      
      print('POST Logout response status: ${response.statusCode}');
      print('POST Logout response body: ${response.body}');
      
      if (response.statusCode >= 200 && response.statusCode < 400) {
        print('Successfully logged out from Keycloak via POST');
        return;
      }
    } catch (e) {
      print('POST logout failed: $e');
    }
    
    // Method 3: Fallback to opening URL in browser
    try {
      print('Falling back to opening logout URL in browser...');
      await _urlLauncher(logoutUrl);
      await Future.delayed(Duration(seconds: 3)); // Give time for browser logout
      print('Opened logout URL in browser');
    } catch (urlError) {
      print('Could not open logout URL: $urlError');
      throw Exception('All logout methods failed');
    }
  }
  
  String _extractRealm(String logoutUrl) {
    // Extract realm from logout URL like: http://localhost:8080/realms/timeasy/protocol/openid-connect/logout
    final uri = Uri.parse(logoutUrl);
    final pathSegments = uri.pathSegments;
    final realmIndex = pathSegments.indexOf('realms');
    if (realmIndex != -1 && realmIndex + 1 < pathSegments.length) {
      return pathSegments[realmIndex + 1];
    }
    return 'timeasy'; // Default realm
  }

  Future<bool> refreshToken(ApiCredentials apiCredential) async {
    if (apiCredential.credentialJson == null) {
      return false;
    }
    var credential =
        Credential.fromJson(json.decode(apiCredential.credentialJson!));
    var token = await credential.getTokenResponse();
    if (token.accessToken == null) {
      apiCredential.clear();
      return false;
    }
    apiCredential.accessToken = token.accessToken;
    apiCredential.refreshToken = token.refreshToken;
    apiCredential.credentialJson = json.encode(credential.toJson());
    return true;
  }

  _urlLauncher(String url) async {
    var uri = Uri.parse(url);
    if (await canLaunchUrl(uri) || Platform.isAndroid) {
      await launchUrl(uri);
    } else {
      throw 'Could not launch $url';
    }
  }

  void _closeWebView() {
    if (Platform.isAndroid || Platform.isIOS) {
      closeInAppWebView();
    }
  }

  Future<Client> getClient() async {
    var uri = Uri.parse(keycloakUri);
    var issuer = await Issuer.discover(uri);
    return Client(issuer, clientId);
  }

  Future<Credential?> getRedirectResult(Client client,
      {List<String> scopes = const []}) async {
    return null;
  }
}
