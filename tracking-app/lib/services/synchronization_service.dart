import 'package:timeasy/environment.dart';
import 'package:timeasy/repositories/settings_repository.dart';
import 'package:timeasy/services/sync_data_retriever.dart';
import 'package:timeasy/services/sync_data_sender.dart';
import 'package:timeasy/services/synchronization_api_service.dart';
import 'package:timeasy/services/external_integration_service.dart';
import 'package:timeasy/tools/retrieve_changes_result.dart';

class SynchronizationService {
  late SynchronizationApiService _apiService;
  final SettingsRepository _settingsRepository = new SettingsRepository();
  final ExternalIntegrationService _externalService = ExternalIntegrationService();

  SynchronizationService(String token) {
    _apiService = SynchronizationApiService(
        baseUrl: Environment.apiBaseUrl, token: token);
    _externalService.initialize(Environment.apiBaseUrl, token);
  }

  Future<RetrieveChangesResult> synchronize() async {
    var settings = await _settingsRepository.getSettings();

    // First send local changes
    var dataSender = new SyncDataSender(_apiService);
    await dataSender.sendNewestEntries(settings.clientId);

    // Then retrieve remote changes
    var dataRetriever = new SyncDataRetriever(_apiService);
    var result = await dataRetriever.retrieveNewestEntries(
        settings.latestRemoteChangelogId, settings.clientId);

    await _updateLastSyncTimeSettings();
    
    // Try to resolve pending external references
    try {
      await _externalService.resolvePendingReferences();
      OfflineIssueResolutionService().clearPendingReferences();
    } catch (e) {
      // Log error but don't fail the sync
      print('Failed to resolve pending external references: $e');
    }
    
    return result;
  }

  Future<RetrieveChangesResult> retrieveChangesFromServer() async {
    var settings = await _settingsRepository.getSettings();
    var dataRetriever = new SyncDataRetriever(_apiService);
    var result = await dataRetriever.retrieveNewestEntries(
        settings.latestRemoteChangelogId, settings.clientId);
    await _updateLastSyncTimeSettings();
    return result;
  }

  void updateToken(String token) {
    _apiService.updateToken(token);
    _externalService.initialize(Environment.apiBaseUrl, token);
  }

  Future<void> _updateLastSyncTimeSettings() async {
    var settings = await _settingsRepository.getSettings();
    settings.lastSyncTime = DateTime.now();
    await _settingsRepository.saveSettings(settings);
  }
}
