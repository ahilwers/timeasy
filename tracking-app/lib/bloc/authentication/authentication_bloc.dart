import 'dart:async';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:timeasy/models/api_credentials.dart';
import 'package:timeasy/repositories/api-credentials_repository.dart';
import 'package:timeasy/services/openid-authentication_service.dart';

import 'authentication_event.dart';
import 'authentication_state.dart';

class AuthenticationBloc
    extends Bloc<AuthenticationEvent, AuthenticationState> {
  AuthenticationBloc() : super(AuthenticationInitial()) {
    on<LoginEvent>(_onLogin);
    on<LogoutEvent>(_onLogout);
    on<RefreshTokenEvent>(_onRefreshToken);
    on<SetAuthenticationEvent>((event, emit) {
      emit(AuthenticationAuthenticated(event.credentials));
    });

    on<ClearAuthenticationEvent>((event, emit) {
      emit(AuthenticationInitial());
    });

    on<AuthenticationErrorEvent>((event, emit) {
      emit(AuthenticationError(event.message));
    });
  }

  FutureOr<void> _onLogin(
      LoginEvent event, Emitter<AuthenticationState> emit) async {
    try {
      var apiCredential = await _authenticate();
      if (apiCredential != null) {
        emit(AuthenticationAuthenticated(apiCredential));
      } else {
        emit(AuthenticationError("error during login: credential is null"));
      }
    } catch (e) {
      emit(AuthenticationError(e.toString()));
    }
  }

  FutureOr<void> _onLogout(
      LogoutEvent event, Emitter<AuthenticationState> emit) async {
    try {
      await _logout();
      emit(AuthenticationInitial());
    } catch (e) {
      emit(AuthenticationError(e.toString()));
    }
  }

  Future<ApiCredentials?> _authenticate() async {
    var authenticator = _createAuthenticator();
    var credentials = await authenticator.authenticate();
    if (credentials != null) {
      var repository = new ApiCredentialsRepository();
      await repository.saveApiCredentials(credentials);
    }
    return credentials;
  }

  Future<void> _logout() async {
    if (state is AuthenticationAuthenticated) {
      var credentials = (state as AuthenticationAuthenticated).credentials;
      var authenticator = _createAuthenticator();
      await authenticator.logout(credentials);
      var repository = new ApiCredentialsRepository();
      await repository.deleteApiCredentials();
    } else {
      throw Exception("logout failed, access token not available");
    }
  }

  OpenIdAuthenticationService _createAuthenticator() {
    return new OpenIdAuthenticationService(
        "http://localhost:8180/realms/timeasy");
  }

  FutureOr<void> _onRefreshToken(
      RefreshTokenEvent event, Emitter<AuthenticationState> emit) async {
    try {
      var repository = new ApiCredentialsRepository();
      var credentials = await repository.getApiCredentials();
      var authenticator = _createAuthenticator();
      if (await authenticator.refreshToken(credentials)) {
        emit(AuthenticationAuthenticated(credentials));
      } else {
        emit(AuthenticationInitial());
      }
    } catch (e) {
      emit(AuthenticationError(e.toString()));
      emit(AuthenticationInitial());
    }
  }
}
