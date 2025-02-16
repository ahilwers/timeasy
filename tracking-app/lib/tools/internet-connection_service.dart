import 'dart:async';
import 'dart:io';

import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:flutter/widgets.dart';

class InternetConnectionService with WidgetsBindingObserver {
  bool _isConnected = false;
  late StreamSubscription<List<ConnectivityResult>> _subscription;
  final Function(bool) onConnectionChanged;

  InternetConnectionService({required this.onConnectionChanged}) {
    WidgetsBinding.instance.addObserver(this);
    _checkInternetConnection();
    _listenToInternetConnectionChanges();
  }

  Future<void> _checkInternetConnection() async {
    bool hasInternet = await _hasInternetConnection();
    _isConnected = hasInternet;
    onConnectionChanged(_isConnected);
  }

  void _listenToInternetConnectionChanges() {
    _subscription = Connectivity()
        .onConnectivityChanged
        .listen((List<ConnectivityResult> results) async {
      _isConnected = await _hasInternetConnection();
      onConnectionChanged(_isConnected);
    });
  }

  Future<bool> _hasInternetConnection() async {
    try {
      final result = await InternetAddress.lookup('timeasy.org');
      return result.isNotEmpty && result[0].rawAddress.isNotEmpty;
    } on SocketException catch (_) {
      return false;
    }
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      _checkInternetConnection();
    }
  }

  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _subscription.cancel();
  }
}
