import 'package:uuid/uuid.dart';

class Settings {
  static final String tableName = "TimeEntries";
  static final String idColumn = "id";
  static final String lastSyncTimeColumn = "lastSyncTime";

  late String id;
  DateTime? lastSyncTime;

  Settings() {
    var uuid = new Uuid();
    id = uuid.v4();
  }

  void clear() {
    lastSyncTime = null;
  }

  Settings.fromMap(Map<String, dynamic> map) {
    id = map[idColumn];
    int syncTimeMillis = map[lastSyncTimeColumn];
    if (syncTimeMillis > 0) {
      lastSyncTime =
          new DateTime.fromMillisecondsSinceEpoch(syncTimeMillis, isUtc: true);
    }
  }

  Map<String, dynamic> toMap() {
    var map = <String, dynamic>{
      idColumn: id,
      lastSyncTimeColumn: lastSyncTime?.microsecondsSinceEpoch ?? 0,
    };
    return map;
  }
}
