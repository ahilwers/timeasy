import '../../models/api_credentials.dart';

abstract class AuthenticationEvent {}

class SetAuthenticationEvent extends AuthenticationEvent {
  final ApiCredentials credentials;

  SetAuthenticationEvent(this.credentials);
}

class ClearAuthenticationEvent extends AuthenticationEvent {}

class LoginEvent extends AuthenticationEvent {}

/// Opens the Keycloak registration page instead of the login page, so users
/// can create an account before subscribing (the Keycloak user ID is the
/// RevenueCat App User ID, so an account is required for in-app purchases).
class RegisterEvent extends AuthenticationEvent {}

class LogoutEvent extends AuthenticationEvent {}

class RefreshTokenEvent extends AuthenticationEvent {}

class AuthenticationErrorEvent extends AuthenticationEvent {
  final String message;

  AuthenticationErrorEvent(this.message);
}
