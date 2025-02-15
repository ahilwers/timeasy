import 'package:timeasy/dataaccess/database.dart';
import 'package:timeasy/models/api_credentials.dart';

class ApiCredentialsRepository {
  saveApiCredentials(ApiCredentials apiCredentials) async {
    var existingCredentials = await getApiCredentials();
    apiCredentials.id = existingCredentials.id;
    final db = await DBProvider.dbProvider.database;
    return await db.update(ApiCredentials.tableName, apiCredentials.toMap(),
        where: "${ApiCredentials.idColumn} = ?",
        whereArgs: [apiCredentials.id]);
  }

  Future<ApiCredentials> getApiCredentials() async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(ApiCredentials.tableName);
    if (queryResult.isNotEmpty) {
      return queryResult
          .map((entry) => ApiCredentials.fromMap(entry))
          .toList()
          .first;
    }
    var apiCredentials = new ApiCredentials();
    await _addApiCredentials(apiCredentials);
    return apiCredentials;
  }

  deleteApiCredentials() async {
    final apiCredentials = await getApiCredentials();
    apiCredentials.clear();
    saveApiCredentials(apiCredentials);
  }

  _addApiCredentials(ApiCredentials apiCredentials) async {
    final db = await DBProvider.dbProvider.database;
    return await db.insert(ApiCredentials.tableName, apiCredentials.toMap());
  }
}
