import 'package:timeasy/dataaccess/database.dart';
import 'package:timeasy/models/settings.dart';

class SettingsRepository {
  saveSettings(Settings apiCredentials) async {
    var existingCredentials = await getSettings();
    apiCredentials.id = existingCredentials.id;
    final db = await DBProvider.dbProvider.database;
    return await db.update(Settings.tableName, apiCredentials.toMap(),
        where: "${Settings.idColumn} = ?", whereArgs: [apiCredentials.id]);
  }

  Future<Settings> getSettings() async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(Settings.tableName);
    if (queryResult.isNotEmpty) {
      return queryResult.map((entry) => Settings.fromMap(entry)).toList().first;
    }
    var apiCredentials = new Settings();
    await _addSettings(apiCredentials);
    return apiCredentials;
  }

  deleteSettings() async {
    final apiCredentials = await getSettings();
    apiCredentials.clear();
    saveSettings(apiCredentials);
  }

  _addSettings(Settings apiCredentials) async {
    final db = await DBProvider.dbProvider.database;
    return await db.insert(Settings.tableName, apiCredentials.toMap());
  }
}
