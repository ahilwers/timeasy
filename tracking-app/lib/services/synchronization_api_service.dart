import 'package:timeasy/models/sync_data.dart';
import 'package:timeasy/services/api_service.dart';

class SynchronizationApiService extends ApiService {
  SynchronizationApiService({required String baseUrl, String? token})
      : super(baseUrl: baseUrl, token: token) {}

  Future<SyncData> getChangedData(DateTime? changedAfter) async {
    var changedAfterStr = changedAfter != null
        ? (changedAfter.millisecondsSinceEpoch ~/ 1000).toString()
        : '0';
    var url = '/sync/changed/${changedAfterStr}';

    var response = await get(url);
    if (isSuccessStatusCode(response.statusCode)) {
      return SyncData.fromJson(response.body);
    }
    throw Exception(
        'Failed to get changed data, status code: ${response.statusCode}');
  }

  Future<void> sendSyncData(SyncData syncData) async {
    var url = "/sync/changed";
    var response = await post(url, data: syncData.toJson());
    if (!isSuccessStatusCode(response.statusCode)) {
      throw Exception(
          'Failed to send changed data, status code: ${response.statusCode}');
    }
  }
}
