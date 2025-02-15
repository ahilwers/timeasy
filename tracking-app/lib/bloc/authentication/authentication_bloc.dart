import 'dart:async';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:timeasy/models/api_credential.dart';
import 'package:timeasy/tools/openid_authenticator.dart';

import 'authentication_event.dart';
import 'authentication_state.dart';

class AuthenticationBloc
    extends Bloc<AuthenticationEvent, AuthenticationState> {
  AuthenticationBloc() : super(AuthenticationInitial()) {
    on<LoginEvent>(_onLogin);
    on<LogoutEvent>(_onLogout);
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

  Future<ApiCredential?> _authenticate() async {
    var authenticator = _createAuthenticator();
    return await authenticator.authenticate();
  }

  Future<void> _logout() async {
    if (state is AuthenticationAuthenticated) {
      var credentials = (state as AuthenticationAuthenticated).credentials;
      var authenticator = _createAuthenticator();
      await authenticator.logout(credentials);
    } else {
      throw Exception("logout failed, access token not available");
    }
  }

  OpenIdAuthenticator _createAuthenticator() {
    return new OpenIdAuthenticator("http://localhost:8180/realms/timeasy");
  }
}
