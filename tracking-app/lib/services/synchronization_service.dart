import 'package:timeasy/environment.dart';
import 'package:timeasy/repositories/settings_repository.dart';
import 'package:timeasy/services/sync_data_retriever.dart';
import 'package:timeasy/services/sync_data_sender.dart';
import 'package:timeasy/services/synchronization_api_service.dart';

class SynchronizationService {
  late SynchronizationApiService _apiService;
  final SettingsRepository _settingsRepository = new SettingsRepository();

  SynchronizationService(String token) {
    _apiService = SynchronizationApiService(
        baseUrl: Environment.apiBaseUrl, token: token);
  }

  Future<void> synchronize() async {
    var settings = await _settingsRepository.getSettings();
    
    // First send local changes
    var dataSender = new SyncDataSender(_apiService);
    await dataSender.sendNewestEntries(settings.clientId);
    
    // Then retrieve remote changes
    var dataRetriever = new SyncDataRetriever(_apiService);
    await dataRetriever.retrieveNewestEntries(settings.latestRemoteChangelogId, settings.clientId);
    
    await _updateLastSyncTimeSettings();
  }

  void updateToken(String token) {
    _apiService.updateToken(token);
  }


  Future<void> _updateLastSyncTimeSettings() async {
    var settings = await _settingsRepository.getSettings();
    settings.lastSyncTime = DateTime.now();
    await _settingsRepository.saveSettings(settings);
  }
}
