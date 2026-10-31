import 'dart:async';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:timeasy/dataaccess/database.dart';
import 'package:timeasy/environment.dart';
import 'package:timeasy/models/api_credentials.dart';
import 'package:timeasy/repositories/api_credentials_repository.dart';
import 'package:timeasy/repositories/settings_repository.dart';
import 'package:timeasy/services/openid_authentication_service.dart';
import 'package:timeasy/utils/app_logger.dart';

import 'authentication_event.dart';
import 'authentication_state.dart';

class AuthenticationBloc
    extends Bloc<AuthenticationEvent, AuthenticationState> {
  AuthenticationBloc() : super(AuthenticationInitial()) {
    on<LoginEvent>(_onLogin);
    on<RegisterEvent>(_onRegister);
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
    await _authenticateAndEmit(emit);
  }

  FutureOr<void> _onRegister(
      RegisterEvent event, Emitter<AuthenticationState> emit) async {
    await _authenticateAndEmit(emit, register: true);
  }

  Future<void> _authenticateAndEmit(Emitter<AuthenticationState> emit,
      {bool register = false}) async {
    try {
      final settingsRepository = SettingsRepository();
      final settings = await settingsRepository.getSettings();
      final storedUserId =
          (await ApiCredentialsRepository().getApiCredentials()).userId;
      // The owner is persisted separately from the credentials so an expired
      // session still protects the data when another account signs in.
      final previousUserId = settings.ownerUserId ?? storedUserId;
      var apiCredential = await _authenticate(register: register);
      if (apiCredential != null) {
        final newUserId = apiCredential.userId;
        if (previousUserId != null &&
            newUserId != null &&
            previousUserId != newUserId) {
          // A different account signed in on this device — never keep another
          // user's local projects and time entries.
          await DBProvider.dbProvider.clearUserData();
        }
        settings.ownerUserId = newUserId;
        await settingsRepository.saveSettings(settings);
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

  Future<ApiCredentials?> _authenticate({bool register = false}) async {
    var authenticator = _createAuthenticator();
    var credentials = await authenticator.authenticate(register: register);
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
      // Remove locally cached projects and time entries so the next account
      // on this device starts clean.
      await DBProvider.dbProvider.clearUserData();
      var settingsRepository = new SettingsRepository();
      var settings = await settingsRepository.getSettings();
      settings.ownerUserId = null;
      await settingsRepository.saveSettings(settings);
    } else {
      throw Exception("logout failed, access token not available");
    }
  }

  OpenIdAuthenticationService _createAuthenticator() {
    print("Auth-Url: ${Environment.authUrl}");
    return new OpenIdAuthenticationService(Environment.authUrl);
  }

  FutureOr<void> _onRefreshToken(
      RefreshTokenEvent event, Emitter<AuthenticationState> emit) async {
    try {
      var repository = new ApiCredentialsRepository();
      var credentials = await repository.getApiCredentials();
      var authenticator = _createAuthenticator();
      if (await authenticator.refreshToken(credentials)) {
        emit(AuthenticationAuthenticated(credentials));
        repository.saveApiCredentials(credentials);
      } else if (credentials.refreshToken != null) {
        // Refresh failed but we still have a refresh token (transient error).
        // Keep the user authenticated so we can retry later.
        if (state is! AuthenticationAuthenticated) {
          emit(AuthenticationAuthenticated(credentials));
        }
      } else {
        // No refresh token left — credentials were explicitly invalidated.
        emit(AuthenticationInitial());
        repository.saveApiCredentials(credentials);
      }
    } catch (e) {
      // Don't log out on transient errors — keep existing auth state.
      AppLogger.e('Token refresh error: $e');
    }
  }
}
