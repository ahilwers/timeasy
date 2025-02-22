import 'package:timeasy/services/synchronization_api_service.dart';

class SynchronizationService {
  DateTime? _syncTime;
  late SynchronizationApiService _apiService;

  SynchronizationService(String token) {
    const baseUrl = "http://localhost:8080/api/v1";
    _apiService = SynchronizationApiService(baseUrl: baseUrl, token: token);
  }

  void synchronize() async {
    _syncTime = DateTime.now();
    await _sendNewestEntries();
    await _retrieveNewestEntries();
    // set lastSyncTime to syncTime and store it

    // better idea: store last updated time of retrieved entries as last sync time
  }

  void updateToken(String token) {
    _apiService.updateToken(token);
  }

  Future<void> _sendNewestEntries() async {}

  Future<void> _retrieveNewestEntries() async {
    try {
      var syncData = await _apiService.getChangedData(null);
      print(syncData);
    } catch (e) {
      print(e.toString());
    }
  }
}
