import '../../models/api_credential.dart';

abstract class AuthenticationState {}

class AuthenticationInitial extends AuthenticationState {}

class AuthenticationAuthenticated extends AuthenticationState {
  final ApiCredential credentials;

  AuthenticationAuthenticated(this.credentials);
}

class AuthenticationError extends AuthenticationState {
  final String message;

  AuthenticationError(this.message);
}
