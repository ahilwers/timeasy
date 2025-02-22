import 'dart:async';

import 'package:timeasy/services/synchronization_service.dart';

class BackgroundSyncService {
  bool _isSyncing = false;
  Timer? _timer;
  late SynchronizationService _syncService;

  BackgroundSyncService(String token) {
    _syncService = SynchronizationService(token);
  }

  void updateToken(String token) {
    _syncService.updateToken(token);
  }

  void startSync() {
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
    if (_isSyncing) return;
    _isSyncing = true;
    try {
      _syncService.synchronize();
      await Future.delayed(Duration(seconds: 70));
    } catch (e) {
    } finally {
      _isSyncing = false;
    }
  }
}
