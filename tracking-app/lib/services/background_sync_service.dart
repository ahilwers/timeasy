import 'dart:async';

import 'package:timeasy/bloc/synchronization/synchronization_bloc.dart';
import 'package:timeasy/bloc/synchronization/synchronization_event.dart';
import 'package:timeasy/services/synchronization_service.dart';

class BackgroundSyncService {
  bool _isSyncing = false;
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
      if (!_isSyncing) {
        await synchronize();
      }
    });
  }

  void stopSync() {
    _timer?.cancel();
  }

  Future<void> synchronize() async {
    _synchronizationBloc.add(SynchonizationStartEvent());
    if (_isSyncing) return;
    _isSyncing = true;
    try {
      await _syncService.synchronize();
      _synchronizationBloc.add(SynchronizationSuccessEvent());
    } catch (e) {
      _synchronizationBloc.add(SynchronizationErrorEvent(e.toString()));
    } finally {
      _isSyncing = false;
    }
  }
}
