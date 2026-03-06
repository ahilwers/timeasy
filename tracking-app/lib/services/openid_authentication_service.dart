import 'dart:convert';

import 'package:flutter_appauth/flutter_appauth.dart';
import 'package:http/http.dart' as http;
import 'package:timeasy/models/api_credentials.dart';
import 'package:timeasy/utils/app_logger.dart';

class OpenIdAuthenticationService {
  final List<String> scopes = ['openid', 'profile', 'offline_access'];
  final String clientId = 'timeasy-tracking-app';
  final String keycloakUri;
  final String redirectUri = 'org.timeasy.app://callback';
  final FlutterAppAuth _appAuth = FlutterAppAuth();

  OpenIdAuthenticationService(this.keycloakUri);

  String get _issuer => keycloakUri;

  Future<ApiCredentials?> authenticate() async {
    const method = 'OpenIdAuthenticationService.authenticate';
    AppLogger.i('Starting authentication flow', method: method);
    AppLogger.d('Issuer: $_issuer, RedirectUri: $redirectUri', method: method);

    try {
      AppLogger.d('Calling authorizeAndExchangeCode', method: method);
      final AuthorizationTokenResponse? result =
          await _appAuth.authorizeAndExchangeCode(
        AuthorizationTokenRequest(
          clientId,
          redirectUri,
          issuer: _issuer,
          scopes: scopes,
          preferEphemeralSession: true,
        ),
      );

      if (result == null) {
        AppLogger.w('Authentication returned null result', method: method);
        return null;
      }

      AppLogger.i('Authentication successful, processing tokens', method: method);
      AppLogger.d(
          'Access token received: ${result.accessToken != null}, '
          'Refresh token received: ${result.refreshToken != null}, '
          'ID token received: ${result.idToken != null}',
          method: method);

      var apiCredential = ApiCredentials();
      apiCredential.accessToken = result.accessToken;
      apiCredential.refreshToken = result.refreshToken;
      apiCredential.logoutUrl = _buildLogoutUrl(result.idToken);

      // Decode user info from ID token
      if (result.idToken != null) {
        final idTokenPayload = _decodeIdToken(result.idToken!);
        apiCredential.email = idTokenPayload['email'];
        apiCredential.username = idTokenPayload['preferred_username'];
        apiCredential.name = idTokenPayload['name'];
        AppLogger.d(
            'User info decoded: username=${apiCredential.username}, '
            'email=${apiCredential.email}',
            method: method);
      }

      // Store token info for refresh
      apiCredential.credentialJson = json.encode({
        'accessToken': result.accessToken,
        'refreshToken': result.refreshToken,
        'idToken': result.idToken,
        'accessTokenExpirationDateTime':
            result.accessTokenExpirationDateTime?.toIso8601String(),
      });

      AppLogger.i(
          'Authentication completed successfully for user: ${apiCredential.username}',
          method: method);
      return apiCredential;
    } catch (e, stackTrace) {
      AppLogger.e('Authentication failed',
          error: e, stackTrace: stackTrace, method: method);
      rethrow;
    }
  }

  Future<void> logout(ApiCredentials apiCredential) async {
    const method = 'OpenIdAuthenticationService.logout';
    AppLogger.i('Starting logout process', method: method);
    AppLogger.d('Logout URL: ${apiCredential.logoutUrl}', method: method);

    if (apiCredential.logoutUrl != null) {
      try {
        await _performKeycloakLogout(apiCredential);
        AppLogger.i('Logout completed successfully', method: method);
      } catch (e, stackTrace) {
        AppLogger.e('All logout attempts failed',
            error: e, stackTrace: stackTrace, method: method);
      }
    } else {
      AppLogger.w('No logout URL available', method: method);
    }
  }

  Future<void> _performKeycloakLogout(ApiCredentials apiCredential) async {
    const method = 'OpenIdAuthenticationService._performKeycloakLogout';
    final logoutUrl = apiCredential.logoutUrl!;
    AppLogger.d('Attempting logout with URL: $logoutUrl', method: method);

    // Method 1: Try GET request first
    try {
      AppLogger.d('Trying GET request to logout URL', method: method);
      final response = await http.get(Uri.parse(logoutUrl));
      AppLogger.d('GET Logout response status: ${response.statusCode}',
          method: method);

      if (response.statusCode >= 200 && response.statusCode < 400) {
        AppLogger.i('Successfully logged out from Keycloak via GET',
            method: method);
        return;
      }
    } catch (e) {
      AppLogger.d('GET logout failed: $e', method: method);
    }

    // Method 2: Try POST request with refresh token
    try {
      AppLogger.d('Trying POST request to logout endpoint', method: method);
      final uri = Uri.parse(logoutUrl);
      final logoutEndpoint =
          '${uri.scheme}://${uri.host}:${uri.port}/realms/${_extractRealm(logoutUrl)}/protocol/openid-connect/logout';

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

      AppLogger.d('POST Logout response status: ${response.statusCode}',
          method: method);

      if (response.statusCode >= 200 && response.statusCode < 400) {
        AppLogger.i('Successfully logged out from Keycloak via POST',
            method: method);
        return;
      }
    } catch (e) {
      AppLogger.d('POST logout failed: $e', method: method);
    }

    AppLogger.w('Logout completed with best effort', method: method);
  }

  String _extractRealm(String logoutUrl) {
    final uri = Uri.parse(logoutUrl);
    final pathSegments = uri.pathSegments;
    final realmIndex = pathSegments.indexOf('realms');
    if (realmIndex != -1 && realmIndex + 1 < pathSegments.length) {
      return pathSegments[realmIndex + 1];
    }
    return 'timeasy';
  }

  Future<bool> refreshToken(ApiCredentials apiCredential) async {
    const method = 'OpenIdAuthenticationService.refreshToken';
    AppLogger.d('Starting token refresh', method: method);

    if (apiCredential.refreshToken == null) {
      AppLogger.w('No refresh token available', method: method);
      return false;
    }

    try {
      final TokenResponse? result = await _appAuth.token(
        TokenRequest(
          clientId,
          redirectUri,
          issuer: _issuer,
          refreshToken: apiCredential.refreshToken,
          scopes: scopes,
        ),
      );

      if (result == null || result.accessToken == null) {
        AppLogger.w('Token refresh returned null or no access token',
            method: method);
        apiCredential.clear();
        return false;
      }

      apiCredential.accessToken = result.accessToken;
      apiCredential.refreshToken = result.refreshToken;
      apiCredential.credentialJson = json.encode({
        'accessToken': result.accessToken,
        'refreshToken': result.refreshToken,
        'idToken': result.idToken,
        'accessTokenExpirationDateTime':
            result.accessTokenExpirationDateTime?.toIso8601String(),
      });

      AppLogger.i('Token refresh successful', method: method);
      return true;
    } catch (e, stackTrace) {
      AppLogger.e('Token refresh failed',
          error: e, stackTrace: stackTrace, method: method);
      // Don't clear credentials on transient errors (network issues, timeouts).
      // The refresh token may still be valid and can be retried later.
      // Returning false without clearing allows the bloc to keep the user
      // authenticated with the existing tokens.
      return false;
    }
  }

  String? _buildLogoutUrl(String? idToken) {
    if (idToken == null) return null;
    final logoutEndpoint = '$_issuer/protocol/openid-connect/logout';
    return '$logoutEndpoint?id_token_hint=$idToken&post_logout_redirect_uri=$redirectUri';
  }

  Map<String, dynamic> _decodeIdToken(String idToken) {
    final parts = idToken.split('.');
    if (parts.length != 3) {
      return {};
    }

    final payload = parts[1];
    final normalized = base64Url.normalize(payload);
    final decoded = utf8.decode(base64Url.decode(normalized));
    return json.decode(decoded);
  }
}
