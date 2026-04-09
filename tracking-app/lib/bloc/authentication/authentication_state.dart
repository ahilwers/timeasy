import 'package:timeasy/models/api_credentials.dart';

abstract class AuthenticationState {}

class AuthenticationInitial extends AuthenticationState {}

class AuthenticationAuthenticated extends AuthenticationState {
  final ApiCredentials credentials;

  AuthenticationAuthenticated(this.credentials);
}

class AuthenticationError extends AuthenticationState {
  final String message;

  AuthenticationError(this.message);
}

class AuthenticationSubscriptionExpired extends AuthenticationState {}
