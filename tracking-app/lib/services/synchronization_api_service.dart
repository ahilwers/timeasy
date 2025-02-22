import 'package:timeasy/models/sync_data.dart';
import 'package:timeasy/services/api_service.dart';

class SynchronizationApiService extends ApiService {
  SynchronizationApiService({required String baseUrl, String? token})
      : super(baseUrl: baseUrl, token: token) {}

  Future<SyncData> getChangedData(String? timestamp) async {
    var timestampStr = timestamp != null ? timestamp : '0';
    var url = '/sync/changed/${timestampStr}';

    var response = await get(url);
    if (isSuccessStatusCode(response.statusCode)) {
      return SyncData.fromJson(response.body);
    }
    throw Exception(
        'Failed to get changed data, status code: ${response.statusCode}');
  }
}
