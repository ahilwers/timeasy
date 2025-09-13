import 'dart:async';

import 'package:timeasy/bloc/authentication/authentication_bloc.dart';
import 'package:timeasy/bloc/authentication/authentication_state.dart';
import 'package:timeasy/bloc/synchronization/synchronization_bloc.dart';
import 'package:timeasy/bloc/synchronization/synchronization_event.dart';
import 'package:timeasy/bloc/synchronization/synchronization_state.dart';
import 'package:timeasy/services/synchronization_service.dart';

class BackgroundSyncService {
  Timer? _timer;
  late SynchronizationService _syncService;
  late SynchronizationBloc _synchronizationBloc;
  AuthenticationBloc? _authenticationBloc;

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
    synchronize();
    _timer = Timer.periodic(Duration(minutes: 1), (timer) async {
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
    } catch (e) {
      _synchronizationBloc.add(SynchronizationErrorEvent(e.toString()));
    }
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
