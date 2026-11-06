import 'dart:async';

import 'package:timeasy/bloc/authentication/authentication_bloc.dart';
import 'package:timeasy/bloc/authentication/authentication_state.dart';
import 'package:timeasy/bloc/synchronization/synchronization_bloc.dart';
import 'package:timeasy/bloc/synchronization/synchronization_event.dart';
import 'package:timeasy/bloc/synchronization/synchronization_state.dart';
import 'package:timeasy/services/rate_limited_exception.dart';
import 'package:timeasy/services/subscription_required_exception.dart';
import 'package:timeasy/services/synchronization_service.dart';
import 'package:timeasy/tools/retrieve_changes_result.dart';

class BackgroundSyncService {
  Timer? _timer;
  late SynchronizationService _syncService;
  late SynchronizationBloc _synchronizationBloc;
  AuthenticationBloc? _authenticationBloc;

  /// When the server rate limited us, the periodic sync waits until this
  /// point in time before trying again.
  DateTime? _skipUntil;

  BackgroundSyncService(String token, SynchronizationBloc synchronizationBloc,
      {AuthenticationBloc? authenticationBloc}) {
    _syncService = SynchronizationService(token);
    _synchronizationBloc = synchronizationBloc;
    _authenticationBloc = authenticationBloc;
  }

  void updateToken(String token) {
    _syncService.updateToken(token);
  }

  void startSync() {
    // startSync can be called again on token refresh; keep a single timer.
    _timer?.cancel();
    _skipUntil = null;
    // Sync right away so a freshly signed-in user sees their data immediately
    // instead of waiting for the first periodic run.
    synchronize();
    _timer = Timer.periodic(Duration(minutes: 1), (timer) async {
      if (_isRateLimitBackoffActive()) {
        return;
      }
      if (!isSyncing()) {
        await synchronize();
      }
    });
  }

  void stopSync() {
    _timer?.cancel();
  }

  bool isSyncing() {
    return _synchronizationBloc.state is SynchronizationInProgress;
  }

  Future<void> synchronize() async {
    if (isSyncing()) {
      return;
    }

    if (!isAuthenticated()) {
      return;
    }

    _synchronizationBloc.add(SynchonizationStartEvent());

    try {
      var retrieveResult = await _syncService.synchronize();
      _synchronizationBloc.add(SynchronizationSuccessEvent(retrieveResult));
    } on SubscriptionRequiredException {
      _synchronizationBloc.add(SynchronizationSubscriptionRequiredEvent());
    } on RateLimitedException catch (e) {
      _handleRateLimited(e);
    } catch (e) {
      _synchronizationBloc.add(SynchronizationErrorEvent(e.toString()));
    }
  }

  /// Rate limiting is transient: local changes stay in the offline outbox and
  /// the next periodic run retries. The synchronization bloc still has to
  /// leave its in-progress state, otherwise no further syncs would start.
  void _handleRateLimited(RateLimitedException e) {
    final retryAfterSeconds = e.retryAfterSeconds;
    if (retryAfterSeconds != null) {
      _skipUntil = DateTime.now().add(Duration(seconds: retryAfterSeconds));
    }
    _synchronizationBloc
        .add(SynchronizationSuccessEvent(RetrieveChangesResult(false, false)));
  }

  bool _isRateLimitBackoffActive() {
    final skipUntil = _skipUntil;
    return skipUntil != null && DateTime.now().isBefore(skipUntil);
  }

  bool isAuthenticated() {
    if (_authenticationBloc == null) {
      return false;
    }
    final state = _authenticationBloc!.state;
    final isAuth = state is AuthenticationAuthenticated;
    return isAuth;
  }
}
