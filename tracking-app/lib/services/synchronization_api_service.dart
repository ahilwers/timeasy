import 'package:timeasy/models/sync_data.dart';
import 'package:timeasy/services/api_service.dart';

class SynchronizationApiService extends ApiService {
  SynchronizationApiService({required String baseUrl, String? token})
      : super(baseUrl: baseUrl, token: token) {}

  Future<SyncData> getChangedData(int? sinceChangelogId, String? clientId) async {
    var sinceChangelogIdStr = sinceChangelogId?.toString() ?? '0';
    var url = '/sync/changed/${sinceChangelogIdStr}';
    if (clientId != null) {
      url += '?clientId=${clientId}';
    }

    var response = await get(url);
    if (isSuccessStatusCode(response.statusCode)) {
      return SyncData.fromJson(response.body);
    }
    throw Exception(
        'Failed to get changed data, status code: ${response.statusCode}');
  }

  Future<void> sendSyncData(SyncData syncData, String? clientId) async {
    var url = "/sync/changed";
    if (clientId != null) {
      url += '?clientId=${clientId}';
    }
    var response = await post(url, data: syncData.toJson());
    if (!isSuccessStatusCode(response.statusCode)) {
      throw Exception(
          'Failed to send changed data, status code: ${response.statusCode}');
    }
  }
}
