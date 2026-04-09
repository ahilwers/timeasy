import '../../models/api_credentials.dart';

abstract class AuthenticationEvent {}

class SetAuthenticationEvent extends AuthenticationEvent {
  final ApiCredentials credentials;

  SetAuthenticationEvent(this.credentials);
}

class ClearAuthenticationEvent extends AuthenticationEvent {}

class LoginEvent extends AuthenticationEvent {}

class LogoutEvent extends AuthenticationEvent {}

class RefreshTokenEvent extends AuthenticationEvent {}

class AuthenticationErrorEvent extends AuthenticationEvent {
  final String message;

  AuthenticationErrorEvent(this.message);
}

class SubscriptionExpiredEvent extends AuthenticationEvent {}
