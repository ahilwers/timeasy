import 'dart:async';

import 'package:timeasy/bloc/synchronization/synchronization_bloc.dart';
import 'package:timeasy/bloc/synchronization/synchronization_event.dart';
import 'package:timeasy/bloc/synchronization/synchronization_state.dart';
import 'package:timeasy/services/synchronization_service.dart';

class BackgroundSyncService {
  Timer? _timer;
  late SynchronizationService _syncService;
  late SynchronizationBloc _synchronizationBloc;

  BackgroundSyncService(String token, SynchronizationBloc synchronizationBloc) {
    _syncService = SynchronizationService(token);
    _synchronizationBloc = synchronizationBloc;
  }

  void updateToken(String token) {
    _syncService.updateToken(token);
  }

  void startSync() {
    synchronize();
    _timer = Timer.periodic(Duration(minutes: 1), (timer) async {
      // Only synchronize if not already in progress
      if (!isSyncing()) {
        await synchronize();
      }
    });
  }

  void stopSync() {
    _timer?.cancel();
  }

  /// Checks if synchronization is currently in progress
  bool isSyncing() {
    return _synchronizationBloc.state is SynchronizationInProgress;
  }

  Future<void> synchronize() async {
    // Check if synchronization is already in progress
    if (isSyncing()) {
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
}
