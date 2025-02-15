import '../../models/api_credential.dart';

abstract class AuthenticationEvent {}

class SetAuthenticationEvent extends AuthenticationEvent {
  final ApiCredential credentials;

  SetAuthenticationEvent(this.credentials);
}

class ClearAuthenticationEvent extends AuthenticationEvent {}

class LoginEvent extends AuthenticationEvent {}

class LogoutEvent extends AuthenticationEvent {}

class AuthenticationErrorEvent extends AuthenticationEvent {
  final String message;

  AuthenticationErrorEvent(this.message);
}
