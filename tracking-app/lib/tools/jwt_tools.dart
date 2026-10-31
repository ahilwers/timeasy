import 'dart:convert';

/// Decodes the payload of a JWT without verifying its signature.
///
/// Used to read identity claims from Keycloak tokens (e.g. the `sub` claim,
/// which is the Keycloak user ID and doubles as the RevenueCat App User ID).
Map<String, dynamic> decodeJwtPayload(String token) {
  final parts = token.split('.');
  if (parts.length != 3) {
    return {};
  }
  try {
    final normalized = base64Url.normalize(parts[1]);
    final decoded = utf8.decode(base64Url.decode(normalized));
    final payload = json.decode(decoded);
    return payload is Map<String, dynamic> ? payload : {};
  } catch (_) {
    return {};
  }
}
