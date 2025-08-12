import 'package:timeasy/bloc/authentication/authentication_bloc.dart';
import 'package:timeasy/bloc/authentication/authentication_state.dart';
import 'package:timeasy/bloc/internetconnection/internet_connection_bloc.dart';
import 'package:timeasy/bloc/internetconnection/internet_connection_state.dart';
import 'package:timeasy/bloc/synchronization/synchronization_bloc.dart';
import 'package:timeasy/bloc/synchronization/synchronization_event.dart';
import 'package:timeasy/bloc/synchronization/synchronization_state.dart';
import 'package:timeasy/services/synchronization_service.dart';

/// A service that handles synchronization triggered by specific events in the app.
///
/// This service checks for internet connection and authentication before
/// triggering synchronization. It's designed to be used when a manual synchronization
/// needs to be triggered.
class EventSyncService {
  static EventSyncService? _instance;

  AuthenticationBloc? _authBloc;
  InternetConnectionBloc? _internetBloc;
  SynchronizationBloc? _syncBloc;
  SynchronizationService? _syncService;
  bool _isInitialized = false;

  EventSyncService._();

  factory EventSyncService() {
    _instance ??= EventSyncService._();
    return _instance!;
  }

  void initialize({
    required AuthenticationBloc authBloc,
    required InternetConnectionBloc internetBloc,
    required SynchronizationBloc syncBloc,
    required String initialToken,
  }) {
    if (_isInitialized) return;

    _authBloc = authBloc;
    _internetBloc = internetBloc;
    _syncBloc = syncBloc;
    _syncService = SynchronizationService(initialToken);
    _isInitialized = true;
  }

  void updateToken(String token) {
    _ensureInitialized();
    _syncService!.updateToken(token);
  }

  bool isSyncing() {
    _ensureInitialized();
    return _syncBloc!.state is SynchronizationInProgress;
  }

  bool isAuthenticated() {
    _ensureInitialized();
    return _authBloc!.state is AuthenticationAuthenticated;
  }

  bool isConnected() {
    _ensureInitialized();
    return _internetBloc!.state is InternetConnectionConnected;
  }

  bool isInitialized() {
    return _isInitialized;
  }

  /// Triggers synchronization if the user is authenticated and has internet connection.
  /// It runs synchronization in the background and doesn't block the UI.
  Future<void> synchronizeOnEvent() async {
    if (!canSync()) {
      return;
    }

    _syncBloc!.add(SynchonizationStartEvent());
    try {
      var retrieveResult = await _syncService!.synchronize();
      _syncBloc!.add(SynchronizationSuccessEvent(retrieveResult));
    } catch (e) {
      _syncBloc!.add(SynchronizationErrorEvent(e.toString()));
    }
  }

  /// Sends the changed local data to the server if the user is authenticated and has internet connection.
  /// This method is designed to be called when data is changed and a manual synchronization
  /// needs to be triggered.
  /// It runs synchronization in the background and doesn't block the UI.
  Future<void> sendDataToServer() async {
    if (!canSync()) {
      return;
    }
    try {
      // Note: sendChangesToServer is not needed as synchronize() handles both send and receive
      await _syncService!.synchronize();
    } catch (e) {
      _syncBloc!.add(SynchronizationErrorEvent(e.toString()));
    }
  }

  bool canSync() {
    if (!_isInitialized) {
      return false;
    }

    if (!isAuthenticated()) {
      return false;
    }
    if (!isConnected()) {
      return false;
    }
    if (isSyncing()) {
      return false;
    }
    return true;
  }

  void _ensureInitialized() {
    if (!_isInitialized) {
      throw Exception(
          'EventSyncService must be initialized before use. Call initialize() first.');
    }
  }
}
