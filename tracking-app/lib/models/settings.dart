import 'package:uuid/uuid.dart';

class Settings {
  static final String tableName = "Settings";
  static final String idColumn = "id";
  static final String lastSyncTimeColumn = "lastSyncTime";
  static final String latestRemoteChangelogIdColumn = "latestRemoteChangelogId";
  static final String latestLocalChangelogIdColumn = "latestLocalChangelogId";
  static final String clientIdColumn = "clientId";

  late String id;
  DateTime? lastSyncTime;
  int? latestRemoteChangelogId;
  int? latestLocalChangelogId;
  String? clientId;

  Settings() {
    var uuid = new Uuid();
    id = uuid.v4();
    if (clientId == null) {
      clientId = uuid.v4();
    }
  }

  void clear() {
    lastSyncTime = null;
    latestRemoteChangelogId = null;
    latestLocalChangelogId = null;
  }

  Settings.fromMap(Map<String, dynamic> map) {
    id = map[idColumn];
    int syncTimeMillis = map[lastSyncTimeColumn];
    if (syncTimeMillis > 0) {
      lastSyncTime = new DateTime.fromMillisecondsSinceEpoch(syncTimeMillis, isUtc: true);
    }
    latestRemoteChangelogId = map[latestRemoteChangelogIdColumn];
    latestLocalChangelogId = map[latestLocalChangelogIdColumn];
    clientId = map[clientIdColumn];
    if (clientId == null) {
      var uuid = new Uuid();
      clientId = uuid.v4();
    }
  }

  Map<String, dynamic> toMap() {
    var map = <String, dynamic>{
      idColumn: id,
      lastSyncTimeColumn: lastSyncTime?.millisecondsSinceEpoch ?? 0,
      latestRemoteChangelogIdColumn: latestRemoteChangelogId ?? 0,
      latestLocalChangelogIdColumn: latestLocalChangelogId ?? 0,
      clientIdColumn: clientId
    };
    return map;
  }
}
